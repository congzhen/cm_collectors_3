<template>
  <div class="tag-container">
    <el-scrollbar>
      <div class="tag-block-list">
        <el-collapse v-model="activeNames">
          <div v-for="leftDisplay, key in store.appStoreData.currentFilesBasesAppConfig.leftDisplay" :key="key">
            <tagCollapseItem v-if="leftDisplay !== E_tagType.DiyTag" :name="leftDisplay" :title="leftDisplay"
              :tag-type="leftDisplay" :data-list="getTagDataList(leftDisplay)" :logic="getLogic(leftDisplay)">
            </tagCollapseItem>
            <div v-else>
              <tagCollapseItem
                v-for="tagClass in store.appStoreData.currentTagClass.filter(item => item.leftShow && item.status)"
                :key="tagClass.id" :name="tagClass.id" :tag-type="E_tagType.DiyTag" :title="tagClass.name"
                :data-list="getDiyTagDataList(tagClass.id)" :diyTagClassId="tagClass.id">
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
import { searchStoreData } from '@/storeData/search.storeData'
import { ref, onMounted, watch } from 'vue'
import { E_searchLogic } from '@/dataType/search.dataType'
const store = {
  appStoreData: appStoreData(),
  searchStoreData: searchStoreData(),
}
const activeNames = ref<string[]>([])

const allId = store.searchStoreData.allId;
const allName = store.searchStoreData.allName;

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
      const sortSlc: I_tagData[] = [
        { id: 'addTimeDesc', name: '时间倒叙', status: false },
        { id: 'addTimeAsc', name: '时间正序', status: false },
        { id: 'issueNumberDesc', name: '发行倒叙', status: false },
        { id: 'issueNumberAsc', name: '发行正序', status: false },
        { id: 'starDesc', name: '评分倒叙', status: false },
        { id: 'starAsc', name: '评分正序', status: false },
        { id: 'titleDesc', name: '标题倒叙', status: false },
        { id: 'titleAsc', name: '标题正序', status: false },
        { id: 'history', name: '历史记录', status: false },
        { id: 'hot', name: '当前热度', status: false },
        { id: 'youLike', name: '猜你喜欢', status: false },
      ]
      for (let i = 0; i < sortSlc.length; i++) {
        sortSlc[i].status = store.searchStoreData.checkSelected(type, sortSlc[i].id)
      }
      return sortSlc
    case E_tagType.Country:
      const resultCountryArr: I_tagData[] = [
        { id: allId, name: allName, status: store.searchStoreData.checkSelected(type, allId) }
      ];
      store.appStoreData.currentFilesBasesAppConfig.country.forEach(item => {
        resultCountryArr.push({
          id: item,
          name: item,
          status: store.searchStoreData.checkSelected(type, item),
        });
      });
      return resultCountryArr;
    case E_tagType.Definition:
      const resultDefinitionArr: I_tagData[] = [
        { id: allId, name: allName, status: store.searchStoreData.checkSelected(type, allId) }
      ];
      store.appStoreData.currentFilesBasesAppConfig.definition.forEach(item => {
        resultDefinitionArr.push({
          id: item,
          name: item,
          status: store.searchStoreData.checkSelected(type, item),
        });
      });
      return resultDefinitionArr;
    case E_tagType.Year:
      const currentYear = new Date().getFullYear();
      const years: I_tagData[] = [
        { id: allId, name: allName, status: store.searchStoreData.checkSelected(type, allId) }
      ];
      for (let year = currentYear; year >= 2001; year--) {
        const _id = year.toString();
        years.push({ name: `${year}年`, id: _id, status: store.searchStoreData.checkSelected(type, _id) });
      }
      // 添加 "2000年以前"
      years.push({ name: '2000前', id: 'before_2000', status: store.searchStoreData.checkSelected(type, 'before_2000') });

      return years;
    case E_tagType.Performer:
      return []
    case E_tagType.Cup:
      const resultCupArr: I_tagData[] = [
        { id: allId, name: allName, status: store.searchStoreData.checkSelected(type, allId) }
      ];
      dataset.cup.forEach(cup => {
        resultCupArr.push({
          id: cup,
          name: store.appStoreData.cupText(cup),
          status: store.searchStoreData.checkSelected(type, cup),
        });
      });
      return resultCupArr;
    default:
      return []
  }
}
const getDiyTagDataList = (tagClassId: string) => {
  const result: I_tagData[] = [
    { id: allId, name: allName, status: store.searchStoreData.checkSelected(E_tagType.DiyTag, allId, tagClassId) }
  ];
  store.appStoreData.currentTagsByTagClassId(tagClassId).forEach(tag => {
    result.push({
      id: tag.id,
      name: tag.name,
      status: store.searchStoreData.checkSelected(E_tagType.DiyTag, tag.id, tag.tagClass_id),
    });
  });
  return result;
}

const getLogic = (type: E_tagType) => {
  switch (type) {
    case E_tagType.Sort:
      return [E_searchLogic.Single];
    case E_tagType.Country:
    case E_tagType.Definition:
    case E_tagType.Year:
    case E_tagType.Star:
      return [E_searchLogic.Single, E_searchLogic.MultiOr];
    case E_tagType.Performer:
    case E_tagType.Cup:
      return [E_searchLogic.Single, E_searchLogic.MultiOr, E_searchLogic.MultiAnd];
    default:
      return [];
  }
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
