<template>
  <div class="layout-cover-poster-waterfall">
    <div class="main">
      <el-scrollbar>
        <Waterfall ref="waterfallRef" :list="dataList_C" :gutter="10" :breakpoints="waterfallBreakpoints"
          :img-selector="'src'" class="waterfall-list">
          <template #default="{ item }">
            <div class="waterfall-item" @click="selectResourcesHandle(item)">
              <el-image :src="item.src" :title="item.title" @load="onImageLoad" />
              <div class="play-icon" @click.stop="playResource(item)">
                <el-icon>
                  <VideoPlay />
                </el-icon>
              </div>
            </div>
          </template>
        </Waterfall>
      </el-scrollbar>
    </div>
    <div class="tool">
      <el-slider v-model="waterfallColumn" :min="1" :max="20" style="width: 200px;" />
    </div>
  </div>
</template>
<script lang="ts" setup>
import { Waterfall } from 'vue-waterfall-plugin-next'
import 'vue-waterfall-plugin-next/dist/style.css'
import type { I_resource } from '@/dataType/resource.dataType';
import { computed, nextTick, ref, type PropType } from 'vue';
import { getResourceCoverPoster } from '@/common/photo';
import { playResource } from '@/common/play';
import { appStoreData } from '@/storeData/app.storeData';
import { debounceNow } from '@/assets/debounce';
const store = {
  appStoreData: appStoreData(),
}
const props = defineProps({
  dataList: {
    type: Array as PropType<I_resource[]>,
    default: () => [],
  },
})
const emits = defineEmits(['selectResources']);

const waterfallRef = ref<InstanceType<typeof Waterfall>>();
const waterfallColumn = ref(store.appStoreData.currentConfigApp.coverPosterWaterfallColumn);

const dataList_C = computed(() => {
  return props.dataList.map((item: I_resource) => {
    const src = getResourceCoverPoster(item);
    return {
      ...item,
      src,
    }
  })
});

// 计算动态 breakpoints
const waterfallBreakpoints = computed(() => {
  // 可以根据当前列数设置不同的断点
  return {
    9999: { rowPerView: waterfallColumn.value },
  }
});
const selectResourcesHandle = (item: I_resource) => {
  emits('selectResources', item)
}

// 图片加载完成事件处理
const onImageLoad = debounceNow(() => {
  nextTick(() => {
    waterfallRef.value?.renderer();
  });
}, 300);

</script>
<style lang="scss" scoped>
.layout-cover-poster-waterfall {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;

  .main {
    flex: 1;
    overflow: hidden;

    .waterfall-list {
      background-color: unset;
    }

    .waterfall-item {
      position: relative;

      &:hover {
        .play-icon {
          display: block;
        }
      }

      .el-image:hover {
        transform: scale(1.01);
      }

      .play-icon {
        position: absolute;
        z-index: 10;
        margin-left: 0.05em;
        margin-top: -1.5em;
        font-size: 3.8em;
        color: #f3f3f3;
        opacity: 0.75;
        display: none;
        cursor: pointer;
      }
    }



  }

  .tool {
    flex-shrink: 0;
    height: 32px;
    display: flex;
    justify-content: center;
    align-items: center;
  }
}
</style>
