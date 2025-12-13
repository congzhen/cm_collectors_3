<template>
  <div class="layout-cover-poster-style1">
    <el-scrollbar ref="scrollbarRef">
      <moduleContent v-if="store.appStoreData.currentConfigApp.casualViewModule && !isMobileDevice"
        :resourcesShowMode="props.resourcesShowMode" moduleType="casualView"></moduleContent>
      <moduleContent v-if="store.appStoreData.currentConfigApp.historyModule && !isMobileDevice"
        :resourcesShowMode="props.resourcesShowMode" moduleType="history"></moduleContent>
      <moduleContent v-if="store.appStoreData.currentConfigApp.hotModule && !isMobileDevice"
        :resourcesShowMode="props.resourcesShowMode" moduleType="hot"></moduleContent>
      <h2 v-if="showH2()" class="all-h" style="font-size: 16px;">全部资源</h2>
      <ul class="list-ul"
        :class="{ 'mobile-layout': isMobileDevice, 'mobile-layout-columns-two': isMobileLayoutColumnsTwo() }"
        :style="contentLayoutStyle_C">
        <li v-for="(item, key) in props.dataList" :key="key">
          <contentStyleIndex :resources="item" :resourcesShowMode="props.resourcesShowMode"
            @selectResources="selectResourcesHandle(item)" />
        </li>
      </ul>
      <el-backtop class="custom-backtop" target=".layout-cover-poster-style1 .el-scrollbar__wrap" :right="20"
        :bottom="20" />
    </el-scrollbar>
  </div>
</template>
<script lang="ts" setup>
import moduleContent from './moduleContent.vue';
import contentStyleIndex from './contentStyleIndex.vue';
import type { I_resource } from '@/dataType/resource.dataType';
import type { ElScrollbar } from 'element-plus';
import { ref, type PropType, onMounted } from 'vue';
import { isMobile } from '@/assets/mobile';
import { contentLayoutStyle_C } from '@/common/content'
import type { T_resourcesShowMode } from '@/dataType/app.dataType';
import { appStoreData } from '@/storeData/app.storeData';
const store = {
  appStoreData: appStoreData(),
}
const props = defineProps({
  dataList: {
    type: Array as PropType<I_resource[]>,
    default: () => [],
  },
  resourcesShowMode: {
    type: String as PropType<T_resourcesShowMode>,
    required: true,
  },

})
const emits = defineEmits(['selectResources']);

const scrollbarRef = ref<InstanceType<typeof ElScrollbar>>();
const isMobileDevice = ref(false);

const showH2 = () => {
  if (isMobileDevice.value) {
    return false;
  } else if (store.appStoreData.currentConfigApp.casualViewModule || store.appStoreData.currentConfigApp.historyModule || store.appStoreData.currentConfigApp.hotModule) {
    return true;
  }
  return false;
}
const selectResourcesHandle = (item: I_resource) => {
  emits('selectResources', item)
}

const isMobileLayoutColumnsTwo = () => {
  const clumnsTwoSlc: T_resourcesShowMode[] = ['coverPoster']
  return isMobile() && clumnsTwoSlc.includes(props.resourcesShowMode)
}
const change = () => {
  scrollbarRef.value?.setScrollTop(0);
};

onMounted(() => {
  isMobileDevice.value = isMobile();
});

defineExpose({ change });
</script>
<style lang="scss" scoped>
.layout-cover-poster-style1 {
  width: 100%;
  height: 100%;
  overflow: hidden;

  .all-h {
    font-size: 16px;

  }

  .list-ul {
    list-style-type: none;
    display: flex;
    flex-wrap: wrap;
    gap: 0.4em;
    padding-bottom: 1em;
  }
}

.mobile-layout {
  gap: 1.5em !important;

  li {
    width: 100%;
    overflow: hidden;

    /*
  aspect-ratio: 158 / 214;
  */
    :deep(.content-style) {
      width: 100% !important;

      .content-cover {
        width: 100% !important;
      }
    }

    :deep(.content-style2) {
      display: block;

      .content-info {
        width: 100% !important;
      }
    }
  }


}

.mobile-layout-columns-two {
  display: grid !important;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.8em !important;

  li {
    :deep(.content-style) {
      width: 100% !important;
    }
  }
}
</style>
