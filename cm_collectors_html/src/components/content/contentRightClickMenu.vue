<template>
  <div class="content-right-click-menu">
    <contentMenu :title="props.resource.title" :items="contentMenuItems">
      <slot></slot>
    </contentMenu>
  </div>
</template>
<script lang="ts" setup>
import { playOpenResourceFolder, playResource } from '@/common/play'
import { resourceDelete } from '@/common/resource'
import contentMenu from '@/components/com/tool/contentMenu/contentMenu.vue'
import type { I_resource } from '@/dataType/resource.dataType'
import type { PropType } from 'vue'
import { eventBus } from "@/main";
const props = defineProps({
  resource: {
    type: Object as PropType<I_resource>,
    required: true,
  },
})


// 基本菜单项
const contentMenuItems = [
  {
    label: '播放',
    icon: 'VideoPlay',
    handler: () => {
      playResource(props.resource)
    }
  },
  {
    label: '打开文件夹',
    icon: 'Folder',
    handler: () => {
      playOpenResourceFolder(props.resource.id)
    }
  },
  {
    separator: true
  },
  {
    label: '编辑',
    icon: 'Edit',
    handler: () => {
      eventBus.emit('edit-resource', { resource: props.resource });
    }
  },
  {
    label: '删除',
    icon: 'Delete',
    handler: () => {
      resourceDelete(props.resource, () => {
        eventBus.emit('delete-resource-success');
      })
    }
  },
]
</script>
