<template>
  <div class="cover-adjuster-wrapper">
    <el-popover placement="top" :width="isModernAppearance ? 980 : 800" trigger="click"
      :popper-class="isModernAppearance ? 'cover-adjuster-popover-modern' : ''">
      <template #reference>
        <el-button icon="Monitor" size="small" title="显示设置">
          <span v-if="isModernAppearance">显示设置</span>
        </el-button>
      </template>
      <div v-if="isModernAppearance" class="cover-adjuster-container cover-adjuster-container--modern"
        :class="{ 'cover-adjuster-container--bright': isBrightTheme }">
        <div class="modern-adjuster-grid">
          <section class="modern-settings-card modern-mode-card">
            <div class="modern-section-title">
              <el-icon><Menu /></el-icon>
              <span>显示模式</span>
            </div>
            <div class="modern-mode-list">
              <button v-for="item, index in resourcesShowModeList" :key="index" type="button"
                class="modern-mode-item" :class="{ active: resourcesShowMode === item }"
                @click="resourcesShowMode = item">
                <el-icon><Grid /></el-icon>
                <span>{{ $t(`resourcesShowMode.${item}`) }}</span>
              </button>
            </div>
          </section>

          <div class="modern-main-settings">
            <section class="modern-settings-card">
              <div class="modern-section-title">
                <el-icon><Operation /></el-icon>
                <span>尺寸与间距</span>
              </div>
              <div class="modern-slider-list">
                <div class="modern-slider-item">
                  <el-checkbox v-model="coverPosterWidthStatus" label="锁定宽度" />
                  <el-slider v-model="coverPosterWidthBase" :min="10" :max="1000" size="small" show-input />
                </div>
                <div class="modern-slider-item">
                  <el-checkbox v-model="coverPosterHeightStatus" label="锁定高度" />
                  <el-slider v-model="coverPosterHeightBase" :min="10" :max="1000" size="small" show-input />
                </div>
                <div class="modern-slider-item">
                  <span class="modern-control-label">资源间距</span>
                  <el-slider v-model="coverPosterGap" :min="0" :max="50" :step="0.1" size="small" show-input />
                </div>
                <div class="modern-slider-item">
                  <span class="modern-control-label">左右空距</span>
                  <el-slider v-model="contentPadding" :min="0" :max="50" size="small" show-input />
                </div>
                <div class="modern-slider-item">
                  <span class="modern-control-label">封面标签大小</span>
                  <el-slider v-model="coverDisplayTagFontSize" :min="8" :max="24" size="small" show-input />
                </div>
              </div>
            </section>

            <section class="modern-settings-card">
              <div class="modern-section-title">
                <el-icon><Picture /></el-icon>
                <span>封面内容</span>
              </div>
              <div class="modern-option-item">
                <span class="modern-control-label">封面图填充方式</span>
                <el-radio-group v-model="coverImageFit" size="small">
                  <el-radio-button v-for="item, index in dataset.coverImageFit" :key="index" :label="item"
                    :value="item" />
                </el-radio-group>
              </div>
              <div class="modern-option-item">
                <span class="modern-control-label">资源对齐方式</span>
                <el-radio-group v-model="resourceJustifyContent" size="small">
                  <el-radio-button label="start" value="flex-start" />
                  <el-radio-button label="center" value="center" />
                  <el-radio-button label="end" value="flex-end" />
                  <el-radio-button label="between" value="space-between" />
                  <el-radio-button label="around" value="space-around" />
                </el-radio-group>
              </div>
              <el-checkbox v-model="showVideoDuration" label="显示视频时长" />
            </section>
          </div>

          <section class="modern-settings-card modern-page-card">
            <div class="modern-section-title">
              <el-icon><SetUp /></el-icon>
              <span>页面与交互</span>
            </div>
            <div class="modern-page-settings">
              <div class="modern-option-item">
                <span class="modern-control-label">分页显示数</span>
                <el-input-number v-model="pageLimit" />
              </div>
              <div class="modern-option-item">
                <span class="modern-control-label">左侧边栏显示模式</span>
                <selectLeftColumnMode v-model="leftColumnMode" :teleported="false" />
              </div>
              <el-checkbox v-model="leftColumnFloatAutoHide" label="浮动模式自动隐藏" />
              <el-checkbox v-model="showCustomTagResourceCount" label="显示自定义标签资源数量" />
              <div class="modern-option-item">
                <span class="modern-control-label">详情显示模式</span>
                <selectResourceDetailsShowMode v-model="resourceDetailsShowMode" :teleported="false" />
              </div>
              <div class="modern-option-item">
                <span class="modern-control-label">封面标题对齐方式</span>
                <el-radio-group v-model="coverTitleAlign" size="small">
                  <el-radio-button label="左对齐" value="left" />
                  <el-radio-button label="居中" value="center" />
                  <el-radio-button label="右对齐" value="right" />
                </el-radio-group>
              </div>
              <div class="modern-module-options">
                <el-checkbox label="随便看看" v-model="casualViewModuleStatus" />
                <el-checkbox label="历史记录" v-model="historyModuleStatus" />
                <el-checkbox label="热门资源" v-model="hotModuleStatus" />
              </div>
            </div>
          </section>
        </div>
        <div class="modern-actions">
          <el-button type="primary" @click="saveConfig">保存</el-button>
        </div>
      </div>

      <div v-else class="cover-adjuster-container">
        <div class="cover-adjuster-content">
          <div class="resources-mode-selector">
            <el-radio-group v-model="resourcesShowMode">
              <el-radio v-for="item, index in resourcesShowModeList" :key="index" :label="item">{{
                $t(`resourcesShowMode.${item}`) }}</el-radio>
            </el-radio-group>
          </div>
          <div class="settings-panel">
            <div class="size-settings">
              <div class="setting-item">
                <el-checkbox v-model="coverPosterWidthStatus" label="锁定宽度" />
                <el-slider v-model="coverPosterWidthBase" :min="10" :max="1000" size="small" show-input />
              </div>
              <div class="setting-item">
                <el-checkbox v-model="coverPosterHeightStatus" label="锁定高度" />
                <el-slider v-model="coverPosterHeightBase" :min="10" :max="1000" size="small" show-input />
              </div>
              <div class="setting-item">
                <span class="setting-label">资源间距</span>
                <el-slider v-model="coverPosterGap" :min="0" :max="50" :step="0.1" size="small" show-input />
              </div>
              <div class="setting-item">
                <span class="setting-label">左右空距</span>
                <el-slider v-model="contentPadding" :min="0" :max="50" size="small" show-input />
              </div>
              <div class="setting-item">
                <span class="setting-label">封面标签大小</span>
                <el-slider v-model="coverDisplayTagFontSize" :min="8" :max="24" size="small" show-input />
              </div>
              <div class="setting-item">
                <span class="setting-label">封面图填充方式</span>
                <el-radio-group v-model="coverImageFit" size="small">
                  <el-radio-button v-for="item, index in dataset.coverImageFit" :key="index" :label="item"
                    :value="item" />
                </el-radio-group>
              </div>
              <div class="setting-item">
                <span class="setting-label">资源对齐方式</span>
                <el-radio-group v-model="resourceJustifyContent" size="small">
                  <el-radio-button label="start" value="flex-start" />
                  <el-radio-button label="center " value="center" />
                  <el-radio-button label="end" value="flex-end" />
                  <el-radio-button label="between" value="space-between" />
                  <el-radio-button label="around" value="space-around" />
                </el-radio-group>
              </div>
              <div class="setting-item">
                <el-checkbox v-model="showVideoDuration" label="显示视频时长" />
              </div>
            </div>
            <div class="other-settings">
              <div class="setting-item">
                <span class="setting-label">分页显示数</span>
                <el-input-number v-model="pageLimit" />
              </div>
              <div class="setting-item">
                <span class="setting-label">左侧边栏显示模式</span>
                <selectLeftColumnMode v-model="leftColumnMode" :teleported="false" style="width:160px;" />
              </div>
              <div class="setting-item">
                <el-checkbox v-model="leftColumnFloatAutoHide" label="浮动模式自动隐藏" />
              </div>
              <div class="setting-item">
                <el-checkbox v-model="showCustomTagResourceCount" label="显示自定义标签资源数量" />
              </div>
              <div class="setting-item">
                <span class="setting-label">详情显示模式</span>
                <selectResourceDetailsShowMode v-model="resourceDetailsShowMode" :teleported="false"
                  style="width:160px;" />
              </div>
              <div class="setting-item">
                <span class="setting-label">封面标题对齐方式</span>
                <el-radio-group v-model="coverTitleAlign" size="small">
                  <el-radio-button label="左对齐" value="left" />
                  <el-radio-button label="居中" value="center" />
                  <el-radio-button label="右对齐" value="right" />
                </el-radio-group>
              </div>
              <div class="setting-item">
                <el-checkbox label="随便看看" v-model="casualViewModuleStatus" />
                <el-checkbox label="历史记录" v-model="historyModuleStatus" />
                <el-checkbox label="热门资源" v-model="hotModuleStatus" />
              </div>
            </div>
          </div>
        </div>
        <div class="actions">
          <el-button type="primary" plain @click="saveConfig">保存</el-button>
        </div>
      </div>
    </el-popover>
  </div>
</template>
<script setup lang="ts">
import { appStoreData } from '@/storeData/app.storeData';
import { computed } from 'vue';
import dataset from '@/assets/dataset';
import { debounceNow } from '@/assets/debounce';
import { filesBasesServer } from '@/server/filesBases.server';
import { ElMessage } from 'element-plus';
import selectLeftColumnMode from '@/components/com/form/selectLeftColumnMode.vue';
import selectResourceDetailsShowMode from '@/components/com/form/selectResourceDetailsShowMode.vue';
import { Grid, Menu, Operation, Picture, SetUp } from '@element-plus/icons-vue';
import { useTheme } from '@/common/theme';

const store = {
  appStoreData: appStoreData(),
}

const resourcesShowModeList = dataset.resourcesShowMode;
const isModernAppearance = computed(() =>
  (store.appStoreData.appConfig.appearanceStyle || store.appStoreData.appConfig.headerStyle || 'modern') === 'modern'
);
const { isBrightTheme } = useTheme();

const resourcesShowMode = computed({
  get: () => store.appStoreData.currentConfigApp.resourcesShowMode,
  set: (value) => {
    store.appStoreData.currentConfigApp.resourcesShowMode = value;
  }
});
const coverPosterWidthStatus = computed({
  get: () => store.appStoreData.currentConfigApp.coverPosterWidthStatus,
  set: (value) => {
    store.appStoreData.currentConfigApp.coverPosterWidthStatus = value;
  }
});

const coverPosterWidthBase = computed({
  get: () => store.appStoreData.currentConfigApp.coverPosterWidthBase,
  set: (value) => {
    store.appStoreData.currentConfigApp.coverPosterWidthBase = value;
  }
});

const coverPosterHeightStatus = computed({
  get: () => store.appStoreData.currentConfigApp.coverPosterHeightStatus,
  set: (value) => {
    store.appStoreData.currentConfigApp.coverPosterHeightStatus = value;
  }
});

const coverPosterHeightBase = computed({
  get: () => store.appStoreData.currentConfigApp.coverPosterHeightBase,
  set: (value) => {
    store.appStoreData.currentConfigApp.coverPosterHeightBase = value;
  }
});


const coverPosterGap = computed({
  get: () => store.appStoreData.currentConfigApp.coverPosterGap,
  set: (value) => {
    store.appStoreData.currentConfigApp.coverPosterGap = value;
  }
});

const contentPadding = computed({
  get: () => store.appStoreData.currentConfigApp.contentPadding,
  set: (value) => {
    store.appStoreData.currentConfigApp.contentPadding = value;
  }
});
const coverDisplayTagFontSize = computed({
  get: () => store.appStoreData.currentConfigApp.coverDisplayTagFontSize,
  set: (value) => {
    store.appStoreData.currentConfigApp.coverDisplayTagFontSize = value;
  }
});

const coverImageFit = computed({
  get: () => {
    if (store.appStoreData.currentConfigApp.coverImageFit) {
      return store.appStoreData.currentConfigApp.coverImageFit;
    }

    return 'cover';
  },
  set: (value) => {
    store.appStoreData.currentConfigApp.coverImageFit = value;
  }
});

const coverTitleAlign = computed({
  get: () => {
    if (store.appStoreData.currentConfigApp.coverTitleAlign) {
      return store.appStoreData.currentConfigApp.coverTitleAlign;
    }
    return 'left';
  },
  set: (value) => {
    store.appStoreData.currentConfigApp.coverTitleAlign = value;
  }
});
const resourceJustifyContent = computed({
  get: () => {
    if (store.appStoreData.currentConfigApp.resourceJustifyContent) {
      return store.appStoreData.currentConfigApp.resourceJustifyContent;
    }
    return 'flex-start';
  },
  set: (value) => {
    store.appStoreData.currentConfigApp.resourceJustifyContent = value;
  }
});

const pageLimit = computed({
  get: () => store.appStoreData.currentConfigApp.pageLimit,
  set: (value) => {
    store.appStoreData.currentConfigApp.pageLimit = value;
  }
});
const leftColumnMode = computed({
  get: () => store.appStoreData.currentConfigApp.leftColumnMode,
  set: (value) => {
    store.appStoreData.currentConfigApp.leftColumnMode = value;
  }
});
const leftColumnFloatAutoHide = computed({
  get: () => store.appStoreData.currentConfigApp.leftColumnFloatAutoHide,
  set: (value) => {
    store.appStoreData.currentConfigApp.leftColumnFloatAutoHide = value;
  }
});
const showCustomTagResourceCount = computed({
  get: () => store.appStoreData.currentConfigApp.showCustomTagResourceCount !== false,
  set: (value) => {
    store.appStoreData.currentConfigApp.showCustomTagResourceCount = value;
    if (value) {
      void store.appStoreData.refreshCurrentTagData();
    }
  }
});
const resourceDetailsShowMode = computed({
  get: () => store.appStoreData.currentConfigApp.resourceDetailsShowMode,
  set: (value) => {
    store.appStoreData.currentConfigApp.resourceDetailsShowMode = value;
  }
});

const casualViewModuleStatus = computed({
  get: () => store.appStoreData.currentConfigApp.casualViewModule,
  set: (value) => {
    store.appStoreData.currentConfigApp.casualViewModule = value;
  }
});
const historyModuleStatus = computed({
  get: () => store.appStoreData.currentConfigApp.historyModule,
  set: (value) => {
    store.appStoreData.currentConfigApp.historyModule = value;
  }
});
const hotModuleStatus = computed({
  get: () => store.appStoreData.currentConfigApp.hotModule,
  set: (value) => {
    store.appStoreData.currentConfigApp.hotModule = value;
  }
});

const showVideoDuration = computed({
  get: () => store.appStoreData.currentConfigApp.showVideoDuration,
  set: (value) => {
    store.appStoreData.currentConfigApp.showVideoDuration = value;
  }
});

const saveConfig = debounceNow(async () => {
  const result = await filesBasesServer.setFilesBasesConfigById(store.appStoreData.currentFilesBases.id, store.appStoreData.currentConfigApp);
  if (!result.status) {
    ElMessage.error(result.msg);
    return;
  } else {
    ElMessage.success('保存成功');
  }
})

</script>
<style lang="scss" scoped>
.cover-adjuster-content {
  display: flex;
  gap: 20px;
  margin-bottom: 20px;

  .resources-mode-selector {
    width: 22%;
    flex-shrink: 0;
  }

  .settings-panel {
    flex: 1;
    display: flex;
    gap: 20px;

    .size-settings {
      flex: 1;
      display: flex;
      flex-direction: column;
      gap: 5px;
    }

    .other-settings {
      width: 30%;
      flex-shrink: 0;
      display: flex;
      flex-direction: column;
      gap: 10px;
    }

    .setting-item {
      display: flex;
      flex-direction: column;

      .el-checkbox {
        height: 18px;
      }
    }

    .setting-label {
      font-size: 14px;
      line-height: 24px;
    }
  }

  .el-slider {
    width: 100%;

    :deep(.el-slider__runway.show-input) {
      margin-right: 12px;
    }

    :deep(.el-input-number) {
      width: 110px;
    }
  }
}

.actions {
  display: flex;
  justify-content: flex-end;
}

.cover-adjuster-container--modern {
  --adjuster-bg: #181a1e;
  --adjuster-panel: #202226;
  --adjuster-panel-hover: #272a2f;
  --adjuster-border: rgba(255, 255, 255, 0.1);
  --adjuster-text: #f0f2f5;
  --adjuster-muted: #a8abb2;
  --adjuster-accent: #20aaa9;
  --adjuster-accent-soft: rgba(32, 170, 169, 0.16);
  overflow: hidden;
  color: var(--adjuster-text);
  background: var(--adjuster-bg);
  border: 1px solid var(--adjuster-border);
  border-radius: 12px;

  &.cover-adjuster-container--bright {
    --adjuster-bg: #f5f7f9;
    --adjuster-panel: #ffffff;
    --adjuster-panel-hover: #f4f7f8;
    --adjuster-border: #dfe5e9;
    --adjuster-text: #27343d;
    --adjuster-muted: #687680;
    --adjuster-accent: #159a9a;
    --adjuster-accent-soft: rgba(21, 154, 154, 0.11);
    box-shadow: 0 18px 48px rgba(31, 45, 61, 0.15);
  }

  .modern-adjuster-grid {
    display: grid;
    grid-template-columns: 270px minmax(0, 1fr) 270px;
    gap: 12px;
    padding: 12px;
  }

  .modern-main-settings {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .modern-settings-card {
    min-width: 0;
    padding: 14px;
    box-sizing: border-box;
    color: var(--adjuster-text);
    background: var(--adjuster-panel);
    border: 1px solid var(--adjuster-border);
    border-radius: 10px;
  }

  .modern-section-title {
    height: 28px;
    margin-bottom: 10px;
    display: flex;
    align-items: center;
    gap: 8px;
    color: var(--adjuster-text);
    font-size: 16px;
    font-weight: 650;

    .el-icon {
      color: var(--adjuster-muted);
      font-size: 17px;
    }
  }

  .modern-mode-card {
    padding: 10px;
  }

  .modern-mode-list {
    overflow: hidden;
    border: 1px solid var(--adjuster-border);
    border-radius: 8px;
  }

  .modern-mode-item {
    width: 100%;
    height: 36px;
    padding: 0 11px;
    display: flex;
    align-items: center;
    gap: 9px;
    border: 0;
    border-bottom: 1px solid var(--adjuster-border);
    color: var(--adjuster-text);
    background: transparent;
    cursor: pointer;
    text-align: left;
    transition: color 0.16s ease, background-color 0.16s ease;

    &:last-child {
      border-bottom: 0;
    }

    &:hover {
      color: var(--adjuster-accent);
      background: var(--adjuster-panel-hover);
    }

    &.active {
      color: var(--adjuster-accent);
      background: var(--adjuster-accent-soft);
      box-shadow: inset 3px 0 0 var(--adjuster-accent);
      font-weight: 600;
    }

    .el-icon {
      flex-shrink: 0;
      font-size: 16px;
    }

    span {
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }

  .modern-slider-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .modern-slider-item {
    min-width: 0;
    min-height: 42px;
    display: grid;
    grid-template-columns: 108px minmax(0, 1fr);
    align-items: center;
    gap: 10px;

    > .el-checkbox,
    > .modern-control-label {
      min-width: 0;
    }

    :deep(.el-slider) {
      min-width: 0;
    }

    :deep(.el-slider__runway.show-input) {
      margin-right: 12px;
    }

    :deep(.el-input-number) {
      width: 104px;
    }
  }

  .modern-control-label {
    display: block;
    color: var(--adjuster-text);
    font-size: 13px;
    line-height: 22px;
  }

  .modern-option-item {
    min-width: 0;
    margin-bottom: 11px;

    > .modern-control-label {
      margin-bottom: 5px;
    }

    > .el-radio-group {
      width: 100%;
      display: flex;

      :deep(.el-radio-button) {
        min-width: 0;
        flex: 1;
      }

      :deep(.el-radio-button__inner) {
        width: 100%;
        padding-left: 8px;
        padding-right: 8px;
      }
    }
  }

  .modern-page-settings {
    display: flex;
    flex-direction: column;
    gap: 3px;

    .modern-option-item {
      margin-bottom: 9px;
    }

    :deep(.el-input-number),
    :deep(.el-select) {
      width: 100%;
    }
  }

  .modern-module-options {
    margin-top: 2px;
    display: flex;
    flex-direction: column;
    align-items: flex-start;

    .el-checkbox {
      height: 27px;
    }
  }

  :deep(.el-checkbox__label) {
    color: var(--adjuster-text);
  }

  :deep(.el-checkbox__input.is-checked + .el-checkbox__label) {
    color: var(--adjuster-accent);
  }

  :deep(.el-checkbox__input.is-checked .el-checkbox__inner),
  :deep(.el-radio-button__original-radio:checked + .el-radio-button__inner) {
    background-color: var(--adjuster-accent);
    border-color: var(--adjuster-accent);
  }

  :deep(.el-slider) {
    --el-slider-main-bg-color: var(--adjuster-accent);
  }

  :deep(.el-radio-button__original-radio:checked + .el-radio-button__inner) {
    color: #ffffff;
    box-shadow: -1px 0 0 0 var(--adjuster-accent);
  }

  .modern-actions {
    padding: 10px 14px;
    display: flex;
    justify-content: flex-end;
    border-top: 1px solid var(--adjuster-border);
    background: var(--adjuster-panel);

    .el-button {
      min-width: 92px;
      --el-button-bg-color: var(--adjuster-accent);
      --el-button-border-color: var(--adjuster-accent);
      --el-button-hover-bg-color: #22b8b6;
      --el-button-hover-border-color: #22b8b6;
    }
  }
}

:global(.cover-adjuster-popover-modern.el-popover.el-popper) {
  max-width: calc(100vw - 28px);
  padding: 0 !important;
  overflow: hidden;
  border: 0 !important;
  border-radius: 12px !important;
  background: transparent !important;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.38) !important;
}

:global(.cover-adjuster-popover-modern .el-popper__arrow) {
  display: none;
}
</style>
