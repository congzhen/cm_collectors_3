<template>
  <div class="tag-container">
    <el-scrollbar>
      <div class="tag-block-list">
        <el-collapse v-model="activeNames">
          <div v-for="leftDisplay, key in store.appStoreData.currentFilesBasesAppConfig.leftDisplay" :key="key">
            <tagCollapseItem v-if="leftDisplay !== E_tagType.DiyTag" :name="leftDisplay" :title="leftDisplay"
              :tag-type="leftDisplay" :data-list="getTagDataList(leftDisplay)">
            </tagCollapseItem>
            <div v-else>
              <tagCollapseItem
                v-for="tagClass in store.appStoreData.currentTagClass.filter(item => item.leftShow && item.status)"
                :key="tagClass.id" :name="tagClass.id" :tag-type="E_tagType.DiyTag" :title="tagClass.name"
                :data-list="getDiyTagDataList(tagClass.id)">
              </tagCollapseItem>
            </div>
            <div
              v-if="leftDisplay == E_tagType.Performer && store.appStoreData.currentFilesBasesAppConfig.plugInUnit_Cup">
              <tagCollapseItem name="Cup" :title="store.appStoreData.currentCupText" :tag-type="E_tagType.Cup"
                :data-list="getTagDataList(E_tagType.Cup)">
              </tagCollapseItem>
            </div>
          </div>
        </el-collapse>
      </div>
    </el-scrollbar>
  </div>
</template>
<script setup lang="ts">
import dataset from '@/assets/dataset'
import { E_tagType, type I_tagData } from '@/dataType/app.dataType'
import tagCollapseItem from '@/components/tag/tagCollapseItem.vue'
import { appStoreData } from '@/storeData/app.storeData'
import { ref, onMounted, watch } from 'vue'
const store = {
  appStoreData: appStoreData(),
}
const activeNames = ref<string[]>([])

watch(
  () => [
    store.appStoreData.currentFilesBasesAppConfig.leftDisplay,
    store.appStoreData.currentTagClass
  ],
  () => {
    init();
  },
  { deep: true }
);

const init = () => {
  activeNames.value = [
    ...store.appStoreData.currentFilesBasesAppConfig.leftDisplay,
    ...store.appStoreData.currentTagClass.filter(item => item.leftShow).map(item => item.id),
    'Cup'
  ]
}


const getTagDataList = (type: E_tagType): I_tagData[] => {
  switch (type) {
    case E_tagType.Sort:
      return [
        { id: 'addTimeDesc', name: '时间倒叙' },
        { id: 'addTimeAsc', name: '时间正序' },
        { id: 'issueNumberDesc', name: '发行倒叙' },
        { id: 'issueNumberAsc', name: '发行正序' },
        { id: 'starDesc', name: '评分倒叙' },
        { id: 'starAsc', name: '评分正序' },
        { id: 'titleDesc', name: '标题倒叙' },
        { id: 'titleAsc', name: '标题正序' },
      ];
    case E_tagType.Country:
      const resultCountryArr: I_tagData[] = [];
      store.appStoreData.currentFilesBasesAppConfig.country.forEach(item => {
        resultCountryArr.push({
          id: item,
          name: item,
        });
      });
      return resultCountryArr;
    case E_tagType.Definition:
      const resultDefinitionArr: I_tagData[] = [];
      store.appStoreData.currentFilesBasesAppConfig.definition.forEach(item => {
        resultDefinitionArr.push({
          id: item,
          name: item,
        });
      });
      return resultDefinitionArr;
    case E_tagType.Year:
      const currentYear = new Date().getFullYear();
      const years: I_tagData[] = [];
      for (let year = currentYear; year >= 2001; year--) {
        years.push({ name: `${year}年`, id: year.toString() });
      }
      // 添加 "2000年以前"
      years.push({ name: '2000前', id: 'before_2000' });

      return years;
    case E_tagType.Performer:
      return []
    case E_tagType.Cup:
      const resultCupArr: I_tagData[] = [];
      dataset.cup.forEach(cup => {
        resultCupArr.push({
          id: cup,
          name: cup + '-' + store.appStoreData.currentCupText
        });
      });
      return resultCupArr;
    default:
      return []
  }
}
const getDiyTagDataList = (tagClassId: string) => {
  const result: I_tagData[] = [];
  store.appStoreData.currentTagsByTagClassId(tagClassId).forEach(tag => {
    result.push({
      id: tag.id,
      name: tag.name,
    });
  });
  return result;
}

onMounted(() => {
  init();
})
defineExpose({ init });
</script>
<style lang="scss" scoped>
.tag-container {
  width: 24.4em;
  height: 100%;
  overflow: hidden;
  border-right: 0.1em solid #414243;
  margin-right: 6px;

  .el-collapse {
    border: 0;

    :deep(.el-collapse-item__header) {
      border: 0;
      height: 3em;
      line-height: 3em;
    }

    :deep(.el-collapse-item__wrap) {
      border: 0;
    }

    :deep(.el-collapse-item__content) {
      padding-bottom: 2px;
    }
  }

  .tag-block-list {
    padding-right: 6px;
    padding-bottom: 10px;
  }
}
</style>
