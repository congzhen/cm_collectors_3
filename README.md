# cm_collectors_3

### Compile and Hot-Reload for Development

```sh
yarn --cwd ./cm_collectors_html dev
```

### build

```sh
yarn --cwd ./cm_collectors_html build-server

set GOOS=windows&& set GOARCH=amd64&& go build -C ./cm_collectors_server -o ../build/start.exe . && copy .\cm_collectors_server\config.yaml .\build\ && robocopy .\cm_collectors_server\ffmpeg .\build\ffmpeg /E
```
