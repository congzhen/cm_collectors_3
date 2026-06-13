<template>
  <div class="ai-tag-setting" v-loading="loading">
    <div class="feature-intro">
      <div class="feature-intro__title">AI 自动标签</div>
      <div class="feature-intro__text">
        根据资源下的视频截图，让 AI 从当前文件库已有标签中选择匹配项并自动写入资源标签。该功能不会创建新标签，也不会按单个视频分别打标签；默认追加写入，已分析且内容与标签池未变化的资源会跳过。
      </div>
      <div class="feature-intro__text">
        使用前请先选择启用文件库，补充标签的 AI 识别说明，并配置支持图片输入的 OpenAI 兼容模型。计划任务和立即执行都会使用这里的设置；暂停后可继续执行所选文件库的待分析资源。
      </div>
    </div>
    <div class="toolbar">
      <div class="toolbar__actions">
        <el-button icon="Connection" @click="testService">测试服务</el-button>
        <el-button icon="ChatLineRound" :loading="modelTestLoading" @click="testConnection">测试模型输出</el-button>
        <el-button type="primary" icon="Check" @click="saveAll">保存设置</el-button>
        <el-button type="success" icon="VideoPlay" :disabled="isTaskRunning || setting.paused"
          @click="runOnce">立即执行</el-button>
        <el-button v-if="!setting.paused" type="warning" icon="VideoPause" :disabled="!isTaskRunning"
          @click="pauseTask">暂停</el-button>
        <el-button v-else type="success" plain icon="VideoPlay" @click="resumeTask">继续执行</el-button>
      </div>
      <div class="toolbar__settings">
        <el-switch v-model="setting.enabled" active-text="启用 AI 自动标签" />
        <el-select v-model="selectedFilesBasesIds" multiple clearable collapse-tags collapse-tags-tooltip
          placeholder="启用文件库" class="files-bases-select">
          <el-option v-for="item in store.filesBasesStoreData.filesBasesStatus" :key="item.id" :label="item.name"
            :value="item.id" />
        </el-select>
      </div>
    </div>

    <el-dialog v-model="modelTestDialogVisible" title="测试模型输出" width="920px"
      :close-on-click-modal="!modelTestLoading">
      <div class="model-test-dialog" v-loading="modelTestLoading" element-loading-text="正在测试模型输出，请等待">
        <div v-if="modelTestLoading" class="model-test-loading">
          正在向当前表单中的模型发送一次轻量 JSON 输出请求，完成后会显示耗时、token、速度和推荐参数。
        </div>
        <template v-else-if="modelTestResult">
          <el-alert :type="modelTestResult.success ? 'success' : 'error'" :closable="false" show-icon>
            <template #title>
              {{ modelTestResult.success ? '模型输出测试成功' : '模型输出测试失败' }}
            </template>
            <div class="alert-help">
              {{ modelTestResult.success ? '当前模型可以返回 AI 自动标签需要的 JSON 输出。' : modelTestResult.error }}
            </div>
          </el-alert>

          <div class="model-test-metrics">
            <div class="metric-item">
              <span>总耗时</span>
              <strong>{{ formatElapsed(modelTestResult.metrics.elapsedMs) }}</strong>
            </div>
            <div class="metric-item">
              <span>输入 Token</span>
              <strong>{{ metricText(modelTestResult.metrics.promptTokens, modelTestResult.metrics.usageReturned) }}</strong>
            </div>
            <div class="metric-item">
              <span>输出 Token</span>
              <strong>{{ metricText(modelTestResult.metrics.completionTokens, modelTestResult.metrics.usageReturned) }}</strong>
            </div>
            <div class="metric-item">
              <span>总 Token</span>
              <strong>{{ metricText(modelTestResult.metrics.totalTokens, modelTestResult.metrics.usageReturned) }}</strong>
            </div>
            <div class="metric-item">
              <span>输出速度</span>
              <strong>{{ speedText(modelTestResult) }}</strong>
            </div>
            <div class="metric-item">
              <span>响应格式</span>
              <strong>{{ modelTestResult.fallbackUsed ? 'text 兼容' : modelTestResult.responseFormat || '-' }}</strong>
            </div>
          </div>

          <div class="model-test-detail">
            <div>请求地址：{{ modelTestResult.endpoint || '-' }}</div>
            <div>模型：{{ modelTestResult.model || '-' }}</div>
            <div>结束原因：{{ modelTestResult.finishReason || '-' }}</div>
            <div v-if="modelTestResult.fallbackUsed">json_schema 首次失败：{{ modelTestResult.firstError || '-' }}</div>
            <div v-if="!modelTestResult.metrics.usageReturned">当前服务未返回标准 usage 字段，token 数和估算速度可能显示为 “-”。</div>
          </div>

          <div class="model-output" v-if="modelTestResult.content">
            <div class="model-output__title">模型返回</div>
            <pre>{{ modelTestResult.content }}</pre>
          </div>

          <div class="recommendation-header">
            <span>推荐参数</span>
            <el-button v-if="modelTestResult.recommendations?.length" type="primary" link @click="applyAllRecommendations">
              全部应用到表单
            </el-button>
          </div>
          <el-empty v-if="!modelTestResult.recommendations?.length" description="当前没有需要调整的推荐参数" />
          <el-table v-else :data="modelTestResult.recommendations" class="recommendation-table">
            <el-table-column prop="label" label="参数" width="120" />
            <el-table-column label="当前值" width="96">
              <template #default="{ row }">{{ formatRecommendationValue(row.currentValue) }}</template>
            </el-table-column>
            <el-table-column label="推荐值" width="96">
              <template #default="{ row }">{{ formatRecommendationValue(row.recommendedValue) }}</template>
            </el-table-column>
            <el-table-column prop="reason" label="原因" min-width="210">
              <template #default="{ row }">
                <div class="recommendation-text">{{ row.reason }}</div>
              </template>
            </el-table-column>
            <el-table-column prop="impact" label="影响" min-width="210">
              <template #default="{ row }">
                <div class="recommendation-text">{{ row.impact }}</div>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="76" align="center">
              <template #default="{ row }">
                <el-button type="primary" link @click="applyRecommendation(row)">应用</el-button>
              </template>
            </el-table-column>
          </el-table>
        </template>
      </div>
      <template #footer>
        <el-button @click="modelTestDialogVisible = false" :disabled="modelTestLoading">关闭</el-button>
        <el-button type="primary" @click="testConnection" :loading="modelTestLoading">重新测试</el-button>
      </template>
    </el-dialog>

    <el-tabs v-model="activeTab">
      <el-tab-pane label="基础设置" name="setting">
        <el-form :model="setting" label-width="160px" label-position="left" class="setting-form">
          <el-form-item label="AI 调用地址">
            <el-input v-model="setting.baseUrl" placeholder="https://api.openai.com/v1" />
            <div class="field-help">填写 OpenAI 兼容接口地址。可以填根地址、/v1，或完整的 /v1/chat/completions 地址。</div>
          </el-form-item>
          <el-form-item label="API Key">
            <el-input v-model="setting.apiKey" type="password" show-password placeholder="保存后脱敏显示" />
            <div class="field-help">仅后端调用 AI 时使用；保存后前端只显示脱敏值。留空保存时会保留原来的密钥。</div>
          </el-form-item>
          <el-form-item label="模型">
            <el-input v-model="setting.model" placeholder="gpt-4.1-mini" />
            <div class="field-help">选择支持图片输入的模型，否则无法根据视频截图识别标签。</div>
          </el-form-item>
          <el-form-item label="AI 请求超时">
            <el-input-number v-model="setting.requestTimeoutSeconds" :min="30" :max="86400" />
            <div class="field-help">单次 AI 请求最多等待多少秒。大模型建议设置较大，例如 1800 秒或更高。</div>
          </el-form-item>
          <el-form-item label="写入策略">
            <el-radio-group v-model="setting.writeMode">
              <el-radio-button label="追加" value="append" />
              <el-radio-button label="仅无标签资源" value="only_empty" />
              <el-radio-button label="替换" value="replace" />
            </el-radio-group>
            <div class="field-help">追加：保留人工标签，只新增 AI 命中的标签。仅无标签资源：已有任意标签的资源会跳过。替换：用 AI 结果覆盖资源当前标签，风险最高。</div>
          </el-form-item>
          <el-form-item label="截图策略">
            <el-radio-group v-model="setting.frameStrategy">
              <el-radio-button label="快速" value="quick" />
              <el-radio-button label="标准高准确" value="high_accuracy_adaptive" />
              <el-radio-button label="极高准确" value="ultra_accuracy" />
            </el-radio-group>
            <div class="field-help">快速：减少截图和调用次数。标准高准确：按视频时长自适应抽帧，默认推荐。极高准确：更偏向长视频和内容变化频繁的资源。</div>
          </el-form-item>
          <el-form-item label="图片模式">
            <el-radio-group v-model="setting.imageResizeMode">
              <el-radio-button label="原始分辨率" value="original" />
              <el-radio-button label="自动降级" value="auto_fallback" />
              <el-radio-button label="固定缩放" value="fixed_resize" />
            </el-radio-group>
            <div class="field-help">原始分辨率：优先准确度，不主动缩小截图。自动降级：AI 服务拒绝或超限时再压缩。固定缩放：始终按下方最大宽度压缩。</div>
          </el-form-item>
          <div class="number-grid">
            <el-form-item label="每轮资源数">
              <el-input-number v-model="setting.maxResourcesPerRun" :min="1" :max="500" />
              <div class="field-help">单次立即执行或计划任务最多处理多少个待分析资源，用来控制一次任务的总耗时。</div>
            </el-form-item>
            <el-form-item label="每资源最多截图">
              <el-input-number v-model="setting.maxFramesPerResource" :min="1" :max="500" />
              <div class="field-help">一个资源下所有视频合计最多抽多少张图。长资源建议保持较高，准确度更好。</div>
            </el-form-item>
            <el-form-item label="每视频最多截图">
              <el-input-number v-model="setting.maxFramesPerVideo" :min="1" :max="200" />
              <div class="field-help">单个视频最多抽帧数量。两小时以上、内容变化多的视频可以适当提高。</div>
            </el-form-item>
            <el-form-item label="每资源最多视频">
              <el-input-number v-model="setting.maxVideosPerResource" :min="1" :max="100" />
              <div class="field-help">一个资源包含很多视频时，最多选取多少个视频参与分析，避免异常资源拖慢整轮任务。</div>
            </el-form-item>
            <el-form-item label="每次请求图片数">
              <el-input-number v-model="setting.maxImagesPerAiRequest" :min="1" :max="50" />
              <div class="field-help">把同一资源的截图分批发给 AI。值越小单次压力越低，批次数会更多。</div>
            </el-form-item>
            <el-form-item label="最低置信度">
              <el-input-number v-model="setting.minConfidence" :min="0" :max="1" :step="0.05" />
              <div class="field-help">AI 返回标签的最低可信度。低于该值的标签会被后端丢弃，不写入资源。</div>
            </el-form-item>
            <el-form-item label="每资源最多写入">
              <el-input-number v-model="setting.maxTagsPerResource" :min="1" :max="50" />
              <div class="field-help">汇总所有批次后，一个资源最多自动写入多少个标签，防止标签过多。</div>
            </el-form-item>
            <el-form-item label="降级最大宽度">
              <el-input-number v-model="setting.fallbackImageMaxWidth" :min="320" :max="4096" />
              <div class="field-help">图片需要压缩时使用的最大宽度；原始分辨率模式下通常不会用到。</div>
            </el-form-item>
            <el-form-item label="JPEG 质量">
              <el-input-number v-model="setting.imageJpegQuality" :min="40" :max="100" />
              <div class="field-help">图片压缩时的 JPEG 质量。100 最接近原图，数值越低体积越小但细节会减少。</div>
            </el-form-item>
          </div>
        </el-form>
      </el-tab-pane>

      <el-tab-pane label="文件库" name="filesBases">
        <el-alert type="info" :closable="false" show-icon>
          <template #title>
            标签默认参与 AI 自动标签；建议为简写标签补充 AI 识别说明，当前未填写说明：{{ missingDescriptionCount }} 个。
          </template>
          <div class="alert-help">启用某个文件库后，计划任务和立即执行才会扫描该库的视频资源。分类筛选只影响喂给 AI 的标签池，不会创建新标签。</div>
        </el-alert>
        <el-table :data="filesBasesSettings" class="files-bases-table">
          <el-table-column label="启用" width="90">
            <template #default="{ row }">
              <el-switch v-model="row.enabled" @change="syncSelectedFromRows" />
            </template>
          </el-table-column>
          <el-table-column label="文件库" min-width="160">
            <template #default="{ row }">
              {{ store.filesBasesStoreData.getFilesBasesNameById(row.filesBasesId) }}
            </template>
          </el-table-column>
          <el-table-column label="只参与这些分类" min-width="220">
            <template #default="{ row }">
              <el-select v-model="row.includeTagClassIds" multiple clearable collapse-tags collapse-tags-tooltip
                placeholder="全部分类">
                <el-option v-for="item in tagClassOptions[row.filesBasesId] || []" :key="item.id" :label="item.name"
                  :value="item.id" />
              </el-select>
              <div class="table-help">不选择时表示全部分类都可参与；选择后只有这些分类下的标签会提供给 AI。</div>
            </template>
          </el-table-column>
          <el-table-column label="排除分类" min-width="220">
            <template #default="{ row }">
              <el-select v-model="row.excludeTagClassIds" multiple clearable collapse-tags collapse-tags-tooltip
                placeholder="不排除">
                <el-option v-for="item in tagClassOptions[row.filesBasesId] || []" :key="item.id" :label="item.name"
                  :value="item.id" />
              </el-select>
              <div class="table-help">这些分类会从 AI 标签池中移除，适合排除人工维护或不适合自动判断的分类。</div>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="任务状态" name="records">
        <div class="stats-row">
          <div class="stat-item"><span>处理中</span><strong>{{ stats.processing }}</strong></div>
          <div class="stat-item"><span>成功</span><strong>{{ stats.success }}</strong></div>
          <div class="stat-item"><span>失败</span><strong>{{ stats.failed }}</strong></div>
          <div class="stat-item"><span>跳过</span><strong>{{ stats.skipped }}</strong></div>
        </div>
        <div class="record-tools">
          <el-select v-model="recordStatus" clearable placeholder="全部状态" class="status-select" @change="loadRecords">
            <el-option label="处理中" value="processing" />
            <el-option label="成功" value="success" />
            <el-option label="失败" value="failed" />
            <el-option label="跳过" value="skipped" />
          </el-select>
          <el-button icon="Refresh" @click="reloadStatus">刷新</el-button>
          <el-button icon="RefreshLeft" :disabled="isTaskRunning" @click="resetFailed">重置失败</el-button>
          <el-button icon="RefreshRight" :disabled="isTaskRunning" @click="resetProcessing">恢复中断</el-button>
          <el-button type="warning" plain icon="Delete" :disabled="isTaskRunning"
            @click="rescan">重扫所选文件库</el-button>
        </div>
        <div class="record-help">立即执行会按当前筛选文件库启动一次后台分析。重置失败会把失败记录改回待处理；恢复中断只用于服务异常关闭后，把残留的处理中记录放回队列；重扫会清空所选文件库的分析记录。</div>
        <el-table :data="records" class="record-table">
          <el-table-column label="操作" width="90" fixed="left">
            <template #default="{ row }">
              <div class="record-actions">
                <el-button size="small" type="success" link @click="playResource(row)">播放</el-button>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="filesBasesName" label="文件库" width="120" show-overflow-tooltip />
          <el-table-column prop="resourceName" label="资源名称" min-width="240" show-overflow-tooltip>
            <template #default="{ row }">
              {{ row.resourceName || row.resourcesId }}
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="110" />
          <el-table-column prop="writtenTagText" label="写入标签" min-width="220" show-overflow-tooltip>
            <template #default="{ row }">
              {{ row.writtenTagText || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="failReason" label="失败原因" min-width="300" show-overflow-tooltip />
          <el-table-column prop="analyzedAt" label="分析时间" width="180" />
        </el-table>
        <div class="pagination">
          <el-pagination layout="prev, pager, next, total" :total="recordTotal" :page-size="recordLimit"
            v-model:current-page="recordPage" @current-change="loadRecords" />
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage, ElMessageBox } from 'element-plus';
import { aiTagServer } from '@/server/aiTag.server';
import { tagServer } from '@/server/tag.server';
import { filesBasesStoreData } from '@/storeData/filesBases.storeData';
import type {
  I_aiTagFilesBasesSetting,
  I_aiTagModelTestResult,
  I_aiTagRecord,
  I_aiTagSetting,
  I_aiTagSettingRecommendation,
  I_aiTagStats,
} from '@/dataType/aiTag.dataType';
import type { I_tagClass } from '@/dataType/tag.dataType';

const store = {
  filesBasesStoreData: filesBasesStoreData(),
}
const router = useRouter();

const defaultSetting = (): I_aiTagSetting => ({
  id: '',
  enabled: false,
  paused: false,
  provider: 'openai',
  baseUrl: '',
  apiKey: '',
  model: '',
  requestTimeoutSeconds: 1800,
  maxResourcesPerRun: 20,
  maxFramesPerResource: 80,
  maxFramesPerVideo: 40,
  maxVideosPerResource: 12,
  maxImagesPerAiRequest: 5,
  frameStrategy: 'high_accuracy_adaptive',
  imageResizeMode: 'original',
  fallbackImageMaxWidth: 1440,
  imageJpegQuality: 100,
  minConfidence: 0.7,
  maxTagsPerResource: 8,
  writeMode: 'append',
});

const activeTab = ref('setting');
const loading = ref(false);
const isExecuting = ref(false);
const modelTestDialogVisible = ref(false);
const modelTestLoading = ref(false);
const modelTestResult = ref<I_aiTagModelTestResult | null>(null);
const setting = ref<I_aiTagSetting>(defaultSetting());
const filesBasesSettings = ref<I_aiTagFilesBasesSetting[]>([]);
const selectedFilesBasesIds = ref<string[]>([]);
const stats = ref<I_aiTagStats>({ pending: 0, processing: 0, success: 0, failed: 0, skipped: 0 });
const records = ref<I_aiTagRecord[]>([]);
const recordTotal = ref(0);
const recordPage = ref(1);
const recordLimit = 20;
const recordStatus = ref('');
const tagClassOptions = ref<Record<string, I_tagClass[]>>({});
const missingDescriptionMap = ref<Record<string, number>>({});
let refreshTimer: ReturnType<typeof setInterval> | null = null;

// 只统计已启用文件库中“参与 AI 但缺少说明”的标签，帮助用户优先补充简写标签解释。
const missingDescriptionCount = computed(() => {
  return filesBasesSettings.value
    .filter(item => item.enabled)
    .reduce((total, item) => total + (missingDescriptionMap.value[item.filesBasesId] || 0), 0);
});

// 任务状态接口支持按单个文件库过滤；多选或全选时传空，表示查看全部启用范围。
const selectedQueryFilesBasesId = computed(() => {
  return selectedFilesBasesIds.value.length === 1 ? selectedFilesBasesIds.value[0] : '';
});

// 前端运行状态由“刚点击立即执行”和“后端仍有 processing 记录”共同决定。
// 这样后台 worker 已启动但 HTTP 请求已经返回时，按钮仍能保持正确禁用。
const isTaskRunning = computed(() => {
  return isExecuting.value || stats.value.processing > 0;
});

// 后端可能返回空 include/exclude，统一成数组，避免 el-select multiple 收到 undefined。
const normalizeFilesBasesSettings = () => {
  filesBasesSettings.value = filesBasesSettings.value.map(item => ({
    filesBasesId: item.filesBasesId,
    enabled: item.enabled,
    includeTagClassIds: item.includeTagClassIds || [],
    excludeTagClassIds: item.excludeTagClassIds || [],
  }));
}

// 顶部多选和文件库表格里的开关是同一份启用状态。
// 这两个同步函数分别处理“表格改动”和“顶部多选改动”的方向。
const syncSelectedFromRows = () => {
  selectedFilesBasesIds.value = filesBasesSettings.value.filter(item => item.enabled).map(item => item.filesBasesId);
}

const syncRowsFromSelected = () => {
  const selected = new Set(selectedFilesBasesIds.value);
  filesBasesSettings.value.forEach(item => {
    item.enabled = selected.has(item.filesBasesId);
  });
}

// 读取每个文件库的标签分类，用于 include/exclude 下拉；
// 同时统计缺少 AI 说明的标签数量，给设置页提示。
const loadTagOptions = async () => {
  const classMap: Record<string, I_tagClass[]> = {};
  const missingMap: Record<string, number> = {};
  for (const item of filesBasesSettings.value) {
    const result = await tagServer.tagDataByFilesBasesId(item.filesBasesId);
    if (result.status) {
      classMap[item.filesBasesId] = result.data.tagClass.filter(tagClass => tagClass.status);
      missingMap[item.filesBasesId] = result.data.tag.filter(tag => tag.status && tag.aiEnabled !== false && !tag.aiDescription).length;
    }
  }
  tagClassOptions.value = classMap;
  missingDescriptionMap.value = missingMap;
}

// 初始化设置页：全局设置和文件库启用配置并行加载，标签分类依赖文件库列表加载后再请求。
const init = async () => {
  loading.value = true;
  try {
    const [settingResult, filesBasesResult] = await Promise.all([
      aiTagServer.setting(),
      aiTagServer.filesBases(),
    ]);
    if (settingResult.status) {
      setting.value = { ...defaultSetting(), ...settingResult.data };
    } else {
      ElMessage.error(settingResult.msg);
    }
    if (filesBasesResult.status) {
      filesBasesSettings.value = filesBasesResult.data;
      normalizeFilesBasesSettings();
      syncSelectedFromRows();
      await loadTagOptions();
    } else {
      ElMessage.error(filesBasesResult.msg);
    }
    await reloadStatus();
  } finally {
    loading.value = false;
  }
}

// 保存设置会同时保存全局配置和文件库启用范围。
// showSuccess=false 用于“立即执行/测试服务”等内部保存，避免重复弹成功提示。
const saveAll = async (showSuccess = true): Promise<boolean> => {
  loading.value = true;
  try {
    syncRowsFromSelected();
    const settingResult = await aiTagServer.saveSetting(setting.value);
    if (!settingResult.status) {
      ElMessage.error(settingResult.msg);
      return false;
    }
    setting.value = { ...defaultSetting(), ...settingResult.data };
    const filesBasesResult = await aiTagServer.saveFilesBases(filesBasesSettings.value);
    if (!filesBasesResult.status) {
      ElMessage.error(filesBasesResult.msg);
      return false;
    }
    if (showSuccess) {
      ElMessage.success('保存成功');
    }
    return true;
  } finally {
    loading.value = false;
  }
}

// 构造一个前端兜底的失败结果。
// 只有 API 层失败时使用；后端模型测试失败会返回结构化报告而不是抛接口错误。
const emptyModelTestResult = (message: string): I_aiTagModelTestResult => ({
  success: false,
  model: setting.value.model,
  endpoint: setting.value.baseUrl,
  responseFormat: '',
  fallbackUsed: false,
  finishReason: '',
  summary: '',
  content: '',
  error: message,
  firstError: '',
  metrics: {
    promptTokens: 0,
    completionTokens: 0,
    totalTokens: 0,
    usageReturned: false,
    elapsedMs: 0,
    estimatedTokensPerSecond: 0,
    serviceTokensPerSecond: 0,
    servicePromptPerSecond: 0,
    serviceGeneratedPerSecond: 0,
  },
  recommendations: [],
});

const formatElapsed = (elapsedMs: number) => {
  if (!elapsedMs) return '-';
  if (elapsedMs < 1000) return `${elapsedMs} ms`;
  return `${(elapsedMs / 1000).toFixed(2)} 秒`;
}

const metricText = (value: number, usageReturned: boolean) => {
  if (!usageReturned || !value) return '-';
  return String(value);
}

const speedText = (result: I_aiTagModelTestResult) => {
  const metrics = result.metrics;
  const speed = metrics.serviceGeneratedPerSecond || metrics.serviceTokensPerSecond || metrics.estimatedTokensPerSecond;
  if (!speed) return '-';
  return `${speed.toFixed(2)} tokens/s`;
}

const formatRecommendationValue = (value: string | number | boolean) => {
  if (value === 'original') return '原始分辨率';
  if (value === 'auto_fallback') return '自动降级';
  if (value === 'fixed_resize') return '固定缩放';
  if (value === true) return '开启';
  if (value === false) return '关闭';
  return String(value);
}

const applyRecommendation = (item: I_aiTagSettingRecommendation) => {
  (setting.value as Record<string, string | number | boolean>)[item.field] = item.recommendedValue;
  ElMessage.success(`${item.label} 已应用到表单，保存后生效`);
}

// 推荐参数只应用到当前表单，不自动保存。
// 这样用户可以逐项检查，最后统一点击“保存设置”落库。
const applyAllRecommendations = () => {
  if (!modelTestResult.value?.recommendations?.length) return;
  modelTestResult.value.recommendations.forEach(item => {
    (setting.value as Record<string, string | number | boolean>)[item.field] = item.recommendedValue;
  });
  ElMessage.success('推荐参数已全部应用到表单，保存后生效');
}

// 测试模型输出使用当前表单值直接请求后端，不会先保存配置。
// 如果表单里是脱敏 API Key，后端会自动沿用已保存的真实 key。
const testConnection = async () => {
  modelTestDialogVisible.value = true;
  modelTestLoading.value = true;
  modelTestResult.value = null;
  try {
    const result = await aiTagServer.testConnection(setting.value);
    if (result.status) {
      modelTestResult.value = result.data;
    } else {
      modelTestResult.value = emptyModelTestResult(result.msg);
    }
  } finally {
    modelTestLoading.value = false;
  }
}

// 测试服务仍复用已保存配置，主要检查 /models 是否可访问和模型名是否存在。
// 它不代表模型一定能按 AI 自动标签要求返回 JSON，真实推理能力看“测试模型输出”。
const testService = async () => {
  const saved = await saveAll(false);
  if (!saved) return;
  const result = await aiTagServer.testService();
  if (result.status) {
    ElMessage.success('服务可用，模型名匹配');
  } else {
    ElMessage.error(result.msg);
  }
}

// 启动后台分析时只发起一次 runOnce 请求。
// 后端 worker 会在后台持续领取资源，因此这里请求返回后仍需要自动刷新任务状态。
const runSelectedFilesBases = async () => {
  if (selectedFilesBasesIds.value.length === 0) {
    ElMessage.warning('请先选择要启用并执行的文件库');
    return;
  }
  isExecuting.value = true;
  startAutoRefresh();
  try {
    const result = await aiTagServer.runOnce('');
    if (!result.status) {
      ElMessage.error(result.msg);
      return;
    }
    if (result.data.running) {
      ElMessage.info('AI 自动标签任务已在后台运行');
    } else if (result.data.started) {
      ElMessage.success('已启动后台分析');
    } else {
      ElMessage.warning('未启动任务，请确认已启用 AI 自动标签并选择文件库');
    }
  } finally {
    isExecuting.value = false;
    await reloadStatus();
  }
}

// 立即执行前先保存当前设置，保证计划范围、截图策略和写入策略与用户看到的一致。
const runOnce = async () => {
  if (isTaskRunning.value) return;
  const saved = await saveAll(false);
  if (!saved) return;
  if (setting.value.paused) {
    ElMessage.warning('当前处于暂停状态，请先点击继续执行');
    return;
  }
  await runSelectedFilesBases();
}

// 暂停只阻止后端继续领取下一个资源，当前资源会继续跑完。
// 因此提示用户“当前步骤结束后停止”，避免误以为会立刻取消正在进行的 AI 请求。
const pauseTask = async () => {
  const result = await aiTagServer.pause();
  if (result.status) {
    setting.value = { ...defaultSetting(), ...result.data };
    ElMessage.success('已暂停。正在处理的资源会在当前步骤结束后停止，后续任务不会继续取新资源。');
    await reloadStatus();
  } else {
    ElMessage.error(result.msg);
  }
}

// 继续执行先清除 paused，再重新触发 runOnce。
// 如果旧 worker 还没退出，后端会标记 restartAfterStop，避免并发 worker。
const resumeTask = async () => {
  const result = await aiTagServer.resume();
  if (result.status) {
    setting.value = { ...defaultSetting(), ...result.data };
    await reloadStatus();
    ElMessage.success('已继续，正在启动后台分析。');
    await runSelectedFilesBases();
  } else {
    ElMessage.error(result.msg);
  }
}

// 状态刷新同时更新统计卡片和记录列表。
const reloadStatus = async () => {
  const [statsResult] = await Promise.all([
    aiTagServer.stats(selectedQueryFilesBasesId.value),
    loadRecords(),
  ]);
  if (statsResult.status) {
    stats.value = statsResult.data;
  }
}

const loadRecords = async () => {
  const result = await aiTagServer.records(selectedQueryFilesBasesId.value, recordStatus.value, recordPage.value, recordLimit);
  if (result.status) {
    records.value = result.data.dataList || [];
    recordTotal.value = result.data.total || 0;
  } else {
    ElMessage.error(result.msg);
  }
}

// 自动刷新只保留一个定时器。
// 任务运行中即使用户不在“任务状态”Tab，也会刷新 processing 状态，保证按钮及时恢复。
const startAutoRefresh = () => {
  if (refreshTimer) return;
  refreshTimer = setInterval(() => {
    reloadStatus();
  }, 3000);
}

const stopAutoRefresh = () => {
  if (!refreshTimer) return;
  clearInterval(refreshTimer);
  refreshTimer = null;
}

// 进入任务状态页或任务运行时开启自动刷新，离开且任务停止后关闭，避免常驻轮询。
const syncAutoRefresh = () => {
  if (activeTab.value === 'records' || isTaskRunning.value) {
    startAutoRefresh();
  } else {
    stopAutoRefresh();
  }
}

// 失败重置和中断恢复都要求后端没有运行中的 worker。
// 前端先挡一次，后端也会再次校验，防止并发状态被改乱。
const resetFailed = async () => {
  if (isTaskRunning.value) {
    ElMessage.warning('当前仍有任务运行，请暂停并等待当前资源结束后再重置失败');
    return;
  }
  const targets = selectedFilesBasesIds.value;
  let reset = 0;
  for (const filesBasesId of targets) {
    const result = await aiTagServer.resetFailed(filesBasesId);
    if (!result.status) {
      ElMessage.error(result.msg);
      return;
    }
    reset += result.data.reset || 0;
  }
  ElMessage.success(`已重置 ${reset} 条失败记录`);
  await reloadStatus();
}

const resetProcessing = async () => {
  if (isTaskRunning.value) {
    ElMessage.warning('当前仍有任务运行，请暂停并等待当前资源结束后再恢复中断记录');
    return;
  }
  const targets = selectedFilesBasesIds.value;
  let reset = 0;
  for (const filesBasesId of targets) {
    const result = await aiTagServer.resetProcessing(filesBasesId);
    if (!result.status) {
      ElMessage.error(result.msg);
      return;
    }
    reset += result.data.reset || 0;
  }
  ElMessage.success(`已恢复 ${reset} 条中断记录`);
  await reloadStatus();
}

// 重扫会删除所选文件库的全部 AI 分析记录。
// 这是“强制重新分析”的入口，所以必须确认且运行中禁止执行。
const rescan = async () => {
  if (isTaskRunning.value) {
    ElMessage.warning('当前仍有任务运行，请暂停并等待当前资源结束后再重扫');
    return;
  }
  if (selectedFilesBasesIds.value.length === 0) {
    ElMessage.warning('请先选择要重扫的文件库');
    return;
  }
  await ElMessageBox.confirm('重扫会清空已选文件库的 AI 分析记录，后续计划任务会重新分析。', '确认重扫', { type: 'warning' });
  for (const filesBasesId of selectedFilesBasesIds.value) {
    const result = await aiTagServer.rescan(filesBasesId);
    if (!result.status) {
      ElMessage.error(result.msg);
      return;
    }
  }
  ElMessage.success('已进入重扫状态');
  await reloadStatus();
}

// AI 记录只提供播放入口，不再做复杂定位。
// 播放页能直接让用户核对资源内容和 AI 写入结果。
const playResource = (record: I_aiTagRecord) => {
  if (!record.resourcesId) {
    ElMessage.warning('缺少资源播放信息');
    return;
  }
  router.push({
    name: 'playMovies',
    params: {
      resourceId: record.resourcesId,
    },
  });
}

// 设置页挂载后加载配置；卸载时关闭轮询，避免离开页面后仍请求状态接口。
onMounted(() => {
  init();
});

// Tab 或任务运行状态变化时同步自动刷新策略。
watch([activeTab, isTaskRunning], () => {
  syncAutoRefresh();
});

onUnmounted(() => {
  stopAutoRefresh();
});
</script>

<style lang="scss" scoped>
.ai-tag-setting {
  height: 100%;
  overflow: auto;
  padding: 12px;

  .feature-intro {
    border-left: 3px solid var(--el-color-primary);
    background: var(--el-color-primary-light-9);
    padding: 10px 12px;
    margin-bottom: 12px;

    &__title {
      color: var(--el-text-color-primary);
      font-size: 14px;
      font-weight: 600;
      line-height: 1.4;
      margin-bottom: 4px;
    }

    &__text {
      color: var(--el-text-color-regular);
      font-size: 12px;
      line-height: 1.7;
    }
  }

  .toolbar {
    display: flex;
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
    margin-bottom: 12px;

    &__settings,
    &__actions {
      display: flex;
      align-items: center;
      flex-wrap: wrap;
    }

    &__settings {
      gap: 8px;
    }

    &__actions {
      gap: 4px;
      justify-content: flex-end;
      width: 100%;

      :deep(.el-button + .el-button) {
        margin-left: 0;
      }
    }

    .files-bases-select {
      width: 220px;
    }
  }

  .setting-form {
    max-width: 980px;

    :deep(.el-form-item__content) {
      align-items: flex-start;
      flex-direction: column;
    }
  }

  .field-help,
  .table-help,
  .record-help,
  .alert-help {
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 1.6;
  }

  .field-help {
    margin-top: 6px;
    max-width: 760px;
  }

  .table-help,
  .alert-help {
    margin-top: 4px;
  }

  .record-help {
    margin-top: 8px;
    margin-bottom: 8px;
  }

  .number-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: 4px;

    :deep(.el-form-item) {
      margin-bottom: 14px;
    }

    :deep(.el-form-item__content) {
      flex-direction: row;
      align-items: center;
      gap: 16px;
    }

    .field-help {
      flex: 1;
      min-width: 360px;
      margin-top: 0;
      max-width: none;
    }
  }

  .files-bases-table,
  .record-table {
    margin-top: 12px;
  }

  .stats-row {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
    gap: 10px;
    margin-bottom: 12px;

    .stat-item {
      border: 1px solid var(--el-border-color);
      background: var(--el-bg-color);
      border-radius: 6px;
      padding: 10px 12px;
      display: flex;
      justify-content: space-between;
      align-items: center;

      span {
        color: var(--el-text-color-secondary);
        font-size: 13px;
      }

      strong {
        color: var(--el-text-color-primary);
        font-size: 20px;
      }
    }
  }

  .record-tools {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;

    .status-select {
      width: 160px;
    }
  }

  .record-actions {
    display: flex;
    align-items: center;
    gap: 8px;

    :deep(.el-button + .el-button) {
      margin-left: 0;
    }
  }

  .pagination {
    margin-top: 12px;
    display: flex;
    justify-content: flex-end;
  }

  .model-test-loading {
    min-height: 120px;
    color: var(--el-text-color-regular);
    font-size: 13px;
    line-height: 1.8;
    display: flex;
    align-items: center;
  }

  .model-test-metrics {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
    gap: 8px;
    margin-top: 12px;

    .metric-item {
      border: 1px solid var(--el-border-color);
      background: var(--el-bg-color);
      border-radius: 6px;
      padding: 8px 10px;
      min-height: 58px;
      display: flex;
      flex-direction: column;
      justify-content: center;
      gap: 4px;

      span {
        color: var(--el-text-color-secondary);
        font-size: 12px;
      }

      strong {
        color: var(--el-text-color-primary);
        font-size: 16px;
        font-weight: 600;
      }
    }
  }

  .model-test-detail {
    color: var(--el-text-color-regular);
    font-size: 12px;
    line-height: 1.8;
    margin-top: 12px;
  }

  .model-output {
    margin-top: 12px;

    &__title {
      color: var(--el-text-color-primary);
      font-size: 13px;
      font-weight: 600;
      margin-bottom: 6px;
    }

    pre {
      background: var(--el-fill-color-light);
      border: 1px solid var(--el-border-color-lighter);
      border-radius: 6px;
      color: var(--el-text-color-primary);
      font-size: 12px;
      line-height: 1.6;
      margin: 0;
      max-height: 160px;
      overflow: auto;
      padding: 8px 10px;
      white-space: pre-wrap;
      word-break: break-word;
    }
  }

  .recommendation-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    color: var(--el-text-color-primary);
    font-size: 13px;
    font-weight: 600;
    margin-top: 14px;
    margin-bottom: 8px;
  }

  .recommendation-table {
    margin-top: 4px;

    :deep(.el-table__cell) {
      padding: 7px 0;
    }
  }

  .recommendation-text {
    color: var(--el-text-color-regular);
    font-size: 12px;
    line-height: 1.5;
    white-space: normal;
    word-break: break-word;
  }
}

:deep(.el-dialog) {
  max-width: calc(100vw - 48px);
}

@media (max-width: 900px) {
  .ai-tag-setting {
    .number-grid {
      :deep(.el-form-item__content) {
        flex-direction: column;
        align-items: flex-start;
        gap: 6px;
      }

      .field-help {
        min-width: 0;
      }
    }
  }
}
</style>
