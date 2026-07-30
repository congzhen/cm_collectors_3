<template>
  <div class="resourceDramaSeries-list">
    <div class="resourceDramaSeries-list-index" v-if="props.showMode === E_detailsDramaSeriesMode.digit">
      <ul>
        <el-popover v-for="(item, key) in props.dramaSeries" :key="item.id || key" trigger="hover" :width="380"
          :show-after="1000" :disabled="metadataDetails(item).length === 0">
          <template #reference>
            <li :class="[selectedClass(item.id)]" @contextmenu.prevent.stop="addToTranscode(item)"
              @click="emits('playResourceDramaSeries', item)">
              {{ (key + 1) }}
            </li>
          </template>
          <div class="metadata-detail">
            <div v-for="detail in metadataDetails(item)" :key="detail.label">
              <span>{{ detail.label }}</span>
              <strong>{{ detail.value }}</strong>
            </div>
          </div>
        </el-popover>
      </ul>
    </div>
    <div class="resourceDramaSeries-list-name" v-else>
      <ul>
        <li :class="[selectedClass(item.id)]" v-for="(item, key) in props.dramaSeries" :key="key"
          @contextmenu.prevent.stop="addToTranscode(item)"
          @click="emits('playResourceDramaSeries', item)">
          <label>{{ (key + 1) }}.</label>
          <div class="series-content">
            <span class="file-name">{{ getFinalPathSegment(item.src) }}</span>
            <div v-if="metadataText(item) || metadataStatusText(item) ||
              item.videoMetadata?.probeStatus === 'failed'" class="metadata-line">
              <el-popover v-if="metadataText(item)" trigger="hover" :width="380">
                <template #reference>
                  <el-icon class="metadata-detail-trigger" title="查看视频详情">
                    <InfoFilled />
                  </el-icon>
                </template>
                <div class="metadata-detail">
                  <div v-for="detail in metadataDetails(item)" :key="detail.label">
                    <span>{{ detail.label }}</span>
                    <strong>{{ detail.value }}</strong>
                  </div>
                </div>
              </el-popover>
              <span v-if="metadataText(item)" class="metadata">{{ metadataText(item) }}</span>
              <span v-if="metadataStatusText(item)" class="metadata metadata-status"
                :class="`metadata-status--${item.videoMetadata?.probeStatus}`">
                {{ metadataStatusText(item) }}
              </span>
              <el-tooltip v-else-if="item.videoMetadata?.probeStatus === 'failed'"
                :content="item.videoMetadata.errorMessage || '视频信息采集失败'">
                <span class="metadata metadata-error">采集失败</span>
              </el-tooltip>
            </div>
          </div>
        </li>
      </ul>
    </div>
  </div>
</template>
<script lang="ts" setup>
import type { PropType } from 'vue'
import { E_detailsDramaSeriesMode } from '@/dataType/app.dataType'
import type { I_resourceDramaSeries } from '@/dataType/resource.dataType';
import { getFinalPathSegment } from '@/assets/tool'
import { appStoreData } from '@/storeData/app.storeData'
import { videoTranscodeServer } from '@/server/videoTranscode.server'
import { ElMessage, ElMessageBox } from 'element-plus'
const props = defineProps({
  showMode: {
    type: String as PropType<(typeof E_detailsDramaSeriesMode)[keyof typeof E_detailsDramaSeriesMode]>,
    default: E_detailsDramaSeriesMode.fileName,
  },
  dramaSeries: {
    type: Array as PropType<I_resourceDramaSeries[]>,
    required: true,
  },
  //当前选中
  selectedId: {
    type: String,
    default: '',
  },
})

const emits = defineEmits(['playResourceDramaSeries']);
const store = { appStoreData: appStoreData() };

const addToTranscode = async (item: I_resourceDramaSeries) => {
  if (!store.appStoreData.displayAdminFn) return;
  try {
    await ElMessageBox.confirm(
      `将“${getFinalPathSegment(item.src)}”加入视频转码列表？`,
      '加入视频转码列表',
      { type: 'warning' },
    );
    const result = await videoTranscodeServer.add({ dramaSeriesIds: [item.id] });
    if (result.status && result.data.added > 0) {
      ElMessage.success('已加入视频转码列表');
    } else if (result.status && result.data.skippedDuplicate > 0) {
      ElMessage.info('该视频已在转码列表中');
    } else {
      ElMessage.error(result.msg || '视频文件不存在或无法加入');
    }
  } catch {
    // 用户取消确认。
  }
};

const selectedClass = (id: string) => {
  if (props.selectedId != '' && id === props.selectedId) {
    return 'selected'
  }
  return '';
}

const metadataText = (item: I_resourceDramaSeries) => {
  const metadata = item.videoMetadata;
  if (!metadata) return '';
  const parts: string[] = [];
  if (metadata.width > 0 && metadata.height > 0) parts.push(`${metadata.width}×${metadata.height}`);
  if (metadata.videoCodec) parts.push(metadata.videoCodec.toUpperCase());
  if (metadata.frameRate > 0) parts.push(`${metadata.frameRate}fps`);
  if (metadata.audioCodec) parts.push(metadata.audioCodec.toUpperCase());
  if (metadata.fileSize > 0) parts.push(formatFileSize(metadata.fileSize));
  return parts.join(' · ');
};

const metadataStatusText = (item: I_resourceDramaSeries) => {
  const status = item.videoMetadata?.probeStatus;
  if (status === 'stale') return '待更新';
  if (status === 'processing') return '采集中';
  if (status === 'failed' && metadataText(item)) return '采集失败';
  return '';
};

const metadataDetails = (item: I_resourceDramaSeries) => {
  const metadata = item.videoMetadata;
  if (!metadata) return [];
  const result: Array<{ label: string; value: string }> = [];
  const push = (label: string, value: string | number) => {
    if (value !== '' && value !== 0) result.push({ label, value: String(value) });
  };
  if (metadata.width > 0 && metadata.height > 0) push('分辨率', `${metadata.width} × ${metadata.height}`);
  if (metadata.frameRate > 0) {
    push('帧率', `${metadata.frameRate} fps${metadata.frameRateRaw ? `（${metadata.frameRateRaw}）` : ''}`);
  }
  push('视频编码', metadata.videoCodec.toUpperCase());
  push('编码配置', metadata.videoProfile);
  push('像素格式', metadata.pixelFormat);
  if (metadata.bitDepth > 0) push('位深', `${metadata.bitDepth} bit`);
  if (metadata.videoBitRate > 0) push('视频码率', formatBitRate(metadata.videoBitRate));
  push('封装格式', metadata.containerFormat);
  push('音频编码', metadata.audioCodec.toUpperCase());
  if (metadata.audioChannels > 0) push('声道数', metadata.audioChannels);
  if (metadata.audioSampleRate > 0) push('音频采样率', `${metadata.audioSampleRate} Hz`);
  if (metadata.fileSize > 0) push('文件大小', formatFileSize(metadata.fileSize));
  return result;
};

const formatFileSize = (size: number) => {
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let value = size;
  let unitIndex = 0;
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex++;
  }
  return `${value.toFixed(unitIndex >= 3 ? 2 : 1)} ${units[unitIndex]}`;
};

const formatBitRate = (bitRate: number) => {
  if (bitRate >= 1_000_000) return `${(bitRate / 1_000_000).toFixed(2)} Mbps`;
  if (bitRate >= 1_000) return `${(bitRate / 1_000).toFixed(0)} Kbps`;
  return `${bitRate} bps`;
};

</script>
<style lang="scss" scoped>
.resourceDramaSeries-list {
  width: 100%;
  min-width: 0;
  max-width: 100%;
  overflow: hidden;
  padding-bottom: 0.5em;
}

.selected {
  background-color: #E6A23C;
  border-radius: 5px;
  color: #303133;
  font-weight: 500;
}

.resourceDramaSeries-list-index {
  padding-top: 0.5em;

  ul {
    width: 100%;
    list-style-type: none;
    display: flex;
    flex-wrap: wrap;
    gap: 5px;

    li {
      width: 17.5%;
      height: 1.2em;
      line-height: 1.2em;
      font-weight: 500;
      text-align: center;
      padding: 0.5em 0;
      border: 1px solid #a8abb2;
      border-radius: 4px;
      cursor: pointer;
      user-select: none;
      /* 新增过渡动画 */
      transition:
        background-color 0.2s ease,
        box-shadow 0.2s ease;

      /* 鼠标悬停时的高亮效果 */
      &:hover {
        color: var(--el-color-primary);
        background-color: var(--el-color-primary-light-9);
        border-color: var(--el-color-primary);
      }
    }
  }
}

.resourceDramaSeries-list-name {
  min-width: 0;

  ul {
    width: calc(100% - 0.4em);
    min-width: 0;
    box-sizing: border-box;
    margin: 0 0.2em;
    padding: 0;
    list-style-type: none;

    li {
      width: calc(100% - 2em);
      min-width: 0;
      box-sizing: content-box;
      line-height: 1.2em;
      font-weight: 500;
      font-style: italic;
      padding: 0.5em 1em;
      border-bottom: 1px dotted rgba(168, 171, 178, 0.5);
      cursor: pointer;
      user-select: none;
      display: flex;
      align-items: flex-start;
      /* 新增过渡动画 */
      transition:
        background-color 0.2s ease,
        box-shadow 0.2s ease;

      /* 鼠标悬停时的高亮效果 */
      &:hover {
        color: var(--el-color-primary);
        background-color: var(--el-color-primary-light-9);
        border-color: var(--el-color-primary);
      }

      label {
        flex-shrink: 0;
        padding-right: 0.8em;
      }

      .series-content {
        flex-grow: 1;
        min-width: 0;
      }

      .file-name {
        display: block;
        width: 100%;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      .metadata-line {
        display: flex;
        align-items: center;
        gap: 0.35em;
        min-width: 0;
        margin-top: 0.3em;
      }

      .metadata-detail-trigger {
        flex-shrink: 0;
        color: var(--el-text-color-secondary);
        cursor: help;
        font-size: 1em;
      }

      .metadata {
        min-width: 0;
        color: var(--el-text-color-secondary);
        font-size: 0.85em;
        font-style: normal;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      .metadata-error {
        color: var(--el-color-danger);
      }

      .metadata-status {
        flex-shrink: 0;
      }

      .metadata-status--stale {
        color: var(--el-color-warning);
      }

      .metadata-status--processing {
        color: var(--el-color-primary);
      }

      .metadata-status--failed {
        color: var(--el-color-danger);
      }
    }
  }
}

.metadata-detail {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px 18px;

  > div {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;

    span {
      color: var(--el-text-color-secondary);
      font-size: 12px;
    }

    strong {
      overflow-wrap: anywhere;
      font-weight: 500;
    }
  }
}
</style>
