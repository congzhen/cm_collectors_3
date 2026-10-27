"""Compile update-capable Windows/Linux binaries without touching release artifacts."""
import json
import os
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[2]
output = root / ".ai/manual/.local/update-check"
temp = output / "tmp"
temp.mkdir(parents=True, exist_ok=True)
version = json.loads((root / "updateSoftConfig.json").read_text(encoding="utf-8"))["version"]
for platform, name, flags in [
    ("windows", "server.exe", []),
    ("windows", "server-tray.exe", ["-tags", "tray"]),
    ("linux", "server-linux", []),
]:
    env = dict(os.environ, GOOS=platform, GOARCH="amd64", CGO_ENABLED="0",
               GOTMPDIR=str(temp), TEMP=str(temp), TMP=str(temp))
    linker = f"-X cm_collectors_server/internal/selfupdate.Version={version}"
    if flags:
        linker += " -H=windowsgui"
    subprocess.run(["go", "build", *flags, "-ldflags", linker, "-o", str(output / name), "."],
                   cwd=root / "cm_collectors_server", env=env, check=True)
    print(f"Built {platform}/amd64: {name}", flush=True)
