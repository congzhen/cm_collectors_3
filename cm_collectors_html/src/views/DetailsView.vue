<template>
  <detailsShowRight v-if="store.appStoreData.currentConfigApp.resourceDetailsShowMode == 'right'"
    :resource="props.resource" @update-resouce-success="updateResourceSuccessHandle"
    @delete-resource-success="deleteResourceSuccessHandle">
  </detailsShowRight>
  <detailsShowDialog ref="detailsShowDialogRef"
    v-else-if="store.appStoreData.currentConfigApp.resourceDetailsShowMode == 'popup'" :resource="props.resource"
    @update-resouce-success="updateResourceSuccessHandle" @delete-resource-success="deleteResourceSuccessHandle">
  </detailsShowDialog>
</template>
<script lang="ts" setup>
import { ref, type PropType } from 'vue'
import type { I_resource } from '@/dataType/resource.dataType'
import { appStoreData } from '@/storeData/app.storeData';
import detailsShowRight from '@/components/details/detailsShowRight.vue';
import detailsShowDialog from '@/components/details/detailsShowDialog.vue';
const store = {
  appStoreData: appStoreData(),
}
const props = defineProps({
  resource: {
    type: Object as PropType<I_resource> | undefined,
    default: undefined
  },
})
const emits = defineEmits(['updateResouceSuccess', 'deleteResourceSuccess'])

const detailsShowDialogRef = ref<InstanceType<typeof detailsShowDialog>>();

const init = () => {
  if (store.appStoreData.currentConfigApp.resourceDetailsShowMode == 'popup') {
    detailsShowDialogRef.value?.open();
  }
}


const updateResourceSuccessHandle = (data: I_resource) => {
  emits('updateResouceSuccess', data)
}

const deleteResourceSuccessHandle = () => {
  emits('deleteResourceSuccess')
}
defineExpose({ init });
</script>
<style lang="scss" scoped>
.details-view-k {
  width: 278px;
  height: 100%;
  padding-left: 2px;
}

.details-view {
  width: 100%;
  height: 100%;
  color: #a8abb2;
  display: flex;
  flex-direction: column;

  .content-cover {
    width: 100%;
    flex-shrink: 0;
    overflow: hidden;

    .el-image {
      width: 100%;
    }
  }

  .tool {
    width: 100%;
    flex-shrink: 0;
    padding-bottom: 5px;

    .el-button-group {
      width: 100%;
      display: flex;
      justify-content: center;

      .el-button {
        width: 25%;
      }
    }
  }

  .details-container {
    flex-grow: 1;
    overflow: hidden;
  }

  .title {
    font-size: 1.1em;
    font-weight: 500;
    line-height: 1.2em;
    color: #ffffff;
  }

  .info-base {
    font-size: 12px;
    line-height: 1.5em;
    padding: 5px 0;

    :deep(.el-breadcrumb) {
      .el-breadcrumb__inner {
        font-size: 12px;
        color: #a8abb2;
      }
    }

    .info-base-rate {
      .el-rate {
        height: 16px;
      }
    }
  }

  .info-block {
    padding: 5px 0;

    .el-alert {
      padding: 4px 8px;
    }

    .performer-list {
      display: flex;
      flex-wrap: wrap;
      gap: 4px;

      .performer-item {
        width: 32%;
        overflow: hidden;
      }
    }

    .tag-list {
      padding: 5px;
      display: flex;
      flex-wrap: wrap;
      gap: 5px;
    }

    .abstract {
      text-indent: 2em;
      padding: 10px;
    }
  }
}
</style>
