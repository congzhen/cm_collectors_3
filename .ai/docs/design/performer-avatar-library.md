# Gfriends 演员头像库设计

## 背景

CM Collectors 3 已有演员资料、头像本地存储和演员资料刮削能力。Gfriends 提供专门的演员头像文件树，但不提供完整演员资料，因此本功能作为独立头像库实现，不加入通用刮削器。

## 目标

- 用户手动下载或更新 Gfriends `Filetree.json`。
- 数据文件不存在时，从单人或批量入口弹出阻断式确认；用户确认后自动下载并继续原操作。
- 按演员姓名和别名查询候选头像。
- 单个演员支持推荐、原图优先、AI 优化图优先和手动选择。
- 批量处理支持全部演员或搜索、分页勾选指定演员；主姓名始终排在别名前，分页大小与演员列表一样由响应式 `pageSize` 变量控制。批量只提供三种自动策略，默认仅补充无头像演员，并且整批只确认一次。
- 头像下载后继续保存到现有 `performerFace` 目录，只更新 `performer.photo`。

## 非目标

- 不从 Gfriends 更新演员生日、三围、国籍或简介。
- 不加入现有网页刮削器配置；计划任务只提供头像缓存清理，不自动匹配演员头像。
- 第一版不提供来源黑名单、自动更新和按分辨率评分。
- 不把头像或文件树打包到项目构建产物。

## 数据目录

可重新下载的运行时文件位于 `performerAvatarLibrary.cachePath`，默认不在数据库备份目录内：

```text
runtime/cache/gfriends/Filetree.json
runtime/cache/gfriends/metadata.json
runtime/cache/gfriends/images/*.cache
```

自定义镜像和默认头像策略统一写入 `config.yaml` 的 `performerAvatarLibrary` 配置段。该配置作为
`core.Config` 的一部分读取和保存，并由 `core.getDefaultConfig()` 补齐默认值；前端通过现有应用配置
接口读写，不维护头像库专用配置文件或旁路保存接口。

更新使用临时文件，完成 JSON 结构和可检索内容校验后才替换旧文件。更新失败时保留旧版本。

头像缓存统计只包含 `images` 目录中的 `.cache` 文件。清理操作同时移除本功能遗留的
`.cache.tmp`、`.cache.bak` 文件，但保留
`Filetree.json`、`metadata.json`、已保存到演员资料目录的正式头像，以及缓存目录中的未知文件。
手动清理、软件启动自动清理和计划任务均调用 `PerformerAvatarLibrary.ClearImageCache`。

默认数据源为 GitHub Raw；下载失败时尝试 Gfriends 文档提供的 jsDelivr 镜像。用户可以配置自定义仓库根地址。

## 索引与匹配

`Filetree.json` 的来源与条目顺序表达上游质量顺序。Go 普通 `map` 没有稳定遍历顺序，因此解析器使用有序结构保存原始顺序，并将后出现的候选作为更高推荐等级。

索引结构为标准化姓名到候选头像列表。匹配顺序如下：

1. 演员主名称精确匹配，命中后不再混入别名候选。
2. 主名称未命中时，保留完整别名并按中英文逗号、顿号、分号、空白等常见分隔符拆分匹配。
3. 精确匹配仍未命中时，按相同顺序尝试简繁体转换结果。
4. 只规范首尾空白、连续空格、全角空格和英文大小写。
5. 不使用拼音、包含关系或编辑距离等模糊匹配。

## 头像策略

- `recommended`：选择上游顺序中等级最高的候选。
- `original`：选择等级最高的非 `AI-Fix` 候选；没有时回退到推荐候选。
- `aiFix`：选择等级最高的 `AI-Fix` 候选；没有时回退到推荐候选。
- `manual`：仅存在于单演员前端对话框，不作为后端批量策略。

## 接口

管理员接口：

- `GET /api/performerAvatarLibrary/status`
- `POST /api/performerAvatarLibrary/updateDataFile`
- `POST /api/performerAvatarLibrary/clearImageCache`
- `GET /api/performerAvatarLibrary/candidates/:performerId`
- `GET /api/performerAvatarLibrary/batchActors/:performerBasesId/:page/:limit`
- `GET /api/performerAvatarLibrary/batchActorIds/:performerBasesId`
- `GET /api/performerAvatarLibrary/preview/:performerId/:candidateId`
- `POST /api/performerAvatarLibrary/apply`
- `POST /api/performerAvatarLibrary/batchPreview`
- `POST /api/performerAvatarLibrary/batchApply`
- `GET /api/performerAvatarLibrary/batchProgress/:batchId`

批量演员列表由服务端分页，分页数量直接使用演员列表当前的响应式 `pageSize`；“选择全部筛选结果”
通过独立 ID 接口完成，不把全部演员资料一次性载入浏览器。`batchApply` 只创建异步任务并返回任务
ID，前端每 3 秒查询一次进度，完成后显示成功、失败、未匹配与跳过数量。

预览图片通过项目统一请求层发起带管理员令牌的 Blob 请求。服务端只代理本地索引中确实属于目标
演员的候选，不能代理任意 URL；单人候选图最多同时加载 4 张，失败时保留具体原因并允许点击重试。

## 自动清理与计划任务

- `performerAvatarLibrary.clearCacheOnStartup` 默认关闭，由 `core.getDefaultConfig()` 提供默认值并经
  统一应用配置接口保存。
- 应用完成核心配置和数据库初始化后调用 `InitPerformerAvatarLibrary`；启用启动清理时由该入口
  调用统一清理函数。
- 计划任务类型 `clearPerformerAvatarCache` 是全局任务，不绑定文件库。创建、编辑、手动执行和
  Cron 调度继续使用现有计划任务模型、运行状态互斥及执行结果记录。
- 清理通过缓存代次和读写锁与并发下载协调：清理不等待远程下载，清理前已开始的下载不会在清理
  后重新写入旧代次缓存。

## 安全与失败恢复

- 索引最大下载大小为 32 MiB，单张图片最大 20 MiB。
- 数据文件、图片请求分别设置超时。
- 候选下载完成后验证 MIME 类型为图片。
- 远程候选图片按带时间戳的候选 ID 缓存在本地；遇到 429 或服务端错误时短暂退避重试，并继续尝试备用镜像。
- 替换头像时先保存新文件并更新数据库，成功后才删除旧头像。
- 批量处理固定三个下载工作线程，避免瞬间产生大量上游请求。
- 更新过程中只允许一个文件树更新任务。

## 验收

- 未初始化时使用单人或批量功能会询问是否立即下载；确认后下载并继续，取消后终止本次操作。
- 数据文件更新失败不会破坏已有索引。
- 单演员可以查看全部候选，并按四种策略选择。
- 批量界面不出现手动策略，执行前只弹出一次确认。
- 默认批量处理跳过已有头像；启用覆盖后明确警告。
- 头像更新不会改变演员其他字段。
- GitHub Raw 不可用时可以切换备用镜像。
- 设置页显示当前缓存头像数量和空间占用；手动清理需要二次确认并反馈清理数量。
- 开启启动清理后，重启软件会清空头像下载缓存但保留头像库索引和演员正式头像。
- “清理演员头像缓存”计划任务不要求选择文件库，并复用手动清理的后端函数。
