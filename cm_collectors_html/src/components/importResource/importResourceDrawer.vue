<template>
  <drawerCommon ref="drawerCommonRef" width="680px" title="资源导入" btnSubmitTitle="导入" @submit="submitHandle">
    <div class="import-resource-main">
      <div class="tool">
        <el-radio-group v-model="modeRadio">
          <el-radio-button label="磁盘扫描" value="scanDisk" />
          <el-radio-button label="nfo导入" value="nfoImport" />
          <el-radio-button label="simple导入" value="simpleImport" />
        </el-radio-group>
      </div>
      <div class="main">
        <modeScanDisk ref="modeScanDiskRef" v-if="modeRadio === 'scanDisk'"></modeScanDisk>
      </div>
    </div>
  </drawerCommon>
</template>
<script lang="ts" setup>
import drawerCommon from '@/components/com/dialog/drawer-common.vue';
import modeScanDisk from './modeScanDisk.vue';
import { nextTick, ref } from 'vue';


const drawerCommonRef = ref<InstanceType<typeof drawerCommon>>();
const modeRadio = ref('scanDisk');
const modeScanDiskRef = ref<InstanceType<typeof modeScanDisk>>();

const init = () => {
  switch (modeRadio.value) {
    case 'scanDisk':
      modeScanDiskRef.value?.init();
      break;
    case 'nfoImport':
      break;
    case 'simpleImport':
      break;
  }
}

const submitHandle = () => {
  switch (modeRadio.value) {
    case 'scanDisk':
      modeScanDiskRef.value?.submit();
      break;
    case 'nfoImport':
      break;
    case 'simpleImport':
      break;
  }
}

const open = async () => {
  drawerCommonRef.value?.open();
  nextTick(() => {
    init();
  });
}
defineExpose({ open })
</script>
<style lang="scss" scoped>
.import-resource-main {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;

  .tool {
    flex-shrink: 0;
    height: 40px;
  }

  .main {
    flex: 1;
  }
}
</style>
