# AI 自动设置已有标签设计方案

## 背景

CM Collectors 3 当前用于收藏和管理视频、图片等资源。资源模型为资源级管理：一个 `resources` 资源下可以包含多个 `resourcesDramaSeries` 视频/分集，标签通过 `resourcesTags` 关联到资源本身。

用户已在系统中维护了一套标签和标签分类，希望 AI 自动分析资源内容，并从已有标签中选择合适标签写入资源，减少手动打标签成本。

本方案强调：AI 只负责自动选择已有标签，不创建新标签，不修改标签体系。

## 目标

- 在设置中提供独立的“AI 自动标签”功能。
- 支持配置 AI 调用地址、模型、密钥等调用参数。
- 支持按文件库启用或关闭 AI 自动标签。
- AI 从已有 `tag` / `tagClass` 中选择标签，并自动写入现有 `resourcesTags`。
- 记录资源是否已经 AI 分析过、分析状态、写入结果和失败原因。
- 通过计划任务执行后台分析，避免依赖前台页面保持打开。
- 支持为已有标签维护 AI 解释，解决标签名为缩写时 AI 难以理解的问题。
- 支持可配置的视频截图策略，在准确率、成本和速度之间取得平衡。

## 非目标

- 不创建新的业务标签。
- 不做视频/分集级标签，标签仍然是资源级。
- 不要求人工确认后再写入。
- 不训练本地模型。
- 第一版不做语音识别、字幕分析或向量检索。

## 现有结构关联

- `resources`：资源主体。
- `resourcesDramaSeries`：资源下的视频/分集，包含视频路径 `src`。
- `tag`：已有标签。
- `tagClass`：已有标签分类。
- `resourcesTags`：资源与标签的现有关联表，AI 最终写入这里。
- `processorsFFmpeg`：已有 FFmpeg 抽帧能力，可复用。
- `cronJobs`：已有计划任务体系，可新增任务类型。
- `videoFingerprint`：已有后台视频派生数据处理模式，可作为 AI 标签任务参考。

## 总体流程

```text
用户在设置中启用 AI 自动标签
  -> 选择启用的文件库
  -> 配置 AI 调用地址、模型和截图策略
  -> 计划任务定期扫描待分析资源
  -> 后端读取资源信息、已有标签池和标签解释
  -> 从资源下视频抽取截图
  -> 调用 AI，让 AI 只能返回已有 tagId
  -> 后端校验 AI 输出
  -> 自动写入 resourcesTags
  -> 写入 AI 分析记录
```

## 数据库设计

### 标签 AI 解释

建议给现有 `tag` 表新增字段：

```text
aiDescription text
aiEnabled bool default true
```

用途：

- 给 AI 解释已有标签的真实含义。
- 标签名为缩写时，以 `aiDescription` 为主要判断依据。
- 不改变标签名称，不新增标签类型。
- `aiEnabled` 控制该标签是否参与 AI 自动标签，必须默认 `true`。

示例：

```text
标签名：OL
AI解释：画面中人物服装或场景明显偏办公室、职场、通勤风格时选择。

标签名：JK
AI解释：画面中出现日式学生制服、校园制服风格时选择。
```

### 标签 AI 解释维护位置

标签 AI 解释应作为现有标签的属性维护，不应在 AI 自动标签页里重新做一套标签系统。

推荐主入口：

```text
设置 -> 标签设置 -> 编辑标签
```

在现有标签编辑弹窗中增加字段：

```text
AI 识别说明
是否参与 AI 自动标签
```

字段说明：

- `AI 识别说明`：填写给 AI 使用的自然语言解释。
- `是否参与 AI 自动标签`：默认开启；关闭后，该标签不会出现在 AI 可选标签池中。

迁移兼容规则：

```text
已有标签在新增 aiEnabled 字段后，默认全部参与 AI 自动标签。
新建标签默认参与 AI 自动标签。
只有用户明确关闭 aiEnabled 时，标签才会被排除。
```

示例：

```text
标签名：OL
AI 识别说明：办公室、职场、通勤服装或职业场景。

标签名：JK
AI 识别说明：日式学生制服、校园制服风格。

标签名：SS
AI 识别说明：短片、片段、剪辑合集，而不是完整长视频。
```

AI 自动标签页可以提供辅助入口，但不作为主维护入口：

```text
标签解释维护
未填写 AI 识别说明的标签：23 个
未启用 AI 自动标签的标签：5 个
```

点击后跳转到标签设置，或打开一个专门的批量维护视图。

### 批量维护标签解释

由于已有标签中可能存在大量缩写，建议提供批量维护视图：

```text
标签分类 | 标签名 | AI识别说明 | 是否参与AI自动标签
```

能力建议：

- 按文件库筛选。
- 按标签分类筛选。
- 只看未填写 AI 识别说明的标签。
- 支持连续编辑多个标签说明。
- 支持批量开启或关闭“参与 AI 自动标签”。
- 后续可支持导入/导出 JSON 或表格，方便一次性整理大量缩写标签。

批量导出格式示例：

```json
[
  {
    "tagId": "tag_001",
    "tagClassName": "服装",
    "name": "OL",
    "aiDescription": "办公室、职场、通勤服装或职业场景。",
    "aiEnabled": true
  }
]
```

### 标签导入导出

标签 TXT 导入导出需要支持 AI 字段，便于批量迁移和维护 `aiDescription` / `aiEnabled`。

导出格式使用 v2 文本格式：

```text
// CM Collectors 标签数据 v2

分类名
-标签名
  aiEnabled: true
  aiDescription: 适合 AI 理解该标签含义的说明
```

兼容规则：

- 旧格式仍可导入：
  ```text
  分类名
  -标签名
  ```
- AI 属性可省略。
- 新建标签未写 `aiEnabled` 时默认 `true`。
- 新建标签未写 `aiDescription` 时默认空字符串。
- 已有标签未写 `aiEnabled` 时保留原值。
- 已有标签未写 `aiDescription` 时保留原值。
- 写 `aiDescription:` 空值表示主动清空已有说明。
- 重命名语法继续支持：
  ```text
  分类旧名=>分类新名
  -旧标签名=>新标签名
    aiEnabled: false
    aiDescription: 新说明
  ```

导入导出只处理 AI 字段，不导入导出标签本身启用状态、排序、热度等普通业务字段，避免误改标签显示和排序。

### AI 使用标签解释的规则

构造 AI 标签池时：

```text
1. 只包含当前文件库下的标签。
2. 只包含 status = true 的标签。
3. 只包含 aiEnabled = true 的标签。
4. 受文件库 include_tag_class_ids / exclude_tag_class_ids 限制。
5. 标签名和 AI 识别说明都发送给 AI。
```

Prompt 中必须明确：

```text
标签名可能是缩写。
判断标签含义时优先参考 AI 识别说明。
如果标签没有 AI 识别说明，可以参考标签名和标签分类，但需要更谨慎。
```

如果标签缺少 `aiDescription`：

- 不建议直接禁止使用，因为部分中文全称标签可能不需要解释。
- 但应在设置页提示未填写数量。
- 对纯缩写标签，应优先补充解释后再启用 AI 自动标签。

### 标签解释变更后的重扫

标签 AI 解释会影响 AI 判断，因此应纳入 `tag_version_hash`。

当以下内容变化时：

```text
- 标签名
- 标签分类名
- AI 识别说明
- 是否参与 AI 自动标签
- 文件库标签分类 include/exclude 设置
```

系统应认为标签池版本已变化。已有成功分析记录不必立即自动重扫，但设置页应提供：

```text
标签解释已变更，建议重扫相关文件库
```

用户点击重扫后，相关资源重新进入待分析队列。

### AI 自动标签设置表

建议新增 `ai_tag_settings`：

```text
id
enabled
provider
base_url
api_key
model
max_resources_per_run
max_frames_per_resource
max_frames_per_video
max_videos_per_resource
max_images_per_ai_request
frame_strategy
image_resize_mode
fallback_image_max_width
image_jpeg_quality
min_confidence
max_tags_per_resource
write_mode
paused
request_timeout_seconds
created_at
updated_at
```

字段说明：

- `enabled`：总开关。
- `base_url`：AI 调用地址。
- `api_key`：AI 服务密钥。
- `model`：模型名。
- `max_resources_per_run`：每次计划任务最多处理的资源数。
- `max_frames_per_resource`：每个资源最多截图数量。
- `max_frames_per_video`：每个视频最多截图数量。
- `max_videos_per_resource`：每个资源最多分析的视频数量。
- `max_images_per_ai_request`：每次 AI 请求最多携带的图片数量，用于把同一个资源的截图分段发送给 AI。
- `frame_strategy`：截图策略，例如 `quick` / `high_accuracy_adaptive` / `ultra_accuracy`。
- `image_resize_mode`：图片处理方式，建议支持 `original` / `auto_fallback` / `fixed_resize`。
- `fallback_image_max_width`：触发降级时的最大宽度。
- `image_jpeg_quality`：JPEG 质量，默认 `original`，触发降级时按设置处理。
- `min_confidence`：最低写入置信度。
- `max_tags_per_resource`：每个资源最多写入标签数量。
- `write_mode`：写入策略。
- `paused`：暂停标记；暂停只阻止继续领取下一个资源，不中断正在处理的当前资源。
- `request_timeout_seconds`：单次 AI 请求超时时间；本地大模型可能需要较长时间，默认应明显大于普通 HTTP 请求超时。

`write_mode` 建议取值：

```text
append     追加 AI 标签，不删除已有标签
replace    用 AI 标签替换已有标签
only_empty 只处理没有任何标签的资源
```

默认建议：

```text
write_mode = append
```

### 文件库启用表

建议新增 `ai_tag_enabled_files_bases`：

```text
id
files_bases_id
enabled
include_tag_class_ids
exclude_tag_class_ids
created_at
updated_at
```

字段说明：

- `files_bases_id`：文件库 ID。
- `enabled`：该文件库是否启用 AI 自动标签。
- `include_tag_class_ids`：允许 AI 使用的标签分类，可存 JSON。
- `exclude_tag_class_ids`：排除 AI 使用的标签分类，可存 JSON。

如果 `include_tag_class_ids` 为空，则默认允许该文件库下全部可用标签分类；再根据 `exclude_tag_class_ids` 排除。

### AI 分析记录表

建议新增 `ai_tag_analysis_records`：

```text
id
resources_id
files_bases_id
status
src_hash
tag_version_hash
prompt_version
model
recommended_tag_ids
written_tag_ids
result_json
evidence_json
fail_reason
analyzed_at
created_at
updated_at
```

状态建议：

```text
pending
processing
success
failed
skipped
```

字段说明：

- `resources_id`：资源 ID。
- `files_bases_id`：文件库 ID。
- `src_hash`：资源下视频路径、数量、排序等内容生成的 hash。
- `tag_version_hash`：参与 AI 分析的标签、分类、AI 解释生成的 hash。
- `prompt_version`：Prompt 版本。
- `recommended_tag_ids`：AI 推荐的合法标签 ID 列表。
- `written_tag_ids`：最终写入 `resourcesTags` 的标签 ID 列表。
- `result_json`：AI 原始结构化结果。
- `evidence_json`：本次分析使用的资源标题、视频路径、截图位置等证据信息。
- `fail_reason`：失败原因。
- `analyzed_at`：分析完成时间。

`src_hash` 和 `tag_version_hash` 用于判断是否需要重新分析。

## 后端模块建议

建议新增：

```text
cm_collectors_server/models/aiTagSetting.models.go
cm_collectors_server/models/aiTagEnabledFilesBases.models.go
cm_collectors_server/models/aiTagAnalysisRecord.models.go
cm_collectors_server/processors/aiTag.processors.go
cm_collectors_server/controllers/aiTag.controllers.go
cm_collectors_server/processorsAI/client.go
cm_collectors_server/processorsAI/tagPrompt.go
cm_collectors_server/processorsAI/types.go
```

职责划分：

- `models`：保存配置、启用文件库和分析记录。
- `processors/aiTag`：扫描资源、抽帧、调用 AI、校验结果、写入标签。
- `processorsAI`：封装 AI 服务请求和 Prompt 构造。
- `controllers`：提供设置页和手动触发使用的 API。

## API 设计

建议前缀：

```text
/api/aiTag
```

接口建议：

```text
GET /aiTag/setting
PUT /aiTag/setting
GET /aiTag/filesBases
PUT /aiTag/filesBases
GET /aiTag/stats
GET /aiTag/records
POST /aiTag/runOnce
POST /aiTag/resetFailed
POST /aiTag/resetProcessing
POST /aiTag/pause
POST /aiTag/resume
POST /aiTag/rescan
POST /aiTag/testService
POST /aiTag/testConnection
```

说明：

- `runOnce`：立即执行一次后台分析，仍由后端任务执行，不依赖前台保持打开。
- `resetFailed`：将失败记录重置为待分析。
- `resetProcessing`：仅用于服务异常关闭后，把残留的 `processing` 记录恢复为待处理；普通继续执行不应自动重置 `processing`。
- `pause`：设置暂停标记。暂停后当前正在分析的资源继续跑完，后续不再领取新资源。
- `resume`：取消暂停标记；前端可在恢复后重新启动所选文件库的待分析任务。
- `rescan`：按文件库重置分析记录，使资源可重新分析。
- `testService`：快速检查 `/models` 服务是否可访问、模型名是否存在。
- `testConnection`：实际发起一次模型输出测试，确认模型可以按要求返回结构化结果。

## 计划任务设计

新增计划任务类型：

```text
aiTag
```

执行逻辑：

```text
1. 检查 AI 自动标签总开关。
2. 查询启用 AI 自动标签的文件库。
3. 按文件库扫描待分析资源。
4. 每轮最多处理 max_resources_per_run 个资源。
5. 对每个资源：
   - 建立或更新分析记录为 processing。
   - 读取资源信息、分集视频、标签池和标签解释。
   - 抽取视频截图。
   - 调用 AI。
   - 校验 AI 返回的 tagId。
   - 根据写入策略写入 resourcesTags。
   - 更新分析记录为 success 或 failed。
```

暂停语义：

```text
- 暂停只在领取下一个资源前生效。
- 已进入 analyzeResource 的当前资源继续完成完整流程。
- 当前资源完成后必须写入 success 或 failed，不应因为暂停残留 processing。
- 普通 runOnce/resume 不自动把 processing 重置为 pending。
- 只有用户明确点击“恢复中断”时，才把服务异常关闭遗留的 processing 记录恢复为 pending。
```

待分析资源规则：

```text
- 文件库启用 AI 自动标签。
- 资源状态有效。
- 资源类型为 movies 或 videoLink。
- 没有分析记录，或没有 success 记录。
- 或 src_hash 与 success 记录不一致。
- 或 tag_version_hash 与 success 记录不一致。
- processing 记录不会被普通扫描重复领取。
```

`tag_version_hash` 稳定性要求：

```text
- 参与 hash 的标签分类必须按 sort、id 稳定排序。
- 每个分类下的标签必须按 sort、id 稳定排序。
- hash 输入包含标签名、分类名、aiDescription、aiEnabled、include/exclude 设置。
- 禁止直接使用 Go map 遍历顺序生成 hash，否则会导致无变化时重复分析。
```

默认失败策略：

```text
failed 不自动重复重试。
用户在设置页点击“重置失败”后重新进入待分析。
```

## AI 输入设计

AI 输入应包含：

- 资源标题。
- 番号/识别号。
- 资源简介。
- 国家、清晰度等结构字段。
- 演员/导演名称。
- 视频文件名或路径末尾，避免发送完整敏感路径。
- 资源封面。
- 视频截图。
- 参与 AI 的已有标签列表。

如果一个资源截图数量较多，不应一次性全部发送给 AI。应按 `max_images_per_ai_request` 分段发送。每个分段请求都带上完整资源信息、完整标签池和标签 AI 解释，但只携带当前分段的截图。

`max_images_per_ai_request` 控制的是单次请求压力，不会丢弃图片。一个资源的所有批次结果都会进入后端汇总；值越小，单次上下文压力越低、请求次数越多，适合 8192 context 或标签池较大的本地模型。

标签输入格式示例：

```json
[
  {
    "tagClassId": "class_001",
    "tagClassName": "服装",
    "tags": [
      {
        "tagId": "tag_001",
        "name": "OL",
        "aiDescription": "画面中人物服装或场景明显偏办公室、职场、通勤风格时选择。"
      }
    ]
  }
]
```

Prompt 关键约束：

```text
只能从给定标签列表中选择 tagId。
禁止创建新标签。
禁止返回标签名以外的自造标签。
如果证据不足，返回空 tags 数组并设置 uncertain=true。
标签名可能是缩写，判断时优先参考 aiDescription。
必须只返回一个合法 JSON object。
禁止 Markdown、代码块、解释文字、前后缀文本。
```

## AI 输出设计

AI 返回格式：

```json
{
  "batchIndex": 1,
  "tags": [
    {
      "tagId": "tag_001",
      "confidence": 0.91,
      "reason": "关键帧中多次出现办公室场景和职业服装。"
    }
  ],
  "summary": "资源整体符合职场、室内场景。",
  "uncertain": false
}
```

AI 请求格式策略：

```text
优先使用 OpenAI 兼容的 response_format=json_schema，约束 batchIndex/tags/summary/uncertain。
如果服务端或模型返回空响应、JSON 解析失败，后端自动回退到 response_format=text 再请求一次。
不强制设置 max_tokens，由用户在模型服务端自行控制输出长度。
```

后端校验规则：

- `tagId` 必须存在。
- `tagId` 必须属于当前文件库可用标签。
- `tagId` 必须属于允许 AI 使用的标签分类。
- `confidence` 必须大于等于 `min_confidence`。
- 去重。
- 最多保留 `max_tags_per_resource` 个标签。
- AI 返回非法 JSON 时记录失败。

### 分段分析与结果汇总

当一个资源截图数超过 `max_images_per_ai_request` 时，后端应将该资源拆成多个 AI 分析请求：

```text
资源截图总数：80
max_images_per_ai_request：5
请求批次数：16
```

每个批次请求包含：

```text
- 相同的资源标题、简介、番号、演员、导演等基础信息。
- 相同的标签池和标签 AI 解释。
- 当前批次的截图。
- 当前批次截图对应的视频 ID、视频序号、时间点和百分比位置。
```

每个批次输出候选标签：

```json
{
  "batchIndex": 1,
  "tags": [
    {
      "tagId": "tag_001",
      "confidence": 0.86,
      "reason": "本批次截图中多次出现办公室场景。"
    }
  ],
  "summary": "本批次主要是室内职场场景。",
  "uncertain": false
}
```

后端需要将批次结果汇总成资源级标签：

```text
1. 合并所有批次返回的合法 tagId。
2. 统计每个标签出现的批次数、最高置信度、平均置信度和证据时间点。
3. 丢弃低于 min_confidence 的标签。
4. 按最终分数排序。
5. 最多保留 max_tags_per_resource 个标签。
```

推荐汇总分数：

```text
final_score = max_confidence * 0.5 + avg_confidence * 0.3 + evidence_batch_ratio * 0.2
```

其中：

```text
evidence_batch_ratio = 推荐该标签的批次数 / 总批次数
```

如果批次数较多，也可以增加一次“最终汇总请求”：

```text
1. 前 N 个批次分别分析截图。
2. 后端收集每批 summary、候选标签、置信度、证据时间点。
3. 再调用一次 AI，只发送文本汇总，不再发送图片。
4. 让 AI 从候选标签中做最终取舍。
```

最终汇总请求仍然只能返回已有 `tagId`，不能创建新标签。第一版可以先使用后端规则汇总，后续再增加 AI 文本汇总。

分析记录中的 `evidence_json` 应保存分段信息：

```json
{
  "totalFrames": 80,
  "maxImagesPerRequest": 5,
  "batches": [
    {
      "batchIndex": 1,
      "frames": [
        {
          "dramaSeriesId": "ds_001",
          "videoIndex": 1,
          "positionPercent": 3,
          "timeSeconds": 216
        }
      ]
    }
  ]
}
```

这样可以减少单次 AI 请求压力，同时保留对长视频的大量截图覆盖；后续排查标签来源时，也能知道某个标签由哪些视频片段支持。

## 自动写入策略

默认策略为 `append`：

```text
最终标签 = 现有标签 + AI 合法推荐标签
```

`append` 不删除人工已有标签，风险最低。

`replace` 仅建议高级用户开启：

```text
最终标签 = AI 合法推荐标签
```

`only_empty` 适合初始批量整理：

```text
只有资源没有任何标签时才分析并写入。
```

写入完成后更新：

```text
recommended_tag_ids
written_tag_ids
result_json
evidence_json
status = success
analyzed_at
```

分析记录落库要求：

```text
- 写入 processing、success、failed 时都应处理 SQLite busy 重试。
- 如果 AI 分析、标签写入或结果解析失败，应写入 failed 和 fail_reason。
- 如果当前资源流程正常完成，应写入 success，即使用户在处理中点击了暂停。
- success 记录只有在 src_hash 或稳定的 tag_version_hash 变化时才重新进入待分析。
```

## 截图策略

目标以准确率优先，分析耗时和调用成本作为次要约束。由于部分资源的视频可能长达 2 小时以上，且内容变化非常频繁，固定抽取少量截图不足以支撑可靠打标签。截图策略应支持“高密度时长覆盖”和“内容差异补充采样”，而不是只使用固定少量截图。

默认建议使用“高准确自适应”：

```text
max_frames_per_resource = 80
max_frames_per_video = 40
max_videos_per_resource = 12
max_images_per_ai_request = 5
image_resize_mode = original
image_jpeg_quality = original
min_sample_interval_seconds = 120
```

图片默认不缩略，优先使用 FFmpeg 输出的原始帧分辨率。只有在 AI 服务限制单张图片大小、总请求大小或模型不接受原始分辨率时，才按配置降级压缩。

### 按视频时长自适应抽帧

每个视频的截图数量不应固定，而应根据视频时长计算：

```text
小于 5 分钟：6 张
5 到 20 分钟：10 张
20 到 60 分钟：16 张
60 到 120 分钟：24 张
120 到 180 分钟：32 张
180 分钟以上：40 张
```

截图位置应均匀覆盖主体内容，并避开片头片尾：

```text
起始范围：3%
结束范围：97%
在 3% 到 97% 之间等距抽样。
```

示例：

```text
2 小时视频，抽 32 张：
3%、6%、9%、12%、15%、18%、21%、24%、27%、30%、33%、36%、39%、42%、45%、48%、
52%、55%、58%、61%、64%、67%、70%、73%、76%、79%、82%、85%、88%、91%、94%、97%
```

### 多视频资源分配策略

一个资源下有多个视频时，先按视频时长为每个视频分配基础截图数，再受 `max_frames_per_resource` 限制。

```text
1. 统计资源下所有视频时长。
2. 每个视频至少抽 4 张。
3. 长视频按时长获得更多截图配额。
4. 总截图数不超过 max_frames_per_resource。
5. 如果视频数量过多，优先覆盖：
   - 第 1 个视频
   - 时长最长的视频
   - 中间位置的视频
   - 最后 1 个视频
   - 其余按时长从高到低补齐

```

例如一个资源包含 4 个视频：

```text
视频 A：10 分钟 -> 4 张
视频 B：35 分钟 -> 12 张
视频 C：130 分钟 -> 32 张
视频 D：8 分钟 -> 8 张
总计 56 张
```

截图处理建议：

- 避开视频开头和结尾，减少黑屏、片头、片尾干扰。
- 对长视频使用更密集的时间覆盖，避免只看到局部内容。
- 图片默认不缩略，保留原始分辨率。
- 图片默认不降低 JPEG 质量。
- 仅当 AI 服务接口限制单张图片大小或请求总大小时，才启用降级压缩。
- 小于 5 秒的视频可以跳过或只截 50% 一张。

设置页可提供三档：

```text
快速：每资源最多 24 张，每视频最多 12 张，最多 6 个视频，可缩略
标准高准确：每资源最多 80 张，每视频最多 40 张，最多 12 个视频，默认原始分辨率
极高准确：每资源最多 160 张，每视频最多 80 张，不限制视频数量，默认原始分辨率
```

默认使用“标准高准确”。

### 二阶段分析策略

为兼顾成本和准确率，建议支持二阶段分析：

```text
第一阶段：快速分析
- 先抽 24 到 40 张代表帧。
- 如果 AI 返回高置信度标签，则直接写入。

第二阶段：增强分析
- 如果 AI 返回 uncertain=true、推荐标签为空、或最高置信度低于阈值。
- 自动追加抽帧到 80 或 160 张。
- 使用更多截图重新分析。
```

这样短视频和内容单一的视频可以较快完成，长视频或内容复杂的视频会自动提高截图覆盖率。若用户选择极高准确模式，可以跳过快速阶段，直接使用最大截图量分析。

二阶段分析也应使用分段请求。例如第一阶段抽 40 张图、每次喂 10 张，则分 4 次请求；第二阶段追加到 80 张时，只需分析新增批次，或重新按完整截图集合分批分析。

### 内容变化频繁视频的增强策略

后续可引入基于画面差异的采样优化：

```text
1. 先按较高频率抽取候选帧，例如每 1 到 2 分钟一张。
2. 对候选帧计算感知哈希或颜色/边缘差异。
3. 去掉高度相似、黑屏、模糊、片头片尾帧。
4. 选择差异最大的前 N 张发给 AI。
```

项目已有视频指纹和帧提取能力，后续可以复用相近思路做“差异帧筛选”。如果以准确度为第一优先级，第一版也可以直接实现“高密度等距抽帧 + 差异帧筛选”，避免把大量重复画面发给 AI，同时保留内容变化最大的帧。

### 图片降级策略

默认不缩略截图。只有遇到以下情况才降级：

```text
- AI 服务限制单张图片大小。
- AI 服务限制请求总大小。
- 图片数量过多导致请求超限。
- 用户在设置中主动选择快速模式。
```

降级顺序：

```text
1. 优先降低 JPEG 质量到 90。
2. 再限制最大宽度到 1440。
3. 再限制最大宽度到 1080。
4. 最后才限制到 768。
```

不建议默认使用 768 缩略图，因为细节类标签可能依赖服装、物品、场景文字或远景元素。

## 设置页设计

建议在设置页新增 Tab：

```text
AI 自动标签
```

区域：

```text
基础设置
- 启用 AI 自动标签
- AI 调用地址
- API Key
- 模型
- 测试连接

文件库设置
- 文件库列表
- 是否启用
- 参与标签分类
- 排除标签分类

写入策略
- 追加到现有标签
- 仅无标签资源
- 替换现有标签
- 最低置信度
- 每资源最多写入标签数量

截图策略
- 快速 / 标准高准确 / 极高准确
- 最大截图数
- 每次 AI 请求图片数
- 最大视频数
- 图片处理模式：原始分辨率 / 自动降级 / 固定缩略
- 降级后的最大宽度
- 降级后的 JPEG 质量

任务状态
- 待分析数量
- 已成功数量
- 失败数量
- 上次运行时间
- 立即执行一次
- 暂停
- 继续执行
- 重置失败
- 恢复中断
- 重扫当前文件库
```

按钮行为：

```text
立即执行：当前没有任务运行且未暂停时可用。
暂停：仅任务运行时可用；不会中断当前 AI 请求，只阻止继续领取新资源。
继续执行：取消暂停，并可重新启动所选文件库的待分析任务。
恢复中断：仅用于服务异常关闭后恢复残留 processing 记录，不用于普通暂停/继续。
任务状态页：任务运行时自动定时刷新。
```

标签 AI 解释维护主入口放在现有标签设置中，AI 自动标签页只提供状态提示和跳转入口：

```text
未填写 AI 识别说明的标签数量
未启用 AI 自动标签的标签数量
去维护标签解释
```

批量维护视图建议展示：

```text
标签分类 | 标签名 | AI识别说明 | 是否参与AI自动标签
```

## 安全和隐私

设置页需要明确提示：

```text
AI 自动标签会把资源标题、标签列表、标签解释和视频截图发送到配置的 AI 服务。
```

建议默认只发送路径末尾文件名，不发送完整本地路径。

API Key 保存应尽量避免在前端列表中明文回显。前端读取设置时可返回脱敏值：

```text
sk-****abcd
```

## 默认配置建议

```text
enabled = false
write_mode = append
min_confidence = 0.70
max_tags_per_resource = 8
max_frames_per_resource = 80
max_frames_per_video = 40
max_videos_per_resource = 12
max_images_per_ai_request = 5
image_resize_mode = original
image_jpeg_quality = original
max_resources_per_run = 20
frame_strategy = high_accuracy_adaptive
```

## 分阶段实施

### 第一阶段：核心闭环

- 新增标签 `aiDescription` 字段。
- 新增 AI 设置、文件库启用和分析记录表。
- 新增 AI 自动标签计划任务类型。
- 实现单轮后台自动分析。
- 实现 AI 输出校验。
- 自动写入现有 `resourcesTags`。
- 设置页支持基础配置和立即执行。

### 第二阶段：可维护性增强

- 设置页展示分析记录、失败原因和统计信息。
- 支持重置失败和按文件库重扫。
- 支持标签 AI 解释批量维护。
- 支持 include/exclude 标签分类。

### 第三阶段：准确率和成本优化

- 支持截图策略三档配置。
- 支持按视频时长自适应抽帧。
- 支持按 `max_images_per_ai_request` 将同一资源截图分段发送给 AI，并汇总批次结果。
- 支持二阶段分析，低置信度时自动追加截图重新分析。
- 支持基于画面差异的候选帧筛选。
- 支持原始分辨率截图输入，并仅在接口限制时自动降级。
- 支持资源封面作为额外输入。
- 支持只分析未打标签资源或缺少指定分类标签的资源。
- 支持 prompt 版本管理和模型切换后重扫。

## 验收标准

- AI 不会创建任何新标签。
- AI 返回不存在的 `tagId` 不会写入。
- 启用文件库中的资源能通过计划任务自动分析。
- 分析成功后，合法标签会写入现有 `resourcesTags`。
- 已成功分析的资源不会在无变化时重复消耗 AI 调用。
- `tag_version_hash` 在标签池无变化时保持稳定，不受 map 遍历顺序影响。
- 点击暂停后，当前资源继续完成并最终写入 `success` 或 `failed`，不会因暂停残留 `processing`。
- 视频路径变化后可重新进入待分析。
- 标签解释变化后可通过重扫重新分析。
- 标签 TXT 导入导出支持可选 `aiEnabled` 和 `aiDescription`，旧格式仍可导入。
- AI 返回优先使用 `json_schema` 约束；若本地兼容服务返回空响应或解析失败，可回退 text 模式重试。
- 前台页面关闭后，只要后端服务运行，计划任务仍可继续执行。
- 设置页能看到成功、失败、待分析数量和失败原因。
