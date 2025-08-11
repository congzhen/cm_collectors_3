<template>
  <div class="content-list">
    <layoutCoverPoster v-if="store.appStoreData.currentConfigApp.resourcesShowMode == 'coverPoster'"
      :data-list="props.dataList" @select-resources="selectResourcesHandle">
    </layoutCoverPoster>
    <layoutCoverPosterBox v-else-if="store.appStoreData.currentConfigApp.resourcesShowMode == 'coverPosterBox'"
      :data-list="props.dataList" @select-resources="selectResourcesHandle">
    </layoutCoverPosterBox>
    <layoutTable v-else-if="store.appStoreData.currentConfigApp.resourcesShowMode == 'table'"
      :data-list="props.dataList" @select-resources="selectResourcesHandle">
    </layoutTable>
  </div>
</template>
<script lang="ts" setup>
import layoutCoverPoster from './layoutCoverPoster.vue';
import layoutCoverPosterBox from './layoutCoverPosterBox.vue';
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

}
</style>
