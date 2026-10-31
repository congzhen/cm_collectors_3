<template>
  <div v-if="durationText || groupCount" class="video-duration-badge" :class="{ 'is-mobile-adaptive': props.adaptiveMobile }"
    :style="{ right: props.offsetRight, bottom: props.offsetBottom }">
    <span v-if="groupCount" class="group-count" :aria-label="`关联组共 ${groupCount} 个资源`"><Collection aria-hidden="true" />{{ groupCount }} 项</span>
    <span v-if="groupCount && durationText" class="badge-divider" aria-hidden="true">·</span>
    <span v-if="durationText" class="duration-text">{{ durationText }}</span>
  </div>
</template>
<script setup lang="ts">
import { computed, inject, type PropType } from 'vue';
import { Collection } from '@element-plus/icons-vue';
import type { I_resource } from '@/dataType/resource.dataType';
import { appStoreData } from '@/storeData/app.storeData';
import { getResourceDurationText } from '@/common/videoDuration';
import { resourceGroupContextKey } from './resourceGroups';

const store = {
  appStoreData: appStoreData(),
}
const groupContext = inject(resourceGroupContextKey, undefined);

const props = defineProps({
  resource: {
    type: Object as PropType<I_resource>,
    required: true,
  },
  // 不同封面布局的底部内容高度不一样：
  // 普通海报样式底部有标题遮罩，需要把角标上移；其它样式默认贴近封面右下角即可。
  offsetBottom: {
    type: String,
    default: '4px',
  },
  // 默认为 6px，与大多数封面布局的边缘留白一致；个别布局可以按需要单独覆盖。
  offsetRight: {
    type: String,
    default: '4px',
  },
  // compactText 用于移动端短视频等窄封面场景：只显示总时长，避免“几个视频 · 总时长”过宽。
  compactText: {
    type: Boolean,
    default: false,
  },
  // adaptiveMobile 用于移动端短视频：结合父级容器查询，根据封面宽度自动缩小角标。
  adaptiveMobile: {
    type: Boolean,
    default: false,
  },
})

const groupCount = computed(() => {
  if (!groupContext?.enabled.value || !groupContext.pageIds.value.has(props.resource.id)) return 0;
  const count = groupContext.summaries.value.get(props.resource.associationGroupId || '')?.count;
  return count && Number.isFinite(count) && count > 1 ? count : 0;
});
const durationText = computed(() => {
  // 时长开关仅隐藏时长，合并组仍保留成员数量。
  if (!store.appStoreData.currentConfigApp.showVideoDuration) {
    return '';
  }
  return getResourceDurationText(props.resource, { compact: props.compactText });
})
</script>
<style lang="scss" scoped>
.video-duration-badge {
  display: flex;
  align-items: center;
  gap: 5px;
  box-sizing: border-box;
  position: absolute;
  right: 6px;
  bottom: 6px;
  z-index: 12;
  max-width: calc(100% - 12px);
  padding: 2px 6px;
  border-radius: 4px;
  background-color: rgba(0, 0, 0, 0.68);
  color: #f3f3f3;
  font-size: 12px;
  line-height: 18px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  pointer-events: none;

  .group-count { display: inline-flex; align-items: center; gap: 4px; flex-shrink: 0; }
  .group-count svg { width: 12px; height: 12px; flex-shrink: 0; }
  .badge-divider { opacity: .6; flex-shrink: 0; }
  .duration-text { min-width: 0; overflow: hidden; text-overflow: ellipsis; }

  &.is-mobile-adaptive {
    padding: 1px 4px;
    border-radius: 3px;
    font-size: 10px;
    line-height: 14px;
    max-width: calc(100% - 8px);
  }
}

// 移动端短视频的封面宽度变化很明显，这里用容器查询让角标跟着封面变小。
// 父级布局只要设置 container-type: inline-size，角标就会按封面宽度自动套用下面的规则。
@container (max-width: 120px) {
  .video-duration-badge.is-mobile-adaptive {
    padding: 1px 3px;
    font-size: 9px;
    line-height: 12px;
    max-width: calc(100% - 6px);
  }
}

@container (max-width: 88px) {
  .video-duration-badge.is-mobile-adaptive {
    padding: 0 2px;
    font-size: 8px;
    line-height: 10px;
    border-radius: 2px;
    max-width: calc(100% - 4px);
  }
}

@container (max-width: 72px) {
  .video-duration-badge.is-mobile-adaptive {
    display: none;
  }
}
</style>
