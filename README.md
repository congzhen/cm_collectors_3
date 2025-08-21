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

#### 同时运行前后端服务

````sh
# 启动前端开发服务器
yarn --cwd ./cm_collectors_html dev

# 在另一个终端中启动后端服务
go run -C cm_collectors_server main.go

### build

```sh
# 构建前端静态文件到后端html目录
yarn --cwd ./cm_collectors_html build-server

# 构建Windows可执行文件
set GOOS=windows&& set GOARCH=amd64&& go build -C ./cm_collectors_server -o ../build/start.exe . && copy .\cm_collectors_server\config.yaml .\build\ && robocopy .\cm_collectors_server\ffmpeg .\build\ffmpeg /E

# 构建Linux可执行文件
set GOOS=linux&& set GOARCH=amd64&& go build -C ./cm_collectors_server -o ../build/start . && cp .\cm_collectors_server\config.yaml .\build\
````
