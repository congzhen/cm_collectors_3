<template>
  <div class="content-list">
    <layoutCoverPoster v-if="store.appStoreData.currentConfigApp.resourcesShowMode == 'coverPoster'"
      :data-list="props.dataList" @select-resources="selectResourcesHandle">
    </layoutCoverPoster>
    <layoutCoverPosterBox v-else-if="store.appStoreData.currentConfigApp.resourcesShowMode == 'coverPosterBox'"
      :data-list="props.dataList" @select-resources="selectResourcesHandle">
    </layoutCoverPosterBox>
    <layoutCoverPosterWaterfall
      v-else-if="store.appStoreData.currentConfigApp.resourcesShowMode == 'coverPosterWaterfall'"
      :data-list="props.dataList" @select-resources="selectResourcesHandle">
    </layoutCoverPosterWaterfall>
    <layoutTable v-else-if="store.appStoreData.currentConfigApp.resourcesShowMode == 'table'"
      :data-list="props.dataList" @select-resources="selectResourcesHandle">
    </layoutTable>
  </div>
</template>
<script lang="ts" setup>
import layoutCoverPoster from './layoutCoverPoster.vue';
import layoutCoverPosterBox from './layoutCoverPosterBox.vue';
import layoutCoverPosterWaterfall from './layoutCoverPosterWaterfall.vue';
import layoutTable from './layoutTable.vue';
import type { I_resource } from '@/dataType/resource.dataType';
import type { PropType } from 'vue';
import { appStoreData } from '@/storeData/app.storeData';
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

const selectResourcesHandle = (item: I_resource) => {
  emits('selectResources', item)
}

</script>
<style lang="scss" scoped>
.content-list {
  width: 100%;
  height: 100%;
  overflow: hidden;

  :deep(.play-icon) {
    .el-icon {
      color: #f3f3f3;
      filter: drop-shadow(0 0 4px rgba(0, 0, 0, 0.9));
    }
  }
}
</style>
