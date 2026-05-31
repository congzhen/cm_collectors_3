# 开发指南

本文档记录 CM Collectors 3 的环境依赖、开发命令、构建发布和维护约定。

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

## 构建与发布

前端 `build-server` 会输出到 `cm_collectors_server/html`，后端通过 `//go:embed html` 把前端静态文件编入可执行文件。

Docker 构建前需要先生成 Linux 服务端可执行文件到 `build/cm_collectors_server`。

典型 Docker 运行方式会映射数据库和媒体目录，例如：

```sh
docker run -d --name cm_collectors_3 -p 12345:12345 -v E:\tg_test\db:/app/db -v E:\tg_test\video:/tg -e GIN_MODE=release congzhen/cm_collectors_3:latest
```

## 维护注意事项

- 当前仓库包含多个子项目，不同目录有独立 `go.mod` 或 `package.json`，执行命令前要进入对应目录。
- 本项目的 Git commit 说明统一使用中文撰写，并使用 Conventional Commits 带 scope 的格式：`type(scope): 中文说明`，例如 `fix(update): 切换更新检测到GitHub`。如果是修复类提交，必须写成 `fix(xxx): ...`，不要写成无 scope 的 `fix: ...`。
- Git commit 不要只写一行标题；除非用户明确要求极简提交，否则提交信息需要包含详尽正文，说明本次改动内容、关键实现原因、用户可感知行为变化和验证结果。
- 不要把主前端和 Wails 前端混淆：主前端是 Vue，Wails 前端是 Svelte。
- 服务端启动可能杀掉占用默认端口 `12345` 的进程，运行前先确认本机端口状态。
- 前端 API 依赖 `/api` 前缀；开发时由 Vite proxy 转发，生产时由后端同源提供。
- 后端配置文件缺失时会自动创建默认 `config.yaml`；配置保存逻辑会写回该文件。
- 刮削相关功能依赖 Chrome 或可用浏览器路径，Docker full 镜像才包含 Chrome。
- `cm_collectors_server/html` 是前端构建产物目录，通常由 `yarn build-server` 生成，修改源代码时优先改 `cm_collectors_html/src`。
- 仓库内有二进制/构建产物，如 APK、syso、icon、zip、build 目录等，常规代码改动应避免无关触碰。
- 源码中部分中文注释或字符串在 PowerShell 默认读取时可能显示为乱码；README 用 UTF-8 读取正常。不要在无需求时做大规模编码转换。

## 命令权限与验证命令习惯

- 需要重复执行的验证命令应保持命令字符串稳定，并通过工具的 `workdir` 指定目录，不要把 `cd ...;`、环境变量设置、额外输出或其它组合语句混进命令字符串。
- 如果用户已经对某个命令选择“以后都可以执行”，后续应尽量复用完全相同的命令形式，例如 `yarn.cmd run check`、`yarn.cmd type-check`、`yarn.cmd build`，避免在无必要时改成等价但不同的写法。
- 这类记录只能帮助代理减少因命令前缀不同造成的重复授权询问，不能替代沙箱/权限系统本身的审批规则。
