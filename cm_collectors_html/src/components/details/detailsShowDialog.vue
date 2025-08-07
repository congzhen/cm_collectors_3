<template>
  <dialogCommon ref="dialogCommonRef" :title="title_C" :footer="false" width="1000px">
    <div class="details-dialog-container">
      <div class="left">
        <div class="content-cover">
          <el-image :src="getResourceCoverPoster(props.resource)" fit="contain" />
        </div>
        <div class="tool">
          <detailsBtn :resource="props.resource" @paly="close" @update-resouce-success="updateResourceSuccessHandle"
            @delete-resource-success="deleteResourceSuccessHandle"></detailsBtn>
        </div>
      </div>
      <div class="right">
        <detailsInfo :resource="props.resource"></detailsInfo>
      </div>
    </div>
  </dialogCommon>
</template>
<script lang="ts" setup>
import dialogCommon from '../com/dialog/dialog.common.vue';
import { computed, ref, type PropType } from 'vue'
import type { I_resource } from '@/dataType/resource.dataType'
import { getResourceCoverPoster } from '@/common/photo';
import detailsInfo from '@/components/details/detailsInfo.vue'
import detailsBtn from '@/components/details/detailsBtn.vue'

const props = defineProps({
  resource: {
    type: Object as PropType<I_resource> | undefined,
    default: undefined
  },
})
const emits = defineEmits(['updateResouceSuccess', 'deleteResourceSuccess']);

const dialogCommonRef = ref<InstanceType<typeof dialogCommon>>();

const title_C = computed(() => {
  return props.resource?.title || '资源详情';
})

const updateResourceSuccessHandle = (data: I_resource) => {
  emits('updateResouceSuccess', data)
}

const deleteResourceSuccessHandle = () => {
  emits('deleteResourceSuccess')
}

const open = () => {
  dialogCommonRef.value?.open();
}
const close = () => {
  dialogCommonRef.value?.close();
}
defineExpose({ open });
</script>
<style scoped lang="scss">
.details-dialog-container {
  width: 100%;
  height: 560px;
  display: flex;
  gap: 20px;

  .left {
    width: 60%;
    height: 100%;
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;

    .content-cover {
      flex: 1;
      overflow: hidden;

      .el-image {
        width: 100%;
        height: 100%;
      }
    }

    .tool {
      margin-top: 10px;
      flex-shrink: 0;

      :deep(.el-button) {
        width: auto;
      }
    }
  }

  .right {
    flex: 1;
  }
}
</style>
