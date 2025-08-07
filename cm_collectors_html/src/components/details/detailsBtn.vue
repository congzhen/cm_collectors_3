<template>
  <div class="tool" v-if="props.resource">
    <el-button-group>
      <el-button icon="VideoPlay" @click="playResourceHandle" />
      <el-button icon="Folder" @click="playOpenResourceFolder(props.resource.id)" />
      <el-button icon="Edit" @click="editResourceHandle" />
      <el-button icon="Delete" @click="resourceDeleteHandle" />
    </el-button-group>
    <resourceFormDrawer ref="resourceFormDrawerRef" @success="updateResourceSuccessHandle"></resourceFormDrawer>
  </div>
</template>
<script lang="ts" setup>
import type { I_resource } from '@/dataType/resource.dataType'
import { ref, type PropType } from 'vue'
import resourceFormDrawer from '@/components/resource/resourceFormDrawer.vue';
import { resourceDelete } from '@/common/resource';
import { playResource, playOpenResourceFolder } from '@/common/play'
const props = defineProps({
  resource: {
    type: Object as PropType<I_resource> | undefined,
    default: undefined
  },
})
const emits = defineEmits(['paly', 'updateResouceSuccess', 'deleteResourceSuccess'])

const resourceFormDrawerRef = ref<InstanceType<typeof resourceFormDrawer>>()

const playResourceHandle = () => {
  if (!props.resource) return
  playResource(props.resource)
  emits('paly')
}
const editResourceHandle = () => {
  resourceFormDrawerRef.value?.open('edit', props.resource)
}
const updateResourceSuccessHandle = (data: I_resource) => {
  emits('updateResouceSuccess', data)
}

const resourceDeleteHandle = () => {
  if (!props.resource) return
  resourceDelete(props.resource, () => {
    emits('deleteResourceSuccess')
  })
}
</script>
<style scoped lang="scss">
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
</style>
