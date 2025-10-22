<template>
  <el-image :src="getResourceCoverPoster(props.resource)" :fit="fit_C" />
</template>
<script setup lang="ts">
import { getResourceCoverPoster } from '@/common/photo';
import type { I_resource } from '@/dataType/resource.dataType';
import { appStoreData } from '@/storeData/app.storeData';
import { computed, type PropType } from 'vue';
const store = {
  appStoreData: appStoreData(),
}
const props = defineProps({
  resource: {
    type: Object as PropType<I_resource>,
    required: true,
  },
})


const fit_C = computed(() => {
  if (store.appStoreData.currentConfigApp.coverImageFit) {
    return store.appStoreData.currentConfigApp.coverImageFit
  }
  return 'cover'
})
</script>
