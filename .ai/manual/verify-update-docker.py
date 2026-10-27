"""Run real Docker update checks using dedicated containers and disposable volumes.

Uses local prepared ZIPs to test installation without publishing a GitHub release.
Never mounts existing application data. Logs and results stay in .local/.
"""
import hashlib
import json
import os
from pathlib import Path
import shutil
import sqlite3
import subprocess
import time
import urllib.request
import uuid
import zipfile

ROOT = Path(__file__).resolve().parents[2]
RUN = "cm-update-check-" + uuid.uuid4().hex[:8]
OUT = ROOT / ".ai/manual/.local" / RUN
OUT.mkdir(parents=True)
DOCKER = ["docker", "--context", "desktop-linux"]
RESULTS = []
containers, volumes, images = [], [], []


def run(args, **kwargs):
    return subprocess.run([str(a) for a in args], check=True, **kwargs)


def docker(*args, **kwargs):
    return run([*DOCKER, *args], **kwargs)


def output(*args):
    return docker(*args, capture_output=True, text=True).stdout.strip()


def passed(message):
    RESULTS.append(message)
    print("PASS: " + message, flush=True)


def wait_for(check, timeout=120):
    deadline = time.monotonic() + timeout
    last = None
    while time.monotonic() < deadline:
        try:
            value = check()
            if value:
                return value
        except Exception as error:
            last = error
        time.sleep(1)
    raise RuntimeError(f"Timed out: {last}")


def build_go(name, version=None, tests=False):
    env = dict(os.environ, GOOS="linux", GOARCH="amd64", CGO_ENABLED="0")
    args = ["go", "test", "-c", "-o", str(OUT / name), "./internal/selfupdate"] if tests else [
        "go", "build", "-ldflags", f"-X cm_collectors_server/internal/selfupdate.Version={version}",
        "-o", str(OUT / name), ".",
    ]
    run(args, env=env, cwd=ROOT / "cm_collectors_server")


def request(base, path, body=None):
    data = None if body is None else json.dumps(body).encode()
    req = urllib.request.Request(base + path, data=data, headers={"Content-Type": "application/json"})
    with urllib.request.urlopen(req, timeout=5) as response:
        return json.load(response)


def test_image(flavor):
    image, volume, name = f"{RUN}:{flavor}", f"{RUN}-{flavor}", f"{RUN}-{flavor}"
    context = OUT / (flavor + "-context")
    (context / "build").mkdir(parents=True)
    shutil.copyfile(OUT / "old-server", context / "build/cm_collectors_server")
    shutil.copyfile(ROOT / ("Dockerfile." + flavor), context / "Dockerfile")
    if flavor == "full":
        shutil.copytree(ROOT / "cm_collectors_server/scraper", context / "cm_collectors_server/scraper")
    images.append(image)
    with (OUT / (flavor + "-build.log")).open("w", encoding="utf-8") as log:
        docker("build", "-t", image, str(context), stdout=log, stderr=subprocess.STDOUT)
    passed(flavor + ": production Dockerfile builds")
    docker("volume", "create", volume, capture_output=True)
    volumes.append(volume)
    containers.append(name)
    docker("run", "-d", "--name", name, "--network", "none", "-v", volume + ":/app",
           "--entrypoint", "/bin/sh", image, "-c", "sleep 600", capture_output=True)
    seed = OUT / (flavor + "-seed")
    job = seed / ".cm-updates/job-local"
    job.mkdir(parents=True)
    (seed / "db").mkdir()
    config = b'general:\n  logoName: update-regression-marker\n  isAdminLogin: false\nsystem:\n  serverHost: 0.0.0.0\n  port: 12345\n  database: sqlite3\n  filePath: ./db/\n  Env: release\n'
    (seed / "config.yaml").write_bytes(config)
    (seed / "db/preserve.txt").write_text("original user data", encoding="utf-8")
    with sqlite3.connect(seed / "db/cm_collectors.db") as db:
        db.execute("create table regression_marker (value text)")
        db.execute("insert into regression_marker values ('original database')")
    package = job / "package.zip"
    with zipfile.ZipFile(package, "w", zipfile.ZIP_DEFLATED) as archive:
        archive.write(OUT / "new-server", "cm_collectors_server")
        archive.writestr("config.yaml", "must not replace user config")
        archive.writestr("db/preserve.txt", "must not replace user data")
        archive.writestr("update-marker.txt", "new package installed")
    asset = {"size": package.stat().st_size, "sha256": hashlib.sha256(package.read_bytes()).hexdigest()}
    state = {"phase": "ready", "version": "3.3.1", "job": "/app/.cm-updates/job-local", "asset": asset}
    (seed / ".cm-updates/state.json").write_text(json.dumps(state), encoding="utf-8")
    docker("cp", str(seed) + "/.", name + ":/app")
    docker("rm", "-f", name, capture_output=True)

    def launch():
        docker("run", "-d", "--name", name, "-p", "127.0.0.1::12345", "-v", volume + ":/app", image,
               capture_output=True)
        port = output("port", name, "12345/tcp").rsplit(":", 1)[1]
        return "http://127.0.0.1:" + port

    base = launch()
    health = wait_for(lambda: request(base, "/api/health"))
    assert health["version"] == "3.3.0", health
    assert request(base, "/api/app/data")["status"] is True
    passed(flavor + ": original application and business API start")
    assert request(base, "/api/update/status")["data"]["phase"] == "ready"
    installed = request(base, "/api/update/install", {"version": "3.3.1", "job": state["job"]})
    assert installed["status"] is True, installed

    def completed():
        value = request(base, "/api/update/status")["data"]
        if value["phase"] in ("ready", "failed") and value.get("error"):
            raise RuntimeError(value["error"])
        return value if value["phase"] == "installed" else None

    final = wait_for(completed)
    assert request(base, "/api/health")["version"] == "3.3.1"
    assert request(base, "/api/app/data")["status"] is True
    assert output("exec", name, "cat", "/app/update-marker.txt") == "new package installed"
    assert output("exec", name, "cat", "/app/db/preserve.txt") == "original user data"
    assert output("exec", name, "cat", "/app/config.yaml").encode() == config.rstrip()
    backup = OUT / (flavor + "-backup")
    backup.mkdir()
    docker("cp", name + ":" + final["backup"] + "/.", str(backup))
    db_backups = list(backup.glob("database-*.sqlite"))
    assert db_backups, "SQLite backup missing"
    with sqlite3.connect(db_backups[0]) as db:
        assert db.execute("select value from regression_marker").fetchone()[0] == "original database"
    assert (backup / "config.yaml").read_bytes() == config
    passed(flavor + ": real install, backup, restart, version check and protected data")
    docker("stop", "-t", "15", name, capture_output=True)
    assert output("inspect", "-f", "{{.State.ExitCode}}", name) == "0"
    docker("rm", name, capture_output=True)
    base = launch()
    health = wait_for(lambda: request(base, "/api/health"))
    assert health["version"] == "3.3.1", health
    assert request(base, "/api/app/data")["status"] is True
    assert output("exec", name, "cat", "/app/db/preserve.txt") == "original user data"
    passed(flavor + ": clean stop and recreated container retains upgraded application")


try:
    output("version", "--format", "{{.Server.Os}}/{{.Server.Arch}}")
    build_go("old-server", "3.3.0")
    build_go("new-server", "3.3.1")
    build_go("selfupdate.test", tests=True)
    with (OUT / "linux-unit.log").open("w", encoding="utf-8") as log:
        docker("run", "--rm", "--network", "none", "-v", str(OUT) + ":/tests:ro", "alpine:3.18",
               "/tests/selfupdate.test", "-test.v", stdout=log, stderr=subprocess.STDOUT)
    passed("Linux updater tests, including executable permissions and real subprocess restart")
    for flavor in ("minimal", "full"):
        test_image(flavor)
finally:
    for name in containers:
        log = subprocess.run([*DOCKER, "logs", name], capture_output=True, text=True, errors="replace")
        (OUT / (name + ".log")).write_text(log.stdout + log.stderr, encoding="utf-8")
        subprocess.run([*DOCKER, "rm", "-f", name], capture_output=True)
    for volume in volumes:
        subprocess.run([*DOCKER, "volume", "rm", volume], capture_output=True)
    for image in images:
        subprocess.run([*DOCKER, "image", "rm", image], capture_output=True)
    (OUT / "results.json").write_text(json.dumps(RESULTS, ensure_ascii=False, indent=2), encoding="utf-8")
    print("Logs: " + str(OUT), flush=True)
