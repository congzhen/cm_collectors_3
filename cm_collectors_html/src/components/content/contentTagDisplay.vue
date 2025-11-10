<template>
  <div class="content-tag-display">
    <contentTag v-for="item, key in topTagList_C" :key="key" :title="item.name" :bg-color="item.bgColor"
      :color="item.textColor"></contentTag>
  </div>
</template>
<script setup lang="ts">
import type { I_resource } from '@/dataType/resource.dataType';
import { computed, type PropType } from 'vue';
import { appStoreData } from '@/storeData/app.storeData';
import contentTag from './contentTag.vue';
import { formatDate } from '@/assets/timer'
import { AppLang } from '@/language/app.lang'
const appLang = AppLang()


interface I_contentTagDisplay {
  name: string;
  textColor: string;
  bgColor: string;
}

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
  const slc: I_contentTagDisplay[] = [];
  const tagsToDisplay = store.appStoreData.currentConfigApp.coverDisplayTag;
  const tagAttribute = store.appStoreData.currentConfigApp.coverDisplayTagAttribute;
  const colorList = store.appStoreData.currentConfigApp.coverDisplayTagColors;
  const rgbaList = store.appStoreData.currentConfigApp.coverDisplayTagRgbas;
  if (!tagsToDisplay && !tagAttribute) return slc;

  let index = 0;

  for (let i = 0; i < tagAttribute.length; i++) {
    const resAttrKey = tagAttribute[i];
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const attrData = (props.resource as Record<string, any>)[resAttrKey]
    if (attrData && attrData != '') {
      // 使用 % 运算循环取值
      const colorIndex = colorList.length > 0 ? index % colorList.length : 0;
      const rgbaIndex = rgbaList.length > 0 ? index % rgbaList.length : 0;
      const textColor = colorList.length > 0 ? colorList[colorIndex] : "#F3F3F3";
      const bgColor = rgbaList.length > 0 ? rgbaList[rgbaIndex] : "rgba(244, 54, 16, 0.75)";
      let name;
      switch (resAttrKey) {
        case 'definition':
          name = appLang.definition(attrData);
          break;
        case 'issuingDate':
          name = formatDate(attrData, 'Y')
          break;
        case 'country':
          name = appLang.country(attrData);
          break;
        case 'stars':
          name = appLang.stars(attrData);
          break;
        case 'score':
          name = appLang.score(attrData);
          break;
        case 'hot':
          name = appLang.attributeTags('hot') + ' ' + props.resource.hot;
          break;
        default:
          name = '';
      }

      slc.push({
        name: name,
        textColor,
        bgColor,
      });
    }
    index++;
  }

  for (let i = 0; i < tagsToDisplay.length; i++) {
    const disTag = tagsToDisplay[i];
    for (let j = 0; j < props.resource.tags.length; j++) {
      const resTag = props.resource.tags[j];
      if (resTag.id === disTag) {
        // 使用 % 运算循环取值
        const colorIndex = colorList.length > 0 ? index % colorList.length : 0;
        const rgbaIndex = rgbaList.length > 0 ? index % rgbaList.length : 0;

        const textColor = colorList.length > 0 ? colorList[colorIndex] : "#F3F3F3";
        const bgColor = rgbaList.length > 0 ? rgbaList[rgbaIndex] : "rgba(244, 54, 16, 0.75)";

        slc.push({
          name: resTag.name,
          textColor,
          bgColor,
        });
      }
    }
    index++;
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
