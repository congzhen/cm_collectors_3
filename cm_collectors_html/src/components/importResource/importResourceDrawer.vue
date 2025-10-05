<template>
  <drawerCommon ref="drawerCommonRef" width="680px" title="资源导入与刮削" :btnSubmitTitle="submitTitle_C"
    @submit="submitHandle">
    <div class="import-resource-main">
      <div class="tool">
        <el-radio-group v-model="modeRadio" @change="changeModeRadioHandle">
          <el-radio-button label="导入" value="scanDisk" />
          <el-radio-button label="刮削" value="scraper" />
        </el-radio-group>
      </div>
      <div class="main">
        <modeScanDisk ref="modeScanDiskRef" v-if="modeRadio === 'scanDisk'" @success="successHandle"></modeScanDisk>
        <scraperData ref="scraperDataRef" v-if="modeRadio === 'scraper'" @success="successHandle"></scraperData>
      </div>
    </div>
  </drawerCommon>
</template>
<script lang="ts" setup>
import drawerCommon from '@/components/com/dialog/drawer-common.vue';
import modeScanDisk from './modeScanDisk.vue';
import scraperData from './scraperData.vue'
import { nextTick, ref, computed } from 'vue';

const emits = defineEmits(['success'])

const drawerCommonRef = ref<InstanceType<typeof drawerCommon>>();
const modeRadio = ref('scanDisk');
const modeScanDiskRef = ref<InstanceType<typeof modeScanDisk>>();
const scraperDataRef = ref<InstanceType<typeof scraperData>>();

const submitTitle_C = computed(() => {
  switch (modeRadio.value) {
    case 'scanDisk':
      return '导入';
    case 'scraper':
      return '刮削';
  }
  return '提交';
});

const init = () => {
  changeModeRadioHandle();
}

const changeModeRadioHandle = () => {
  nextTick(() => {
    switch (modeRadio.value) {
      case 'scanDisk':
        modeScanDiskRef.value?.init();
        break;
      case 'scraper':
        scraperDataRef.value?.init();
        break;
    }
  });
}
const submitHandle = () => {
  switch (modeRadio.value) {
    case 'scanDisk':
      modeScanDiskRef.value?.submit();
      break;
    case 'scraper':
      scraperDataRef.value?.submit();
      break;
  }
}

const successHandle = () => {
  colse();
  emits('success')
}

const open = async () => {
  drawerCommonRef.value?.open();
  init();
}

const colse = () => {
  drawerCommonRef.value?.close();
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
