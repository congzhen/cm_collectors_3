<template>
  <div class="tag-right-click-menu">
    <tagMenu :title="props.tag.name" :items="tagMenuItems_C">
      <slot></slot>
    </tagMenu>
  </div>
</template>
<script lang="ts" setup>
import tagMenu from '@/components/com/tool/rightMenu/tagMenu.vue'
import { computed, type PropType } from 'vue'
import { appStoreData } from '@/storeData/app.storeData';
import type { I_tag } from '@/dataType/tag.dataType';
const store = {
  appStoreData: appStoreData(),
}
const props = defineProps({
  tag: {
    type: Object as PropType<I_tag>,
    required: true,
  },
})
const emits = defineEmits(['edit', 'delete', 'disable', 'enable'])

const tagMenuItems_C = computed(() => {
  const tagMenuItems = [];
  // 基本菜单项
  if (store.appStoreData.displayAdminFn) {
    tagMenuItems.push(...[
      {
        label: '编辑',
        icon: 'Edit',
        handler: () => {
          emits('edit', props.tag)
        }
      },
      {
        label: '启用',
        icon: 'Open',
        handler: () => {
          emits('enable', props.tag)
        },
        disabled: props.tag.status,
      },
      {
        label: '禁用',
        icon: 'TurnOff',
        handler: () => {
          emits('disable', props.tag)
        },
        disabled: !props.tag.status,
      },
      {
        label: '删除',
        icon: 'Delete',
        handler: () => {
          emits('delete', props.tag)
        }
      },
    ])
  }
  return tagMenuItems
})

</script>
