# CM Collectors 3

CM Collectors 是一个多媒体文件管理系统，主要用于收集、管理和播放视频文件。该系统由前端和后端两个部分组成，前端基于 Vue.js 构建，后端采用 Go 语言开发。

## 项目结构

## 环境要求

### 前端环境要求

- Node.js >= 16.x
- Yarn >= 1.22.x

### 后端环境要求

- Go >= 1.23.x
- SQLite3 (默认数据库)

## 快速开始

### 开发环境运行

### 同时运行前后端服务

```sh
# 启动前端开发服务器
cd cm_collectors_html && yarn dev

# 在另一个终端中启动后端服务
cd cm_collectors_server && go run .

# 启动带系统托盘的Windows版
cd cm_collectors_server && go run -tags tray main.go main_tray.go -t
```

### build 构建

```sh
# 构建前端静态文件到后端html目录
yarn --cwd ./cm_collectors_html build-server

# 构建Windows可执行文件（带系统托盘，无控制台窗口）
cd cm_collectors_server && set GOOS=windows&& set GOARCH=amd64&& go build -ldflags -H=windowsgui -tags tray -o ../build/cm_collectors_server.exe . && copy config.yaml ..\build\ && robocopy .\ffmpeg ..\build\ffmpeg /E && cd ..

# 构建Windows可执行文件（带系统托盘）
cd cm_collectors_server && set GOOS=windows&& set GOARCH=amd64&& go build -tags tray -o ../build/cm_collectors_server.exe . && copy config.yaml ..\build\ && robocopy .\ffmpeg ..\build\ffmpeg /E && cd ..

# 构建Windows可执行文件（不带系统托盘）
cd cm_collectors_server && set GOOS=windows&& set GOARCH=amd64&& go build -o ../build/cm_collectors_server.exe . && copy config.yaml ..\build\ && robocopy .\ffmpeg ..\build\ffmpeg /E && cd ..

# 构建Linux可执行文件（不带系统托盘）
cd cm_collectors_server && set GOOS=linux&& set GOARCH=amd64&& go build -o ../build/cm_collectors_server . && copy config.yaml ..\build\ && cd ..

# 构建wails
cd cm_collectors_wails && wails build -o ../../../build/cm_collectors_wails.exe && cd ..

# 构建windows启动器
cd cm_collectors_windows_launcher && go build -ldflags -H=windowsgui -o ../build/CM_launcher.exe . && cd ..

# 构建视频调用器
cd cm_collectors_video_caller && set GOOS=windows&& set GOARCH=amd64&& go build -ldflags -H=windowsgui -tags tray -o ../build/video_caller/cm_collectors_video_caller.exe . && copy config.json ..\build\video_caller\  && copy setup_cm_video_caller.bat ..\build\video_caller\ && copy 请使用管理员身份运行setup文件 ..\build\video_caller\ && cd ..


# 构建docker *构建前需要线构建Linux可执行文件
docker build -t cm_collectors_server .
# 保存docker镜像
docker save cm_collectors_server -o ./build/cm_collectors_server_docker.tar
# 加载docker镜像
docker load -i ./build/cm_collectors_server_docker.tar
#  运行容器
docker run -d  --name cm_collectors_server -p 12345:12345 -v E:\tg_test\db:/app/db -v  E:\tg_test\video:/tg  -e GIN_MODE=release  cm_collectors_server
```
