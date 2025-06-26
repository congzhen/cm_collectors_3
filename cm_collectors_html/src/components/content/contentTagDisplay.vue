<template>
  <div class="content-tag-display">
    <contentTag v-for="item, key in topTagList_C" :key="key" :title="item.tag.name" :bg-color="item.bgColor"
      :color="item.textColor"></contentTag>
  </div>
</template>
<script setup lang="ts">
import type { I_resource, I_resourceDisplayTag } from '@/dataType/resource.dataType';
import { computed, type PropType } from 'vue';
import { appStoreData } from '@/storeData/app.storeData';
import contentTag from './contentTag.vue';
const store = {
  appStoreData: appStoreData(),
}
const props = defineProps({
  resource: {
    type: Object as PropType<I_resource>,
    required: true,
  },
})
const topTagList_C = computed(() => {
  const slc: I_resourceDisplayTag[] = [];
  const tagsToDisplay = store.appStoreData.currentConfigApp.coverDisplayTag;
  const colorList = store.appStoreData.currentConfigApp.coverDisplayTagColors;
  const rgbaList = store.appStoreData.currentConfigApp.coverDisplayTagRgbas;
  if (!tagsToDisplay) return slc;
  for (let i = 0; i < tagsToDisplay.length; i++) {
    const disTag = tagsToDisplay[i];
    for (let j = 0; j < props.resource.tags.length; j++) {
      const resTag = props.resource.tags[j];
      if (resTag.id === disTag) {
        // 使用 % 运算循环取值
        const colorIndex = colorList.length > 0 ? i % colorList.length : 0;
        const rgbaIndex = rgbaList.length > 0 ? i % rgbaList.length : 0;

        const textColor = colorList.length > 0 ? colorList[colorIndex] : "#F3F3F3";
        const bgColor = rgbaList.length > 0 ? rgbaList[rgbaIndex] : "rgba(244, 54, 16, 0.75)";

        slc.push({
          tag: resTag,
          textColor,
          bgColor,
        });
      }
    }
  }
  return slc;
});
</script>
<style lang="scss" scoped>
.content-tag-display {
  display: flex;
  flex-wrap: wrap;
  gap: 2px;
}
</style>
