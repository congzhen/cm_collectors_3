<template>
  <div ref="tagContainerRef" class="tag-container"
    :class="[`tag-container--${appearanceStyle}`, `tag-container--tags-${tagMode}`,
      { 'tag-container--bright': isBrightTheme }]"
    :style="{ ...tagContainerStyle_C }">
    <div v-if="appearanceStyle === 'modern' && props.showModernHeader" class="modern-filter-header">
      <div class="modern-filter-title">
        <el-icon><Filter /></el-icon>
        <span>筛选</span>
      </div>
      <span class="modern-filter-subtitle">标签 / 演员 / 排序</span>
    </div>
    <el-scrollbar>
      <div class="tag-block-list">
        <el-collapse v-model="activeNames">
          <div v-for="leftDisplay, key in store.appStoreData.currentFilesBasesAppConfig.leftDisplay" :key="key">
            <tagCollapseItem v-if="leftDisplay !== E_tagType.DiyTag" :name="leftDisplay"
              :title="appLang.attributeTags(leftDisplay)" :tag-type="leftDisplay"
              :data-list="getTagDataList(leftDisplay)" :logic="getLogic(leftDisplay)"
              :appearance-style="appearanceStyle">
            </tagCollapseItem>
            <div v-else>
              <tagCollapseItem
                v-for="tagClass in store.appStoreData.currentTagClass.filter(item => item.leftShow && item.status)"
                :key="tagClass.id" :name="tagClass.id" :tag-type="E_tagType.DiyTag" :title="tagClass.name"
                :data-list="getDiyTagDataList(tagClass.id)" :diyTagClassId="tagClass.id"
                :appearance-style="appearanceStyle">
              </tagCollapseItem>
            </div>
            <div
              v-if="leftDisplay == E_tagType.Performer && store.appStoreData.currentFilesBasesAppConfig.plugInUnit_Cup">
              <tagCollapseItem name="Cup" :title="store.appStoreData.currentCupText" :tag-type="E_tagType.Cup"
                :data-list="getTagDataList(E_tagType.Cup)" :appearance-style="appearanceStyle">
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
import { ref, onMounted, onUnmounted, watch, computed, type CSSProperties } from 'vue'
import { E_searchLogic } from '@/dataType/search.dataType'
import { AppLang } from '@/language/app.lang'
import type { T_appearanceStyle } from '@/dataType/app.dataType'
const appLang = AppLang()

const props = defineProps({
  showModernHeader: {
    type: Boolean,
    default: true,
  },
})

const store = {
  appStoreData: appStoreData(),
  searchStoreData: searchStoreData(),
}
const activeNames = ref<string[]>([])
const arrowStatus = ref(false);

const tagContainerRef = ref<HTMLDivElement | null>(null)

const allId = store.searchStoreData.allId;
const allName = store.searchStoreData.allName;
const appearanceStyle = computed<T_appearanceStyle>(() =>
  store.appStoreData.appConfig.appearanceStyle
  || store.appStoreData.appConfig.headerStyle
  || 'modern'
)
const isBrightTheme = computed(() => store.appStoreData.appConfig.theme === 'bright')
const tagMode = computed(() => store.appStoreData.currentConfigApp.tagMode === 'fixed' ? 'fixed' : 'auto')

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
    height: '100%',
    left: arrowStatus.value ? '0px' : -store.appStoreData.currentConfigApp.leftColumnWidth + 'px',
    position: store.appStoreData.currentConfigApp.leftColumnMode == 'fixed' ? 'unset' : 'absolute',
    top: store.appStoreData.currentConfigApp.leftColumnMode == 'fixed' ? 'auto' : '0px',
    zIndex: 90,
    transition: 'left 0.3s ease',
    '--tag-row-num': store.appStoreData.currentConfigApp.tagFixedModeRowShowNum || 4,
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
      /*
      const sortSlc: I_tagData[] = [
        { id: 'addTimeDesc', name: '', status: false },
        { id: 'addTimeAsc', name: '', status: false },
        { id: 'issuingDateDesc', name: '', status: false },
        { id: 'issuingDateAsc', name: '', status: false },
        { id: 'issueNumberDesc', name: '', status: false },
        { id: 'issueNumberAsc', name: '', status: false },
        { id: 'scoreDesc', name: '', status: false },
        { id: 'scoreAsc', name: '', status: false },
        { id: 'starDesc', name: '', status: false },
        { id: 'starAsc', name: '', status: false },
        { id: 'titleDesc', name: '', status: false },
        { id: 'titleAsc', name: '', status: false },
        { id: 'history', name: '', status: false },
        { id: 'hot', name: '', status: false },
        //   { id: 'youLike', name: '猜你喜欢', status: false },
      ]
      */
      const sortSlc: I_tagData[] = store.appStoreData.currentConfigApp.resourceSort.map(item => {
        return {
          id: item as string,
          name: appLang.sort(item),
          status: store.searchStoreData.checkSelected(type, item)
        } as I_tagData
      })
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
    case E_tagType.VideoCodec:
      return [
        { id: allId, name: allName, status: store.searchStoreData.checkSelected(type, allId) },
        ...dataset.videoCodec.map(codec => ({
          id: codec,
          name: codec.toUpperCase(),
          status: store.searchStoreData.checkSelected(type, codec),
        })),
      ];
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
      resourceCount: tag.resourceCount,
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
    case E_tagType.VideoCodec:
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

const handleDocumentClick = (event: MouseEvent) => {
  // 只在浮动模式且启用自动隐藏时处理
  if (store.appStoreData.currentConfigApp.leftColumnMode !== 'float' ||
    !store.appStoreData.currentConfigApp.leftColumnFloatAutoHide) {
    return;
  }

  // 如果左侧栏未展开，不处理
  if (!arrowStatus.value) {
    return;
  }

  const target = event.target as HTMLElement;

  // 检查点击的元素是否在左侧栏内或是箭头按钮
  const isClickedInsideTagContainer = tagContainerRef.value?.contains(target);
  const isClickedOnArrow = target.closest('.arrow');

  // 如果点击不在左侧栏内且不是箭头按钮，则隐藏左侧栏
  if (!isClickedInsideTagContainer && !isClickedOnArrow) {
    arrowStatus.value = false;
  }
}

onMounted(() => {
  init();
  document.addEventListener('click', handleDocumentClick);
})

onUnmounted(() => {
  document.removeEventListener('click', handleDocumentClick);
})
defineExpose({ init });
</script>
<style lang="scss" scoped>
.tag-container {
  --tag-count-bg: rgba(55, 198, 202, 0.14);
  --tag-count-text: #67d4d7;
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

    :deep(.el-collapse-item.is-active .el-collapse-item__wrap) {
      overflow: visible;
    }

    :deep(.el-collapse-item__content) {
      padding-bottom: 2px;
    }
  }

  :deep(.tag-logic-container) {
    width: 100%;
  }

  .tag-block-list {
    padding-right: 6px;
    padding-bottom: 10px;
    overflow: hidden;
    background-color: var(--el-fill-color-blank);
  }

  :deep(.el-scrollbar) {
    height: 100%;
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

.tag-container.tag-container--bright {
  --tag-count-bg: rgba(27, 174, 181, 0.11);
  --tag-count-text: #138c92;
}

.tag-container--modern {
  --filter-bg: var(--home-panel-bg, #1f1f1f);
  --filter-panel-bg: var(--home-panel-bg, #1f1f1f);
  --filter-chip-bg: rgba(255, 255, 255, 0.035);
  --filter-chip-hover: rgba(48, 205, 211, 0.12);
  --filter-border: rgba(255, 255, 255, 0.09);
  --filter-text: #e4e7ed;
  --filter-muted: #a8abb2;
  --filter-accent: #37c6ca;
  --filter-accent-soft: rgba(55, 198, 202, 0.14);

  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  margin-right: 8px;
  overflow: hidden;
  color: var(--filter-text);
  background: var(--filter-panel-bg);
  border: 1px solid var(--filter-border);
  border-radius: 8px;

  .modern-filter-header {
    flex-shrink: 0;
    padding: 18px 16px 13px;
    border-bottom: 1px solid var(--filter-border);
  }

  .modern-filter-title {
    display: flex;
    align-items: center;
    gap: 8px;
    color: var(--filter-text);
    font-size: 16px;
    font-weight: 650;
  }

  .modern-filter-subtitle {
    display: block;
    margin-top: 5px;
    color: var(--filter-muted);
    font-size: 11px;
  }

  :deep(.el-scrollbar) {
    flex: 1;
    min-height: 0;
  }

  .tag-block-list {
    box-sizing: border-box;
    padding: 4px 15px 18px;
    background: transparent;
  }

  .el-collapse {
    background: transparent;

    :deep(.el-collapse-item) {
      border-bottom: 1px solid var(--filter-border);
    }

    :deep(.el-collapse-item__header) {
      height: 46px;
      color: var(--filter-text);
      background: transparent;
      border: 0;
      font-size: 13px;
    }

    :deep(.el-collapse-item__arrow) {
      order: -1;
      margin: 0 5px 0 0;
      color: var(--filter-muted);
    }

    :deep(.el-collapse-item__wrap) {
      background: transparent;
      border: 0;
    }

    :deep(.el-collapse-item__content) {
      padding: 0 0 14px;
      color: var(--filter-text);
    }
  }

  :deep(.tag-block .tag-content),
  :deep(.tag-block-performer .tag-content) {
    display: grid;
    grid-template-columns: repeat(var(--tag-row-num), minmax(0, 1fr));
    gap: 6px;
  }

  :deep(.tag-span) {
    box-sizing: border-box;
    width: auto !important;
    min-width: 0;
    padding: 8px 7px;
    color: var(--filter-muted);
    background: var(--filter-chip-bg);
    border: 1px solid var(--filter-border);
    border-radius: 5px;
    line-height: 1;
  }

  :deep(.tag-span:hover) {
    color: var(--filter-accent);
    background: var(--filter-chip-hover);
    border-color: color-mix(in srgb, var(--filter-accent) 42%, transparent);
  }

  :deep(.tag-content .check),
  :deep(.tag-performer .check),
  :deep(.tag-stars .check) {
    color: var(--filter-accent) !important;
    background: var(--filter-accent-soft) !important;
    border-color: var(--filter-accent) !important;
    box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--filter-accent) 18%, transparent);
  }

  :deep(.tag-content .check .tag-resource-count) {
    color: #ffffff;
    background: var(--filter-accent);
  }

  :deep(.tag-block-performer .tag-performer) {
    display: grid;
    grid-template-columns: repeat(var(--tag-row-num), minmax(0, 1fr));
    gap: 6px;
    padding-top: 7px;
  }

  :deep(.tag-block-performer .tag-performer-item),
  :deep(.tag-block-performer .tag-performer .tag-span) {
    box-sizing: border-box;
    width: auto !important;
    min-width: 0;
    border: 1px solid var(--filter-border);
    border-radius: 5px;
    overflow: hidden;
  }

  &.tag-container--tags-auto {
    :deep(.tag-block .tag-content),
    :deep(.tag-block-performer .tag-content),
    :deep(.tag-block-performer .tag-performer) {
      display: flex;
      flex-wrap: wrap;
      align-items: flex-start;
      gap: 6px;
    }

    :deep(.tag-block .tag-span),
    :deep(.tag-block-performer .tag-content .tag-span),
    :deep(.tag-block-performer .tag-performer .tag-span) {
      width: auto !important;
      max-width: 100%;
      flex: 0 0 auto;
    }

    :deep(.tag-block-performer--photo .tag-performer) {
      display: grid;
      grid-template-columns: repeat(var(--tag-row-num), minmax(0, 1fr));
    }
  }

  :deep(.tag-block-stars) {
    display: block;
  }

  :deep(.tag-block-stars .tag-content) {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 6px;
  }

  :deep(.tag-block-stars .tag-stars) {
    display: grid;
    grid-template-columns: 1fr;
    gap: 5px;
    margin: 7px 0 0;
    padding: 0;
  }

  :deep(.tag-block-stars .tag-stars li) {
    box-sizing: border-box;
    min-height: 30px;
    margin: 0;
    padding: 5px 7px;
    display: flex;
    align-items: center;
    background: transparent !important;
    border: 0;
    border-radius: 0;
    box-shadow: none !important;
    opacity: 0.72;
    transition: opacity 0.16s ease, transform 0.16s ease, filter 0.16s ease;
  }

  :deep(.tag-block-stars .tag-stars li:hover),
  :deep(.tag-block-stars .tag-stars li.check) {
    background: transparent !important;
    box-shadow: none !important;
  }

  :deep(.tag-block-stars .tag-stars li.check .el-rate) {
    filter: hue-rotate(135deg) saturate(1.65) brightness(1.18)
      drop-shadow(0 0 3px color-mix(in srgb, var(--filter-accent) 58%, transparent));
  }

  :deep(.tag-block-stars .tag-stars li:hover),
  :deep(.tag-block-stars .tag-stars li.check) {
    opacity: 1;
  }

  :deep(.tag-block-stars .tag-stars li.check) {
    transform: translateX(3px);
  }
}

.tag-container--modern.tag-container--bright {
  --filter-bg: #ffffff;
  --filter-panel-bg: #ffffff;
  --filter-chip-bg: #fbfcfd;
  --filter-chip-hover: rgba(28, 174, 181, 0.08);
  --filter-border: #e4e9ed;
  --filter-text: #28343c;
  --filter-muted: #77838c;
  --filter-accent: #1baeb5;
  --filter-accent-soft: rgba(27, 174, 181, 0.1);
  box-shadow: 0 8px 26px rgba(44, 62, 74, 0.06);
}
</style>
