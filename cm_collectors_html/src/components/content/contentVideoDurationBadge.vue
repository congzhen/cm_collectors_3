<template>
  <div v-if="durationText" class="video-duration-badge" :class="{ 'is-mobile-adaptive': props.adaptiveMobile }"
    :style="{ right: props.offsetRight, bottom: props.offsetBottom }">
    {{ durationText }}
  </div>
</template>
<script setup lang="ts">
import { computed, type PropType } from 'vue';
import type { I_resource } from '@/dataType/resource.dataType';
import { appStoreData } from '@/storeData/app.storeData';
import { getResourceDurationText } from '@/common/videoDuration';

const store = {
  appStoreData: appStoreData(),
}

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

const durationText = computed(() => {
  // 文件库级开关关闭时，角标完全隐藏，也不会展示历史上已采集过的时长。
  if (!store.appStoreData.currentConfigApp.showVideoDuration) {
    return '';
  }
  return getResourceDurationText(props.resource, { compact: props.compactText });
})
</script>
<style lang="scss" scoped>
.video-duration-badge {
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
