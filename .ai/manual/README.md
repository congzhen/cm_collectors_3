# CM Collectors 3 使用手册源工程

最终文件为项目根目录 `CM Collectors 3 使用说明.pdf`。正文位于 `chapters/*.json`；`manual.json` 决定章节顺序。`feature-map.json` 记录稳定功能 ID、来源指纹、截图和本次生成页码，`progress.json` 记录完成范围与验证缺口。

PDF 只发布 `manual.json` 的 `chapters` 列表。`internal_chapters` 中的 `13-reference` 保留制作范围、验证边界与维护说明，供项目内部使用，不进入面向用户的正文、目录或书签。

`review-summary.json` 记录最近一次整本内容与排版检查。用户正文不显示稳定功能 ID，也不放制作验证过程；数字图号和章内位置随发布顺序自动生成。显示模式的共用切换步骤集中在本章开头，各模式页保留效果图与特点。入门顺序为概览、安装、添加、日常操作，再介绍 Docker。

## 更新已有手册

1. 运行 `python .ai/manual/scripts/check_sources.py`，审阅来源变更。不要仅凭源码指纹就认定操作流程已更新。
2. 编辑对应章节 JSON，新增功能插入 `pages` 数组并设置唯一固定 `id`；不要修改旧功能 ID。显示形式放 `03-display`，资源标签放 `04-tags`，视频工具放 `09-video`。这些是稳定 ID，不是当前显示章号；章号按 `manual.json` 的实际顺序生成。
3. 如需截图，复用下述演示环境，只操作全新演示库。补充的数据应足以填满默认列表首屏。
4. 执行 `python .ai/manual/scripts/build_pdf.py`。它输出根目录 PDF、功能映射和 `.local/layout-audit.json`。
5. 执行 `python .ai/manual/scripts/audit_manual.py` 核对覆盖、目录链接及演示备份，再执行 `python .ai/manual/scripts/review_pdf.py` 渲染所有页面和联系表，人工检查每页版式。依赖 `pdftoppm`；查看原尺寸页面确认细小文字。
6. 更新制作记录和变更说明。新的 PDF 页码自动生成，目录可点击，书签与功能 ID 保持关联。

上述 `python` 应为装有 `requirements.txt` 中库的解释器。本次使用 Codex 工作区依赖工具提供的 Python。默认字体使用 Windows 微软雅黑，构建脚本顶部可指定其它具有中文 TrueType 字形的字体；字体文件不复制进 Git。

## 演示材料与首次重建

演示目录 `.local/runtime` 包含独立配置、新建 SQLite、封面、头像、图集、漫画、视频、预览、备份及 `demo-state.json`。**不要读取或复制开发库、build 数据、原配置或现有媒体。** `.local` 只是被 Git 忽略，文件会继续保留在本机，并非自动删除的临时目录。

首次在没有演示状态的新目录中执行：

```powershell
python .ai/manual/scripts/prepare_demo.py
python -m pip install --target .ai/manual/.local/python imageio-ffmpeg==0.6.0
python .ai/manual/scripts/generate_video.py
```

FFprobe 本次从 `@ffprobe-installer/win32-x64@5.1.0` 获取：在 `.ai/manual/.local` 执行 `npm.cmd pack @ffprobe-installer/win32-x64@5.1.0`，再执行 `tar -xzf ffprobe-installer-win32-x64-5.1.0.tgz -C runtime/ffmpeg --strip-components=1 package/ffprobe.exe`。所有下载工具留在忽略目录，不随仓库分发。也可提供自行安装的 FFmpeg / FFprobe；不要复制既有业务构建目录中的媒体或数据库。

编译独立 API 主机：在 `cm_collectors_server` 工作目录执行：

```powershell
go build -o ../.ai/manual/.local/manual-server.exe ../.ai/manual/demo/server/main.go
```

在 `.ai/manual/.local/runtime` 工作目录运行 `../manual-server.exe`。此主机复用应用真实业务处理器与数据库迁移，仅嵌入手册空白 HTML。它不包含生产前端构建和现有数据库，不会抢占或杀掉端口进程，并停止演示计划任务的自动调度。

另开终端，在项目根目录运行隔离前端：

```powershell
node ./cm_collectors_html/node_modules/vite/bin/vite.js --config .ai/manual/scripts/vite.config.mjs
```

需要主前端已有开发依赖；配置将 API 指向 18345，Vite 缓存也写入 `.local`。浏览器截图使用新建无头会话，不读取日常浏览器资料。

仅在全新的演示环境首次执行：

```powershell
node .ai/manual/scripts/seed_demo.mjs
node .ai/manual/scripts/enrich_demo.mjs
node .ai/manual/scripts/enrich_people.mjs
```

`seed_demo.mjs` 发现完整 `demo-state.json` 会复用而不重复创建。若曾中途失败且已有演示库但状态不完整，会拒绝继续，需检查该**演示目录**后决定恢复方案，不要指向任何现有业务数据。`repair_demo_tags.mjs` 是初次制作的修复工具，会重置演示资源标签；正常更新不要运行它。

## 截图与操作验证

在两个独立服务运行时，根目录执行：

```powershell
node .ai/manual/scripts/capture.mjs layouts
node .ai/manual/scripts/capture.mjs tags
node .ai/manual/scripts/capture.mjs settings
node .ai/manual/scripts/capture.mjs resources
node .ai/manual/scripts/capture_details.mjs play
node .ai/manual/scripts/capture_details.mjs advanced
node .ai/manual/scripts/capture_details.mjs library
node .ai/manual/scripts/capture_details.mjs filters
node .ai/manual/scripts/capture_details.mjs files
node .ai/manual/scripts/capture_people.mjs
```

截图通过真实页面与真实本地 API 产生。`tags` 会在演示资源上真实执行单个和批量标签操作；`layouts` 会切换演示库显示形式；`filters` 会保存搜索收藏。不要在实际业务服务地址下执行。

`layouts` 最后会自动运行 `capture_wide_layouts.mjs`，也可单独执行 `node .ai/manual/scripts/capture_wide_layouts.mjs` 重拍宽幕卡廊、封面海报盒子-宽图、封面海报精简-扩展。它复用摄影缓存的横版原图，设置 300×169 封面显示区和 cover 填充，通过真实 API/界面拍摄；完成后恢复并校验原封面字节及本库配置。备份和结果保存在 `.local/runtime/wide-layouts-*.json`；若上次中断，先执行同一命令加 `--restore` 恢复。不要用通用竖版截图覆盖这三个示例。

`node .ai/manual/scripts/verify_workflows.mjs` 会运行演示视频信息采集、指纹计算、预览检查和一项保存新文件的转码，并写入 `.local/runtime/verification.json`。此命令会改变演示操作状态，通常不必为仅修改文字而重复执行。

`node .ai/manual/scripts/capture_review_details.mjs` 补拍资源路径检索和人物批量管理。路径检索以演示库真实路径片段查询，不执行替换；人物只选择三张卡片，不提交标签、迁移或删除。配图分别为 `library-paths.jpg` 和 `performer-batch.jpg`。

## 素材与体积

- 当前截图使用 24 张 Unsplash 风景 / 城市摄影和 24 张 Pexels 人像摄影，作品及人物资料为虚构示例，不表示真实身份、作品或代言。早期程序插画仅保留为重建基础，当前展示已换为照片。
- `photo-sources.json` 保存每张照片的作者、原图页、下载地址、授权链接与 SHA-256；照片缓存位于 `.local/photography`，不进入 Git。
- 仅可重建脚本、正文、来源清单和维护记录进入 Git。全部截图（`assets/screenshots/`）、摄影原图、数据库、媒体、工具、缓存、中间页图和根目录 PDF 都保留本机并忽略 Git。
- 重新克隆仓库不会自带图片，不能立即完整构建 PDF。换机器时需另行复制本机的 `assets/screenshots/` 与 `.local/`，或按本文流程生成独立素材并重新截图；不要用开发或构建数据补齐。
- `.local/runtime/asset-manifest.json` 保存生成参数与作品名，`asset-sources.json` 保存工具来源；已下载照片会保存在本地，日常更新截图无需反复下载。
- 截图的原始分辨率为 1440×1000，手机为 430×932；裁切由 PDF 构建时处理，不覆盖原图。
- 人物页面与瀑布流的尺寸 / 列数会保存在浏览器侧。采集人物页面和瀑布流时，最后运行 `capture_people.mjs`，以保留填满首屏的演示尺寸；若重新采集其它截图，检查首屏密度。
- 不运行主前端 `build-server`，不启动现有 12345 服务，不修改现有 `config.yaml`。

## 摄影素材版更新

仅使用手册隔离目录，依次运行：

```powershell
python .ai/manual/scripts/download_photos.py
python .ai/manual/scripts/apply_photo_media.py
node .ai/manual/scripts/apply_photo_database.mjs
```

前两步缓存原图及竖版封面，并替换演示视频、图集和阅读图页；第三步需启动独立演示 API，更新 96 条资源的封面、24 位虚构人物头像，并通过真实队列重建预览。原资源 ID、标签和收藏保持关联。演示视频中保留 8 对真实文件副本用于重复检测。

素材替换后运行 `node .ai/manual/scripts/refresh_photo_tasks.mjs`，让独立演示库清除旧指纹并重新建立转码草稿；再运行操作验证与截图命令，最后重建 PDF。只在确实替换媒体后执行该刷新，不必为改文字而重跑。`prepare_demo.py` 检测到摄影版标记会保留现有摄影材料，避免后续更新又退回早期插画。照片下载和素材制作不读取任何既有开发或构建数据。

## 实际验证的边界

Windows 本地演示主机和 Vue 源界面已操作。远程刮削、外部 AI、TVBox 实机、GPU、Wails 原生窗口、Linux / Docker 实机部署未做端到端验证；手册已明确标记。源代码阅读不等同于这些链路已经跑通。
