<template>
  <div ref="tagContainerRef" class="tag-container" :style="{ ...tagContainerStyle_C }">
    <el-scrollbar>
      <div class="tag-block-list">
        <el-collapse v-model="activeNames">
          <div v-for="leftDisplay, key in store.appStoreData.currentFilesBasesAppConfig.leftDisplay" :key="key">
            <tagCollapseItem v-if="leftDisplay !== E_tagType.DiyTag" :name="leftDisplay"
              :title="appLang.attributeTags(leftDisplay)" :tag-type="leftDisplay"
              :data-list="getTagDataList(leftDisplay)" :logic="getLogic(leftDisplay)">
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

    <div ref="arrowRef" class="arrow" v-if="store.appStoreData.currentConfigApp.leftColumnMode !== 'fixed'"
      :style="{ left: arrowLeftStyle }" @click="arrowClickHandle">
      <el-icon>
        <ArrowLeftBold v-if="arrowStatus" />
        <ArrowRightBold v-else />
      </el-icon>
    </div>

  </div>
</template>
<script setup lang="ts">
import dataset from '@/assets/dataset'
import { E_tagType, type I_tagData } from '@/dataType/app.dataType'
import tagCollapseItem from '@/components/tag/tagCollapseItem.vue'
import { appStoreData } from '@/storeData/app.storeData'
import { searchStoreData } from '@/storeData/search.storeData'
import { ref, onMounted, watch, computed, type CSSProperties } from 'vue'
import { E_searchLogic } from '@/dataType/search.dataType'
import { appLang } from '@/language/app.lang'
const store = {
  appStoreData: appStoreData(),
  searchStoreData: searchStoreData(),
}
const activeNames = ref<string[]>([])
const arrowStatus = ref(false);

const tagContainerRef = ref<HTMLDivElement | null>(null)

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


const tagContainerStyle_C = computed<CSSProperties>(() => {
  return {
    width: store.appStoreData.currentConfigApp.leftColumnWidth + 'px',
    height: store.appStoreData.currentConfigApp.leftColumnMode == 'fixed' ? '100%' : '100%',
    left: arrowStatus.value ? '0px' : -store.appStoreData.currentConfigApp.leftColumnWidth + 'px',
    position: store.appStoreData.currentConfigApp.leftColumnMode == 'fixed' ? 'unset' : 'absolute',
    zIndex: 90,
    transition: 'left 0.3s ease',
  }
})

const arrowLeftStyle = computed<string>(() => {
  if (arrowStatus.value) {
    return store.appStoreData.currentConfigApp.leftColumnWidth + 'px';
  } else {
    return '0px';
  }
});



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
        { id: 'addTimeDesc', name: '', status: false },
        { id: 'addTimeAsc', name: '', status: false },
        { id: 'issueNumberDesc', name: '', status: false },
        { id: 'issueNumberAsc', name: '', status: false },
        { id: 'starDesc', name: '', status: false },
        { id: 'starAsc', name: '', status: false },
        { id: 'titleDesc', name: '', status: false },
        { id: 'titleAsc', name: '', status: false },
        { id: 'history', name: '', status: false },
        { id: 'hot', name: '', status: false },
        //   { id: 'youLike', name: '猜你喜欢', status: false },
      ]
      for (let i = 0; i < sortSlc.length; i++) {
        sortSlc[i].name = appLang.sort(sortSlc[i].id)
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
          name: appLang.country(item),
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
          name: appLang.definition(item),
          status: store.searchStoreData.checkSelected(type, item),
        });
      });
      return resultDefinitionArr;
    case E_tagType.Year:
      const currentYear = new Date().getFullYear();
      const years: I_tagData[] = [
        { id: allId, name: allName, status: store.searchStoreData.checkSelected(type, allId) }
      ];
      for (let year = currentYear; year >= 2000; year--) {
        const _id = year.toString();
        years.push({ name: appLang.year(_id), id: _id, status: store.searchStoreData.checkSelected(type, _id) });
      }
      // 添加 "2000年以前"
      years.push({ name: appLang.year('before_2000'), id: 'before_2000', status: store.searchStoreData.checkSelected(type, 'before_2000') });

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

const arrowClickHandle = () => {
  arrowStatus.value = !arrowStatus.value;
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
  background-color: #1F1F1F;

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

  .arrow {
    width: 15px;
    height: 80px;
    line-height: 82px;
    overflow: hidden;
    background-color: #262727;
    color: #E4E7ED;
    position: fixed;
    top: 50%;
    margin-top: -40px;
    border-top-right-radius: 7px;
    border-bottom-right-radius: 7px;
    z-index: 89;
    cursor: pointer;

    &:hover {
      background-color: #79BBFF;
    }
  }

}
</style>
