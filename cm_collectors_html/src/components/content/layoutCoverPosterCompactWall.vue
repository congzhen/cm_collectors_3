<template>
  <div class="layout-cover-poster-compact-wall">
    <el-scrollbar ref="scrollbarRef">
      <div v-if="props.dataList.length" class="compact-wall" :style="wallStyle">
        <contentRightClickMenu v-for="resource in props.dataList" :key="resource.id" :resource="resource">
          <article class="compact-card" :class="{ selected: selectedResourceId === resource.id }"
            :style="cardStyle(resource)"
            @click="selectResource(resource)">
            <el-image class="compact-cover" :src="getResourceCoverPoster(resource)" fit="cover" />
            <div class="compact-shade"></div>

            <contentTagDisplay class="compact-tags" :resource="resource" font-size-disable />
            <contentVideoDurationBadge :resource="resource" offset-right="5px" offset-bottom="38px"
              compact-text />

            <button class="compact-play" type="button" aria-label="播放"
              @click.stop="playResource(resource)">
              <el-icon><VideoPlay /></el-icon>
            </button>

            <div class="compact-info" :style="{ textAlign: titleAlign }">
              <strong>{{ resource.title }}</strong>
              <div>
                <span v-if="resource.definition">{{ appLang.definition(resource.definition) }}</span>
                <span v-if="resource.issuingDate">{{ displayYear(resource.issuingDate) }}</span>
                <span v-if="resource.dramaSeries.length">{{ resource.dramaSeries.length }} 项</span>
              </div>
            </div>
          </article>
        </contentRightClickMenu>
      </div>
      <el-empty v-else description="没有符合条件的资源" />
      <el-backtop class="custom-backtop"
        target=".layout-cover-poster-compact-wall .el-scrollbar__wrap" :right="20" :bottom="20" />
    </el-scrollbar>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, type PropType } from 'vue'
import type { ElScrollbar } from 'element-plus'
import contentRightClickMenu from './contentRightClickMenu.vue'
import contentTagDisplay from './contentTagDisplay.vue'
import contentVideoDurationBadge from './contentVideoDurationBadge.vue'
import { coverPosterSize, getResourceCoverPoster } from '@/common/photo'
import { playResource } from '@/common/play'
import type { I_resource } from '@/dataType/resource.dataType'
import { AppLang } from '@/language/app.lang'
import { appStoreData } from '@/storeData/app.storeData'

const appLang = AppLang()
const store = appStoreData()
const props = defineProps({
  dataList: {
    type: Array as PropType<I_resource[]>,
    default: () => [],
  },
})
const emits = defineEmits<{
  selectResources: [resource: I_resource]
}>()

const scrollbarRef = ref<InstanceType<typeof ElScrollbar>>()
const selectedResourceId = ref('')

const wallStyle = computed(() => {
  const gap = Math.min(Math.max(store.currentConfigApp.coverPosterGap || 0, 0), 20)
  const padding = Math.min(Math.max(store.currentConfigApp.contentPadding || 0, 0), 20)
  return {
    gap: `${gap}px`,
    justifyContent: store.currentConfigApp.resourceJustifyContent || 'flex-start',
    paddingRight: `${padding}%`,
    paddingLeft: `${padding}%`,
  }
})
const titleAlign = computed(() => store.currentConfigApp.coverTitleAlign || 'left')
const cardStyle = (resource: I_resource) => {
  const size = coverPosterSize(
    resource.coverPosterWidth,
    resource.coverPosterHeight,
    store.currentConfigApp.coverPosterWidthStatus,
    store.currentConfigApp.coverPosterWidthBase,
    store.currentConfigApp.coverPosterHeightStatus,
    store.currentConfigApp.coverPosterHeightBase,
  )
  return {
    width: `${size.width}px`,
    height: `${size.height}px`,
  }
}
const displayYear = (date: string) => date.slice(0, 4)
const selectResource = (resource: I_resource) => {
  selectedResourceId.value = resource.id
  emits('selectResources', resource)
}
const change = () => {
  selectedResourceId.value = ''
  scrollbarRef.value?.setScrollTop(0)
}

defineExpose({ change })
</script>

<style scoped lang="scss">
.layout-cover-poster-compact-wall {
  width: 100%;
  height: 100%;
  overflow: hidden;
}

.compact-wall {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  gap: 5px;
  padding: 4px 4px 12px;

  :deep(.content-right-click-menu) {
    flex: 0 0 auto;
  }
}

.compact-card {
  position: relative;
  flex: none;
  min-width: 0;
  overflow: hidden;
  border: 1px solid rgba(126, 151, 169, .25);
  border-radius: 5px;
  background: #0b1720;
  cursor: pointer;
  box-sizing: border-box;
  transition: border-color .16s ease, box-shadow .16s ease, transform .16s ease;

  &:hover,
  &.selected {
    border-color: var(--el-color-primary);
    box-shadow: 0 0 0 1px color-mix(in srgb, var(--el-color-primary) 55%, transparent);
  }

  &:hover {
    transform: translateY(-1px);

    .compact-cover {
      transform: scale(1.025);
    }

    .compact-play {
      opacity: 1;
      pointer-events: auto;
      transform: translate(-50%, -50%) scale(1);
    }
  }
}

.compact-cover {
  width: 100%;
  height: 100%;
  transition: transform .2s ease;
}

.compact-shade {
  position: absolute;
  inset: 38% 0 0;
  background: linear-gradient(transparent, rgba(3, 9, 13, .9));
  pointer-events: none;
}

.compact-tags {
  position: absolute;
  z-index: 4;
  top: 4px;
  left: 4px;
  max-width: calc(100% - 8px);
  overflow: hidden;
}

.compact-play {
  position: absolute;
  z-index: 6;
  top: 50%;
  left: 50%;
  width: 30px;
  height: 30px;
  display: grid;
  place-items: center;
  border: 1px solid rgba(255, 255, 255, .55);
  border-radius: 50%;
  color: #fff;
  background: rgba(2, 10, 15, .58);
  opacity: 0;
  pointer-events: none;
  cursor: pointer;
  transform: translate(-50%, -50%) scale(.9);
  transition: .16s ease;
}

.compact-info {
  position: absolute;
  z-index: 5;
  right: 0;
  bottom: 0;
  left: 0;
  padding: 18px 6px 5px;
  color: #fff;

  strong {
    display: block;
    overflow: hidden;
    font-size: 11px;
    line-height: 15px;
    text-overflow: ellipsis;
    white-space: nowrap;
    text-shadow: 0 1px 3px #000;
  }

  div {
    display: flex;
    gap: 6px;
    overflow: hidden;
    color: rgba(231, 239, 243, .76);
    font-size: 8px;
    line-height: 12px;
    white-space: nowrap;
    text-shadow: 0 1px 2px #000;
  }
}

</style>
