# 测试指南

本文档记录 CM Collectors 3 已知测试入口、建议验证方式和文档整理时的验收清单。

## 已知测试入口

- `cm_collectors_html`：`yarn test:unit`、`yarn test:e2e`、`yarn type-check`、`yarn lint`。
- `cm_collectors_server/api/serverFileManagement/serverFileManagement_test.go`：服务端文件管理相关 Go 测试。
- Go 子项目通常可用 `go test ./...` 做基础验证。

## 验证建议

- 前端改动优先运行对应前端项目的类型检查、单元测试或构建。
- 后端改动优先在对应 Go 子项目运行聚焦测试，再按风险决定是否扩大到 `go test ./...`。
- 构建脚本、发布脚本和 Docker 相关改动应结合目标平台做实际构建或至少做命令路径检查。
- 刮削、视频调用、桌面壳等依赖外部程序或平台能力的改动，应明确说明本机是否具备 Chrome、播放器、Wails、Electron 等验证条件。

## 文档整理验收清单

- 根目录 `agents.md` 只保留长期稳定的协作规则和文档索引。
- `.ai/README.md` 说明 `.ai` 目录用途。
- `.ai/docs/INDEX.md` 维护文档总索引。
- `.ai/docs/architecture/` 保存项目地图、模块职责、系统边界。
- `.ai/docs/development/` 保存开发命令、构建、测试、提交约定。
- `.ai/docs/design/` 保存功能设计和技术方案。
- `.ai/docs/testing/` 保存测试计划和验收清单。
