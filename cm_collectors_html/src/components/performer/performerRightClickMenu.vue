<template>
  <div class="content-right-click-menu">
    <performerMenu :title="props.performer.name" :items="contentMenuItems_C">
      <slot></slot>
    </performerMenu>
  </div>
</template>
<script lang="ts" setup>
import performerMenu from '@/components/com/tool/rightMenu/performerMenu.vue'
import { computed, type PropType } from 'vue'
import { appStoreData } from '@/storeData/app.storeData';
import type { I_performer } from '@/dataType/performer.dataType'
const store = {
  appStoreData: appStoreData(),
}
const props = defineProps({
  performer: {
    type: Object as PropType<I_performer>,
    required: true,
  },
})
const emits = defineEmits(['search', 'edit', 'migrate', 'delete'])

const contentMenuItems_C = computed(() => {
  // 基本菜单项
  const contentMenuItems = [
    {
      label: '检索演员资源',
      icon: 'VideoCameraFilled',
      handler: () => {
        emits('search', props.performer)
      }
    },
  ]
  if (store.appStoreData.displayAdminFn) {
    contentMenuItems.push(...[

      {
        label: '编辑',
        icon: 'Edit',
        handler: () => {
          emits('edit', props.performer)
        }
      },
      {
        label: '迁移数据集',
        icon: 'Switch',
        handler: () => {
          emits('migrate', props.performer)
        }
      },
      {
        label: '删除 (移至回收站)',
        icon: 'Delete',
        handler: () => {
          emits('delete', props.performer)
        }
      },
    ])
  }
  return contentMenuItems
})

</script>
