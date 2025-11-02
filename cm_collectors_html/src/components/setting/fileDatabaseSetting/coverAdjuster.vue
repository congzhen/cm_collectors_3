<template>
  <div class="cover-adjuster-wrapper">
    <el-popover placement="top" :width="800" trigger="click">
      <template #reference>
        <el-button icon="Monitor" size="small"></el-button>
      </template>
      <div class="cover-adjuster-container">
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
                <el-slider v-model="coverPosterWidthBase" :min="10" :max="1000" size="small" />
              </div>
              <div class="setting-item">
                <el-checkbox v-model="coverPosterHeightStatus" label="锁定高度" />
                <el-slider v-model="coverPosterHeightBase" :min="10" :max="1000" size="small" />
              </div>
              <div class="setting-item">
                <span class="setting-label">资源间距</span>
                <el-slider v-model="coverPosterGap" :min="0" :max="50" :step="0.1" size="small" />
              </div>
              <div class="setting-item">
                <span class="setting-label">左右空距</span>
                <el-slider v-model="contentPadding" :min="0" :max="50" size="small" />
              </div>
              <div class="setting-item">
                <span class="setting-label">封面图填充方式</span>
                <el-radio-group v-model="coverImageFit" size="small">
                  <el-radio-button v-for="item, index in dataset.coverImageFit" :key="index" :label="item"
                    :value="item" />
                </el-radio-group>
              </div>

            </div>
            <div class="other-settings">
              <div class="setting-item">
                <span class="setting-label">分页显示数</span>
                <el-input-number v-model="pageLimit" />
              </div>
              <div class="setting-item">
                <span class="setting-label">左侧边栏显示模式</span>
                <selectLeftColumnMode v-model="leftColumnMode" style="width:160px;" />
              </div>
              <div class="setting-item">
                <span class="setting-label">详情显示模式</span>
                <selectResourceDetailsShowMode v-model="resourceDetailsShowMode" style="width:160px;" />
              </div>
              <div class="setting-item">
                <span class="setting-label">封面标题对齐方式</span>
                <el-radio-group v-model="coverTitleAlign" size="small">
                  <el-radio-button label="左对齐" value="left" />
                  <el-radio-button label="居中" value="center" />
                  <el-radio-button label="右对齐" value="right" />
                </el-radio-group>
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

const store = {
  appStoreData: appStoreData(),
}

const resourcesShowModeList = dataset.resourcesShowMode;

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

const coverImageFit = computed({
  get: () => {
    console.log(store.appStoreData.currentConfigApp.coverImageFit);
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
const resourceDetailsShowMode = computed({
  get: () => store.appStoreData.currentConfigApp.resourceDetailsShowMode,
  set: (value) => {
    store.appStoreData.currentConfigApp.resourceDetailsShowMode = value;
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
    width: 26%;
    flex-shrink: 0;
  }

  .settings-panel {
    flex: 1;
    display: flex;
    gap: 20px;

    .size-settings,
    .other-settings {
      width: 60%;
      flex-shrink: 0;
      display: flex;
      flex-direction: column;
      gap: 15px;
    }

    .setting-item {
      display: flex;
      flex-direction: column;
      gap: 5px;
    }

    .setting-label {
      font-size: 14px;
    }
  }

  .el-slider {
    width: 100%;
  }
}

.actions {
  display: flex;
  justify-content: flex-end;
}
</style>
