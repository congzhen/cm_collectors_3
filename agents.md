# CM Collectors 3 项目备忘

本文档记录对当前仓库的阅读结果，供后续开发代理或维护者快速进入项目。内容基于当前目录文件静态阅读整理。

## 项目定位

CM Collectors 3 是一个多媒体文件管理系统，用于收集、管理和播放视频、图片等资源。项目主体是 Go 后端 + Vue 3 前端，同时包含桌面壳、Windows 启动器、云播/外部播放器调用器、刮削调试器和 Docker 打包方案。

## 顶层结构

- `cm_collectors_server/`：核心 Go 服务端，使用 Gin 提供 HTTP API，GORM 管理数据库，内嵌前端静态文件。
- `cm_collectors_html/`：主前端，Vue 3 + Vite + TypeScript + Element Plus。
- `cm_collectors_wails/`：Wails 桌面客户端，Go + Svelte 前端。
- `cm_collectors_electron/`：Electron 桌面壳，加载本地前端页面并接收目标服务 URL。
- `cm_collectors_windows_launcher/`：Windows 启动器，启动同目录下的 `cm_collectors_server.exe` 并传入托盘参数。
- `cm_collectors_video_caller/`：云播/协议调用器，接收 `cmcollectorsvideoplay://` 链接并调用系统或配置的播放器。
- `cm_collectors_scraper_debugger/`：刮削器调试工具，用 `test.json` 和 `scraper/` 配置执行单条或批量刮削。
- `build/`：构建产物目录。
- `build.bat`：交互式 Windows 构建脚本，可选择构建前端、服务端、Wails、启动器、视频调用器、刮削调试器等。
- `Dockerfile.minimal`：Alpine 镜像，包含服务端可执行文件和 ffmpeg。
- `Dockerfile.full`：Ubuntu 镜像，额外包含中文环境、Chrome、xvfb 和刮削配置，适合需要浏览器刮削的场景。
- `updateSoftConfig.json`：软件更新配置。

## 环境与依赖

- 后端主服务：Go `1.24.0`，toolchain `go1.24.7`。
- 刮削调试器：Go `1.24`，toolchain `go1.24.8`。
- Wails 客户端：Go `1.23`，Wails v2。
- 主前端：Node/Vite 项目，Vue `3.5`，Vite `6`，TypeScript `5.8`，Element Plus `2.10`。
- Wails 前端：Svelte `3.49`，Vite `3`。
- Electron 壳：Electron `26`，electron-builder `24`。
- README 中写明前端需要 Node.js >= 16.x、Yarn >= 1.22.x；后端需要 Go >= 1.23.x，默认数据库为 SQLite3。

## 常用命令

开发运行：

```sh
cd cm_collectors_html && yarn dev
cd cm_collectors_server && go run .
cd cm_collectors_server && go run -tags tray main.go main_tray.go -t
```

前端构建到后端内嵌目录：

```sh
yarn --cwd ./cm_collectors_html build-server
```

主前端脚本：

```sh
cd cm_collectors_html
yarn dev
yarn build
yarn build-server
yarn test:unit
yarn test:e2e
yarn lint
yarn format
```

服务端构建示例：

```sh
cd cm_collectors_server
go build -o ../build/cm_collectors_server.exe .
go build -tags tray -o ../build/cm_collectors_server.exe .
```

一站式构建可运行根目录 `build.bat`，默认倾向构建前端、Linux 服务端和 Windows 托盘 GUI 版服务端。

## 后端概览

服务入口是 `cm_collectors_server/main.go`。

启动流程：

1. `core.Init()` 读取配置、初始化 JWT 证书、缓存、日志和 GORM。
2. `models.DB_Init(core.DBS())` 注册关联表并执行数据库迁移。
3. `processors.InitCronjob()` 初始化计划任务。
4. 根据命令行参数 `--tray` 或 `-t` 决定托盘模式或普通服务器模式。
5. 创建 Gin 路由并监听 `config.yaml` 中的 `system.serverHost` 和 `system.port`。

需要注意：服务启动前会检测配置端口是否被占用，并尝试杀掉占用该端口的进程。默认端口是 `12345`。本地调试时不要在同端口运行重要进程。

核心目录职责：

- `core/`：配置、日志、缓存、GORM、JWT、UUID、关闭回调等基础设施。
- `config/`：配置结构体定义。
- `routers/`：Gin 路由注册。静态文件从 `html` embed 目录提供，API 统一在 `/api` 下。
- `controllers/`：HTTP 控制器。
- `processors/`：业务处理层，覆盖资源、演员、标签、文件库、播放、导入、刮削、计划任务等。
- `models/`：GORM 模型与迁移。
- `api/cm_scraper/`：刮削逻辑，依赖 chromedp/goquery。
- `api/serverFileManagement/`：服务端文件管理能力。
- `processorsFFmpeg/`：视频信息、缩略图、转码、m3u8、关键帧等 ffmpeg 相关处理。
- `middleware/`：登录鉴权、云播放限流等中间件。
- `response/` 和 `errorMessage/`：统一响应与错误码/错误信息。
- `scraper/`：内置刮削器 JSON 配置。
- `cert/`：JWT 公私钥。

## API 与路由要点

公共 API 包括登录、更新配置、应用数据、标签/标签分类、文件库配置、演员、资源列表、播放信息、图片/视频流、TVBox 等。

管理员 API 通过 `middleware.AdminLoginApiMiddleware()` 保护，覆盖配置保存、数据库清理/备份、资源增删改、文件库设置、演员库导入导出、标签管理、ffmpeg 缩略图、扫描导入、刮削、计划任务等。

服务端文件管理 API 也走管理员鉴权，前缀形如 `/api/sfm/...`，支持目录列表、搜索、创建、打开、保存、重命名、上传、下载、复制、移动、压缩、解压、权限修改和删除。

非 `/api/` 的未知路由会重定向到 `/`，交给 Vue Router 处理。

健康检查：

- `/health`
- `/api/health`

## 数据库与数据目录

默认配置在 `cm_collectors_server/config.yaml`：

- 默认数据库：SQLite3。
- SQLite 文件：`./db/cm_collectors.db`。
- 系统文件路径：`./db/`。
- 静态资源目录示例：`./db/performerFace`、`./db/resCoverPoster`。
- 可选 MySQL 配置也存在，`system.database` 可切换为其它数据库类型。
- 缓存默认 `freeCache`，也支持 Redis 配置。

模型迁移使用 `gormigrate`。初始化时会创建默认文件库 `Default`、默认演员库 `Default`，并建立文件库与演员库关联。

主要模型包括文件库、文件库设置、演员库、演员、资源、资源分集/剧集、资源标签、资源演员、资源导演、标签、标签分类、计划任务等。

## 前端概览

主前端目录是 `cm_collectors_html/`。

技术栈：

- Vue 3 + Vite + TypeScript。
- Element Plus + Element Plus Icons。
- Pinia 状态管理。
- Vue Router。
- Axios 请求封装。
- video.js 播放。
- vue-cropper、瀑布流、拖拽等 UI 能力。

请求封装在 `src/assets/request.ts`：

- API 前缀固定为 `/api`。
- 请求头默认 `Content-Type: application/json`。
- 从 `sessionStorage` 注入 `token` 和 `adminToken`。
- 401 时跳转 `/adminLogin`。

Vite 开发服务器把 `/api` 代理到 `http://127.0.0.1:12345`。

主要路由：

- `/`：首页/资源内容页。
- `/mobile`：移动端入口。
- `/adminLogin`：管理员登录。
- `/performer/:mainPerformerBasesId`：演员页。
- `/performer/basesList/:filesBasesId`：演员库列表。
- `/setting`：设置页。
- `/play/movies/:resourceId/:dramaSeriesId?`：视频播放。
- `/play/comic/:resourceId/:dramaSeriesId?`：漫画播放。
- `/play/atlas/:resourceId/:dramaSeriesId?`：图集播放。

前端有移动端路由守卫：移动设备访问不允许移动端的页面时，会重定向到 `/mobile`，除非带 `desktop` 查询参数。

## 配置重点

`config.yaml` 重要字段：

- `general.logoName`：显示名称。
- `general.isAdminLogin` / `adminPassword`：管理员登录开关与密码。
- `general.theme`：主题，当前配置为 `bright`。
- `general.closePlayCloud` / `playCloudMode`：云播放相关开关与模式。
- `general.tvBoxEnabled`：TVBox 功能开关。
- `general.videoRateLimit`：视频流限流配置。
- `system.serverHost`：默认 `0.0.0.0`。
- `system.port`：默认 `12345`。
- `system.Env`：Gin 模式，默认 `debug`。
- `scraper`：浏览器路径、无头模式、日志、超时等刮削配置。
- `play.playVideoFormats` / `play.playAudioFormats`：播放格式白名单。
- `serverFileManagement.rootPath`：允许文件管理器访问的根路径。
- `taryMenu`：托盘菜单应用配置，当前示例为 Chrome。

## 桌面与辅助工具

### Wails

`cm_collectors_wails/` 是一个 Wails v2 项目，前端使用 Svelte。适合构建原生桌面客户端。

### Electron

`cm_collectors_electron/` 是轻量 Electron 壳：

- 默认服务地址为 `http://127.0.0.1:12345`。
- 支持通过 `-url` 或 `-url=...` 传入服务地址。
- 加载 `frontend/index.html`。
- 无边框窗口，默认最大化。
- 提供最小化、最大化、关闭和缩放 IPC。

### Windows Launcher

`cm_collectors_windows_launcher/main.go` 从自身所在目录启动 `cm_collectors_server.exe`，参数为 `-t -o`。代码中服务端只显式识别 `-t`/`--tray`，`-o` 可能由其它文件或旧逻辑使用，修改前需要再确认。

### Video Caller

`cm_collectors_video_caller/` 用于处理 `cmcollectorsvideoplay://` 链接：

- 优先读取同目录 `config.json` 中的播放器路径。
- 配置不存在或播放器路径无效时，尝试获取系统默认播放器。
- 对传入 URL 做协议修补后调用播放器打开。
- 文件名为 `mian.go`，不是常见的 `main.go`，构建命令需要保留当前文件名。

### Scraper Debugger

`cm_collectors_scraper_debugger/` 使用 `test.json` 作为调试入口：

- 加载指定 scraper 配置。
- 可单条按 ID 刮削，也可批量扫描目录。
- 支持跳过已有 NFO、并发数、保存图片、生成 NFO。
- 依赖 Chrome/chromedp，README 也提醒使用刮削器需要安装 Chrome。

## 构建与发布

前端 `build-server` 会输出到 `cm_collectors_server/html`，后端通过 `//go:embed html` 把前端静态文件编入可执行文件。

Docker 构建前需要先生成 Linux 服务端可执行文件到 `build/cm_collectors_server`。

典型 Docker 运行方式会映射数据库和媒体目录，例如：

```sh
docker run -d --name cm_collectors_3 -p 12345:12345 -v E:\tg_test\db:/app/db -v E:\tg_test\video:/tg -e GIN_MODE=release congzhen/cm_collectors_3:latest
```

## 测试与验证

已看到的测试/检查入口：

- `cm_collectors_html`：`yarn test:unit`、`yarn test:e2e`、`yarn type-check`、`yarn lint`。
- `cm_collectors_server/api/serverFileManagement/serverFileManagement_test.go`：服务端文件管理相关 Go 测试。
- Go 子项目通常可用 `go test ./...` 做基础验证。

本次只做静态阅读和文档整理，没有运行测试。

## 维护注意事项

- 当前仓库包含多个子项目，不同目录有独立 `go.mod` 或 `package.json`，执行命令前要进入对应目录。
- 本项目的 Git commit 说明统一使用中文撰写。
- 不要把主前端和 Wails 前端混淆：主前端是 Vue，Wails 前端是 Svelte。
- 服务端启动可能杀掉占用默认端口 `12345` 的进程，运行前先确认本机端口状态。
- 前端 API 依赖 `/api` 前缀；开发时由 Vite proxy 转发，生产时由后端同源提供。
- 后端配置文件缺失时会自动创建默认 `config.yaml`；配置保存逻辑会写回该文件。
- 刮削相关功能依赖 Chrome 或可用浏览器路径，Docker full 镜像才包含 Chrome。
- `cm_collectors_server/html` 是前端构建产物目录，通常由 `yarn build-server` 生成，修改源代码时优先改 `cm_collectors_html/src`。
- 仓库内有二进制/构建产物，如 APK、syso、icon、zip、build 目录等，常规代码改动应避免无关触碰。
- 源码中部分中文注释或字符串在 PowerShell 默认读取时可能显示为乱码；README 用 UTF-8 读取正常。不要在无需求时做大规模编码转换。
