# CM Collectors 3 AI 协作入口

本文件是 Codex/AI 进入仓库时优先阅读的入口，只保留长期稳定的协作规则和文档索引。项目资料、开发命令、测试清单和设计方案统一维护在 `.ai/docs/`。

## 协作规则

- 不要把主前端和 Wails 前端混淆：主前端是 `cm_collectors_html/`，使用 Vue 3；Wails 前端是 `cm_collectors_wails/`，使用 Svelte。
- 修改前先确认所属子项目。当前仓库包含多个独立 `go.mod` 或 `package.json`，执行命令前应进入对应目录，或通过工具的 `workdir` 指定目录。
- 不要改动无关业务代码、构建产物或二进制文件。常规代码改动应避免触碰 APK、syso、icon、zip、`build/` 等产物。
- `cm_collectors_server/html` 是前端构建产物目录，通常由 `yarn build-server` 生成；修改主前端时优先改 `cm_collectors_html/src`。
- 服务端启动可能杀掉占用默认端口 `12345` 的进程，运行前先确认本机端口状态。
- 前端 API 依赖 `/api` 前缀；开发时由 Vite proxy 转发，生产时由后端同源提供。
- 刮削相关功能依赖 Chrome 或可用浏览器路径，Docker full 镜像才包含 Chrome。
- 源码中部分中文注释或字符串在 PowerShell 默认读取时可能显示为乱码；不要在无需求时做大规模编码转换。

## 提交规则

- Git commit 说明统一使用中文撰写，并使用 Conventional Commits 带 scope 的格式：`type(scope): 中文说明`。
- 修复类提交必须写成 `fix(xxx): ...`，不要写成无 scope 的 `fix: ...`。
- 除非用户明确要求极简提交，否则提交信息需要包含正文，说明改动内容、关键实现原因、用户可感知行为变化和验证结果。

## 命令习惯

- 需要重复执行的验证命令应保持命令字符串稳定，并通过工具的 `workdir` 指定目录。
- 不要把 `cd ...;`、环境变量设置、额外输出或其它组合语句混进验证命令字符串。
- 如果用户已经对某个命令选择“以后都可以执行”，后续应尽量复用完全相同的命令形式，例如 `yarn.cmd run check`、`yarn.cmd type-check`、`yarn.cmd build`。

## 文档索引

- `.ai/README.md`：AI 协作工作区说明。
- `.ai/docs/INDEX.md`：项目文档总索引。
- `.ai/docs/architecture/project-map.md`：项目地图、模块职责、系统边界。
- `.ai/docs/development/development-guide.md`：环境依赖、开发运行、构建发布、提交约定。
- `.ai/docs/testing/testing-guide.md`：测试入口、验证建议、验收清单。
- `.ai/docs/design/README.md`：功能设计和技术方案目录。
