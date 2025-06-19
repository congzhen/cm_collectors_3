<template>
  <drawerCommon ref="drawerCommonRef" width="800px" title="标签" :btnSubmit="false">
    <tagList></tagList>
    <template #footerBtn>
      <div>
        <el-button @click="tagClassFormHandle('add', '', '')">添加标签分类</el-button>
      </div>
    </template>
  </drawerCommon>
  <tagFormClassDialog ref="tagFormClassDialogRef"></tagFormClassDialog>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import drawerCommon from '@/components/com/dialog/drawer.common.vue';
import tagList from './tagList.vue';
import tagFormClassDialog from './tagFormClassDialog.vue'

const drawerCommonRef = ref<InstanceType<typeof drawerCommon>>()
const tagFormClassDialogRef = ref<InstanceType<typeof tagFormClassDialog>>();
const tagClassFormHandle = (mode: 'add' | 'edit', id: string, name: string) => {
  tagFormClassDialogRef.value?.open(mode, id, name)
}

const open = () => {
  drawerCommonRef.value?.open()
}
const close = () => {
  drawerCommonRef.value?.close()
}
// eslint-disable-next-line no-undef
defineExpose({ open, close })
</script>
