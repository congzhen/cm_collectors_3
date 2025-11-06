# CM Collectors 3

CM Collectors 3 是一个多媒体文件管理系统，用于收集、管理和播放视频与图片等。主要基于 Vue.js + Go 开发。

## 联系方式

如果有问题或意见，可以通过以下方式联系我们：

- **电子邮件**: czpk673@gmail.com
- **Telegram 群组**: [https://t.me/+rFdWuJ8si_IzNTc1](https://t.me/+rFdWuJ8si_IzNTc1)

<div align="left">
  <img src="cm_collectors_html/public/tg.png" alt="Telegram 二维码" width="200">
</div>

### 下载地址

- **Windows 版**: [cm_collectors_3_windows.zip](https://objectstorageapi.ap-southeast-1.clawcloudrun.com/vj5i0ntw-cm-collectors-3/download/last/cm_collectors_3_windows.zip)
- **Linux 版**: [cm_collectors_3_linux.zip](https://objectstorageapi.ap-southeast-1.clawcloudrun.com/vj5i0ntw-cm-collectors-3/download/last/cm_collectors_3_linux.zip)
- **Docker 镜像**:由于刮削器依赖 Chrome 浏览器和特定操作系统环境导致打包的软件体积超过 1GB，所以暂不提供预构建的 Docker 镜像。请下载 Linux 版本并使用项目中的 Dockerfile 自行构建。
- **云播插件**: [video_caller.zip](https://objectstorageapi.ap-southeast-1.clawcloudrun.com/vj5i0ntw-cm-collectors-3/download/last/video_caller.zip)
- **刮削调试器**:[scraper_debugger.zip(支持并发设置)](https://objectstorageapi.ap-southeast-1.clawcloudrun.com/vj5i0ntw-cm-collectors-3/download/last/scraper_debugger.zip)
- **刮削器配置包**:[scraper.zip(2025-10-20)](https://objectstorageapi.ap-southeast-1.clawcloudrun.com/vj5i0ntw-cm-collectors-3/download/last/scraper.zip)

### 简介视频

[![CM Collectors 3 介绍视频](https://img.youtube.com/vi/3mRCA8Va8G4/0.jpg)](https://youtu.be/3mRCA8Va8G4)

想快速了解 CM Collectors 3 的功能和使用方法？观看我们的 [YouTube 介绍视频](https://youtu.be/3mRCA8Va8G4)

### 说明

使用刮削器时，系统需要已安装 chrome 浏览器

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

# 构建windows视频调用器
cd cm_collectors_video_caller && set GOOS=windows&& set GOARCH=amd64&& go build -ldflags -H=windowsgui -tags tray -o ../build/video_caller/cm_collectors_video_caller.exe . && copy config.json ..\build\video_caller\  && copy setup_cm_video_caller.bat ..\build\video_caller\ && copy setup_cm_video_caller.sh ..\build\video_caller\ && copy uninstall_cm_video_caller.bat ..\build\video_caller\ && copy uninstall_cm_video_caller.sh ..\build\video_caller\ && cd ..

# 构建Linux视频调用器
cd cm_collectors_video_caller && set GOOS=linux&& set GOARCH=amd64&& go build -o ../build/video_caller/cm_collectors_video_caller . && copy config.json ..\build\video_caller\  && copy setup_cm_video_caller.bat ..\build\video_caller\ && copy setup_cm_video_caller.sh ..\build\video_caller\ && copy uninstall_cm_video_caller.bat ..\build\video_caller\ && copy uninstall_cm_video_caller.sh ..\build\video_caller\ && cd ..

# 构建刮削调试器
cd cm_collectors_scraper_debugger && set GOOS=windows&& set GOARCH=amd64&& go build -o ../build/scraper_debugger/cm_collectors_scraper_debugger.exe . && copy test.json ..\build\scraper_debugger\ && robocopy scraper ..\build\scraper_debugger\scraper\ && cd ..


# 构建docker *构建前需要先构建Linux可执行文件
docker build -t cm_collectors_server .
# 保存docker镜像
docker save cm_collectors_server -o ./build/cm_collectors_server_docker.tar
# 加载docker镜像
docker load -i ./build/cm_collectors_server_docker.tar
#  运行容器 *将/app/db文件夹映射到本地，防止数据丢失 *挂载数据库文件到容器内
docker run -d  --name cm_collectors_server -p 12345:12345 -v E:\tg_test\db:/app/db -v  E:\tg_test\video:/tg  -e GIN_MODE=release  cm_collectors_server
```
