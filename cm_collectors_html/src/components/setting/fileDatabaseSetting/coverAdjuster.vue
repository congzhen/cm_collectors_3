<template>
  <div>
    <el-popover placement="top" :width="600" trigger="click">
      <template #reference>
        <el-button icon="Monitor" size="small"></el-button>
      </template>
      <div class=" coverAdjuster">
        <div class="select-resources-show-mode">
          <el-radio-group v-model="resourcesShowMode">
            <el-radio v-for="item, index in resourcesShowModeList" :key="index" :label="item">{{
              $t(`resourcesShowMode.${item}`) }}</el-radio>
          </el-radio-group>
        </div>
        <div class="set-value">
          <div>
            <el-checkbox v-model="coverPosterWidthStatus" label="锁定宽度" />
          </div>
          <el-slider v-model="coverPosterWidthBase" :min="10" :max="1000" />
          <div>
            <el-checkbox v-model="coverPosterHeightStatus" label="锁定高度" />
          </div>
          <el-slider v-model="coverPosterHeightBase" :min="10" :max="1000" />
          <div>
            间距
          </div>
          <el-slider v-model="coverPosterGap" :min="0" :max="50" step="0.1" />
          <div>
            分页显示数量
          </div>
          <el-input-number v-model="pageLimit" />
          <div>
            <el-button type="primary" plain @click="saveConfig">保存</el-button>
          </div>
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

const pageLimit = computed({
  get: () => store.appStoreData.currentConfigApp.pageLimit,
  set: (value) => {
    store.appStoreData.currentConfigApp.pageLimit = value;
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
.coverAdjuster {
  display: flex;
  gap: 5px;

  .select-resources-show-mode {
    width: 50%;
    flex-shrink: 0;
  }

  .set-value {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .el-slider {
    width: 100%;
  }
}
</style>
