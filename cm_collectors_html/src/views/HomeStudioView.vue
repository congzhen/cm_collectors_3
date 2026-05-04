<template>
  <div class="index-view studio-home" :class="{ 'bright-studio': isBrightTheme }" v-loading="loading">
    <HeaderView class="header" @create-resouce-success="createResouceSuccessHandle"></HeaderView>

    <div class="studio-shell" :class="{ 'filter-open': filterPanelVisible }" :style="studioShellStyle">
      <HomeLibraryRail :files-bases-list="store.filesBasesStoreData.filesBasesStatus"
        :current-files-bases-id="store.appStoreData.currentFilesBases.id" :tag-class-count="tagClassCount"
        :performer-bases-count="performerBasesCount" @select="selectFilesBaseHandle" />

      <HomeToolDrawer v-if="filterPanelVisible" title="筛选" subtitle="标签 / 演员 / 排序"
        @close="setFilterPanelVisible(false)">
        <TagView ref="tagViewRef" class="tag"></TagView>
      </HomeToolDrawer>

      <main class="media-stage"
        :class="{ 'filter-open': filterPanelVisible, 'admin-mode': store.appStoreData.adminResourceStatus }">
        <HomeStageHero :title="currentFilesBaseName" :subtitle="contentSubtitle" :filter-visible="filterPanelVisible"
          @toggle-filter="setFilterPanelVisible(!filterPanelVisible)" />

        <section class="workspace-grid">
          <section class="stage-canvas">
            <ContentView ref="contentViewRef" class="content" @select-resources="selectResourcesHandle"></ContentView>
          </section>
        </section>
      </main>
    </div>
    <videoPlayDialog ref="videoPlayDialogRef"></videoPlayDialog>
    <resourceFormDrawer ref="resourceFormDrawerRef" @success="updateResouceSuccessHandle"></resourceFormDrawer>
    <resourceSetTagDialog ref="resourceSetTagDialogRef" @success="updateResouceSuccessHandle"></resourceSetTagDialog>
    <resourceSetPerformerDialog ref="resourceSetPerformerDialogRef" @success="updateResouceSuccessHandle">
    </resourceSetPerformerDialog>
    <playCloudCheckPromptDialog ref="playCloudCheckPromptDialogRef"></playCloudCheckPromptDialog>
  </div>
</template>

<script setup lang="ts">
import HeaderView from './HeaderView.vue'
import HomeLibraryRail from '@/components/home/HomeLibraryRail.vue'
import HomeStageHero from '@/components/home/HomeStageHero.vue'
import HomeToolDrawer from '@/components/home/HomeToolDrawer.vue'
import TagView from './TagView.vue'
import ContentView from './ContentView.vue'
import DetailsView from './DetailsView.vue'
import videoPlayDialog from '@/components/play/videoPlayDialog.vue'
import resourceFormDrawer from '@/components/resource/resourceFormDrawer.vue'
import type { I_resource } from '@/dataType/resource.dataType'
import { computed, ref, onMounted } from 'vue'
import type { I_filesBases } from '@/dataType/filesBases.dataType'
import { appStoreData } from '@/storeData/app.storeData'
import { filesBasesStoreData } from '@/storeData/filesBases.storeData'
import { searchStoreData } from '@/storeData/search.storeData'
import { ElMessage } from 'element-plus'
import { eventBus } from '@/main'
import resourceSetTagDialog from '@/components/resource/resourceSetTagDialog.vue'
import resourceSetPerformerDialog from '@/components/resource/resourceSetPerformerDialog.vue'
import playCloudCheckPromptDialog from '@/components/play/playCloudCheckPromptDialog.vue'
import { resourcePinToTop } from '@/common/resource'

const store = {
  appStoreData: appStoreData(),
  filesBasesStoreData: filesBasesStoreData(),
  searchStoreData: searchStoreData(),
}

const tagViewRef = ref<InstanceType<typeof TagView>>();
const contentViewRef = ref<InstanceType<typeof ContentView>>();
const detailsViewRef = ref<InstanceType<typeof DetailsView>>();
const videoPlayDialogRef = ref<InstanceType<typeof videoPlayDialog>>();
const resourceFormDrawerRef = ref<InstanceType<typeof resourceFormDrawer>>();
const resourceSetTagDialogRef = ref<InstanceType<typeof resourceSetTagDialog>>();
const resourceSetPerformerDialogRef = ref<InstanceType<typeof resourceSetPerformerDialog>>();
const playCloudCheckPromptDialogRef = ref<InstanceType<typeof playCloudCheckPromptDialog>>();

const loading = ref(false);
const resDetails = ref<I_resource | undefined>(undefined);
const filterPanelStorageKey = 'cm-studio-filter-panel-visible'
const filterPanelVisible = ref(localStorage.getItem(filterPanelStorageKey) == 'true');

const resourceModeLabels: Record<string, string> = {
  coverPoster: '海报',
  coverPosterBox: '盒子',
  coverPosterBoxWideSeparate: '宽盒子',
  coverPosterSimple: '简洁',
  coverPosterSimpleExpand: '展开',
  coverPosterWaterfall: '瀑布',
  shortVideo: '短视频',
  table: '表格',
}

const currentFilesBaseName = computed(() => {
  return store.appStoreData.currentFilesBases.name || store.appStoreData.getLogoName;
})

const isBrightTheme = computed(() => store.appStoreData.appConfig.theme == 'bright')
const performerBasesCount = computed(() => store.appStoreData.currentPerformerBasesIds.length)
const tagClassCount = computed(() => store.appStoreData.currentTagClass.filter(item => item.status).length)
const studioShellStyle = computed(() => {
  return {
    '--studio-filter-width': `${store.appStoreData.currentConfigApp.leftColumnWidth || 319}px`,
    '--studio-tag-row-num': store.appStoreData.currentConfigApp.tagFixedModeRowShowNum || 4,
  }
})
const contentSubtitle = computed(() => {
  const mode = resourceModeLabels[store.appStoreData.currentConfigApp.resourcesShowMode] || '自定义';
  const pageLimit = store.appStoreData.currentConfigApp.pageLimit || 0;
  return `${mode}视图 · 每页 ${pageLimit} 项`;
})

const setFilterPanelVisible = (visible: boolean) => {
  filterPanelVisible.value = visible;
  localStorage.setItem(filterPanelStorageKey, visible ? 'true' : 'false');
}

const selectFilesBaseHandle = async (filesBases: I_filesBases) => {
  loading.value = true;
  resDetails.value = undefined;
  const result = await store.appStoreData.initCurrentFilesBases(filesBases.id)
  if (result && !result.status) {
    ElMessage.error(result.message);
    return
  }
  store.searchStoreData.init();
  contentViewRef.value?.init();
  loading.value = false;
};

const selectResourcesHandle = (resource: I_resource | undefined, isInit: boolean) => {
  resDetails.value = resource;
  if (!isInit) {
    detailsViewRef.value?.init();
  }
}

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const createResouceSuccessHandle = (_data: I_resource) => {
  contentViewRef.value?.init();
}

const updateResouceSuccessHandle = async (data: I_resource) => {
  await contentViewRef.value?.init_DataList();
  resDetails.value = data;
}

const deleteResouceSuccessHandle = () => {
  contentViewRef.value?.init_DataList(() => { }, true);
}

const resourceDialogPlayStartHandle = (event: unknown) => {
  const typedEvent = event as { resourceId: string; dramaSeriesId: string };
  videoPlayDialogRef.value?.open(typedEvent.resourceId, typedEvent.dramaSeriesId);
}

const editResourceHandle = (event: unknown) => {
  const typedEvent = event as { resource: I_resource; };
  resourceFormDrawerRef.value?.open('edit', typedEvent.resource)
}

const editResourcePinToTopHandle = (event: unknown) => {
  const typedEvent = event as { resource: I_resource; pinToTopStatus?: boolean };
  resourcePinToTop(typedEvent.resource, () => {
    contentViewRef.value?.init_DataList(() => { }, true);
  }, typedEvent.pinToTopStatus)
}

const editResourceTagHandle = (event: unknown) => {
  const typedEvent = event as { resource: I_resource; };
  resourceSetTagDialogRef.value?.open(typedEvent.resource)
}

const editResourcePerformerHandle = (event: unknown) => {
  const typedEvent = event as { resource: I_resource; };
  resourceSetPerformerDialogRef.value?.open(typedEvent.resource)
}

const deleteResouceSuccessOnHandle = () => {
  deleteResouceSuccessHandle()
}

const playCloundHandle = (event: unknown) => {
  const typedEvent = event as { resourceId: string; dramaSeriesId: string, playSrc: string };
  playCloudCheckPromptDialogRef.value?.open(typedEvent.dramaSeriesId)
}

onMounted(() => {
  eventBus.on('resource-dialog-play-start', resourceDialogPlayStartHandle);
  eventBus.on('edit-resource', editResourceHandle);
  eventBus.on('edit-resource-pinToTop', editResourcePinToTopHandle);
  eventBus.on('edit-resource-tag', editResourceTagHandle);
  eventBus.on('edit-resource-performer', editResourcePerformerHandle);
  eventBus.on('delete-resource-success', deleteResouceSuccessOnHandle);
  eventBus.on('playClound', playCloundHandle);
})
</script>

<style lang="scss" scoped>
.studio-home {
  --home-bg: #16181d;
  --home-panel-bg: #20232a;
  --home-panel-soft-bg: #262a32;
  --home-stage-bg: #1b1e24;
  --home-content-bg: #191c22;
  --home-text: #f2f3f5;
  --home-text-muted: #a8abb2;
  --home-border: rgba(255, 255, 255, 0.09);
  --home-accent: #409eff;
  --home-accent-soft: rgba(64, 158, 255, 0.14);
  --home-shadow: 0 12px 32px rgba(0, 0, 0, 0.28);

  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 8px;
  color: var(--home-text);
  background: var(--home-bg);

  .header {
    width: 100%;
    flex-shrink: 0;
    padding: 6px 10px;
    border: 1px solid var(--home-border);
    border-radius: 8px;
    background: var(--home-panel-bg);
    box-sizing: border-box;
  }

  .studio-shell {
    flex: 1;
    min-height: 0;
    display: grid;
    grid-template-columns: 210px var(--studio-filter-width) minmax(0, 1fr);
    gap: 8px;
    overflow: hidden;
  }

  .studio-shell:not(.filter-open) {
    grid-template-columns: 210px minmax(0, 1fr);
  }

  .media-stage {
    min-width: 0;
    min-height: 0;
    padding: 10px;
    display: flex;
    flex-direction: column;
    gap: 8px;
    overflow: hidden;
    border: 1px solid var(--home-border);
    border-radius: 8px;
    background: var(--home-stage-bg);
  }

  .workspace-grid {
    flex: 1;
    min-height: 0;
    display: grid;
    grid-template-columns: 1fr;
    overflow: hidden;
  }

  .stage-canvas {
    min-width: 0;
    min-height: 0;
    overflow: hidden;
    border-radius: 8px;
    background: var(--home-content-bg);
    box-shadow: inset 0 0 0 1px var(--home-border);
  }

  .content {
    width: 100%;
    height: 100%;
    overflow: hidden;
  }

  :deep(.content-view) {
    padding: 8px 8px 6px;
    box-sizing: border-box;
  }

  :deep(.content-view .list) {
    min-height: 0;
    padding-bottom: 8px;
    box-sizing: border-box;
  }

  :deep(.content-view .paging) {
    width: 100%;
    padding: 8px 4px 0;
    border-top: 1px solid var(--home-border);
    box-sizing: border-box;
  }

  .tag {
    flex: 1;
    min-height: 0;
    width: 100% !important;
    position: relative !important;
    left: auto !important;
    top: auto !important;
    color: var(--home-text);
  }

  .details {
    flex: 1;
    min-height: 0;
    overflow: hidden;
  }

  .empty-detail {
    flex: 1;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    gap: 10px;
    color: var(--home-text-muted);

    .el-icon {
      font-size: 38px;
      color: var(--home-accent);
    }
  }

  :deep(.details-view-k) {
    width: 100%;
    padding-left: 0;
  }

  :deep(.tag-container) {
    background-color: transparent;
  }
}

.studio-home.bright-studio {
  --home-bg: #f4f6f8;
  --home-panel-bg: #ffffff;
  --home-panel-soft-bg: #f5f7fa;
  --home-stage-bg: #ffffff;
  --home-content-bg: #ffffff;
  --home-text: #303133;
  --home-text-muted: #606266;
  --home-border: #dcdfe6;
  --home-accent: #409eff;
  --home-accent-soft: #ecf5ff;
  --home-shadow: 0 10px 28px rgba(31, 45, 61, 0.08);

  color: var(--home-text);
  background: var(--home-bg);
}

.studio-home.bright-studio {
  :deep(.el-button:not(.el-button--primary)) {
    --el-button-bg-color: #ffffff;
    --el-button-border-color: #dcdfe6;
    --el-button-text-color: #606266;
    --el-button-hover-bg-color: #ecf5ff;
    --el-button-hover-border-color: #409eff;
    --el-button-hover-text-color: #409eff;
  }

  :deep(.el-pagination.is-background .btn-prev),
  :deep(.el-pagination.is-background .btn-next),
  :deep(.el-pagination.is-background .el-pager li) {
    background-color: #f4f6f8;
    color: #606266;
  }

  :deep(.el-pagination.is-background .el-pager li.is-active) {
    background-color: #409eff;
    color: #ffffff;
  }

  :deep(.el-table),
  :deep(.el-table tr),
  :deep(.el-table th.el-table__cell) {
    background-color: #ffffff;
    color: #303133;
  }

  :deep(.el-scrollbar__view),
  :deep(.content-view),
  :deep(.content-list) {
    color: #303133;
  }

  :deep(.tag-container),
  :deep(.tag-block-list),
  :deep(.el-collapse),
  :deep(.el-collapse-item__wrap),
  :deep(.el-collapse-item__header) {
    background-color: transparent;
    color: #303133;
    border-color: #dcdfe6;
  }

  :deep(.tag-container .tag-content .tag-span) {
    background-color: #f3f3f3;
    color: #606266;
  }

  :deep(.tag-container .tag-content .check) {
    background-color: #ffaa47 !important;
    color: #ffffff !important;
  }

  :deep(.tag-container .tag-block-stars .tag-span li),
  :deep(.tag-container .tag-block-stars .tag-stars li) {
    background-color: transparent;
  }

  :deep(.tag-container .tag-block-stars .tag-content .check),
  :deep(.tag-container .tag-block-stars .tag-stars .check),
  :deep(.tag-container .tag-block-performer .tag-content .check) {
    background-color: #ffaa47 !important;
    color: #ffffff !important;
  }

  :deep(.tag-container .tag-block-performer .tag-content .tag-span) {
    background-color: #f3f3f3;
    color: #606266;
  }
}

.studio-home {
  :deep(.tag-container .tag-block-performer .tag-performer) {
    display: grid;
    grid-template-columns: repeat(var(--studio-tag-row-num), minmax(0, 1fr));
    gap: 0.3em;
  }

  :deep(.tag-container .tag-block-performer .tag-performer .tag-performer-item),
  :deep(.tag-container .tag-block-performer .tag-performer .tag-span) {
    width: auto !important;
    min-width: 0;
  }
}

@media (max-width: 1100px) {
  .studio-home {
    .studio-shell {
      grid-template-columns: 1fr;
    }

    .studio-shell:not(.filter-open),
    .workspace-grid {
      grid-template-columns: 1fr;
    }
  }
}
</style>
