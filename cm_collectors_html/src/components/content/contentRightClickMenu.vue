<template>
  <div class="content-right-click-menu">
    <contentMenu :title="props.resource.title" :items="contentMenuItems_C">
      <slot></slot>
    </contentMenu>
  </div>
</template>
<script lang="ts" setup>
import { playOpenResourceFolder, playResource } from '@/common/play'
import { resourceDelete } from '@/common/resource'
import contentMenu from '@/components/com/tool/rightMenu/contentMenu.vue'
import type { I_resource } from '@/dataType/resource.dataType'
import { computed, type PropType } from 'vue'
import { eventBus } from "@/main";
import { appStoreData } from '@/storeData/app.storeData';
import { playListAdd } from '@/common/playList'
import { tvboxRecommendServer } from '@/server/tvboxRecommend.server'
import { ElMessage } from 'element-plus'
const store = {
  appStoreData: appStoreData(),
}
const props = defineProps({
  resource: {
    type: Object as PropType<I_resource>,
    required: true,
  },
})


const contentMenuItems_C = computed(() => {
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
      label: '加入播放列表',
      icon: 'Memo',
      handler: () => {
        playListAdd(props.resource.id)
      }
    },
  ]
  if (store.appStoreData.displayAdminFn) {
    contentMenuItems.push(...[
      {
        label: '打开文件夹',
        icon: 'Folder',
        handler: () => {
          playOpenResourceFolder(props.resource.id)
        }
      },
      {
        separator: true
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any,
      {
        label: '编辑',
        icon: 'Edit',
        handler: () => {
          eventBus.emit('edit-resource', { resource: props.resource });
        }
      },
      {
        label: props.resource.pinToTop == 0 ? '置顶' : '取消置顶',
        icon: 'Top',
        handler: () => {
          eventBus.emit('edit-resource-pinToTop', { resource: props.resource });
        }
      },
      {
        label: '打标签',
        icon: 'PriceTag',
        handler: () => {
          eventBus.emit('edit-resource-tag', { resource: props.resource });
        }
      },
      {
        label: '添加演员',
        icon: 'User',
        handler: () => {
          eventBus.emit('edit-resource-performer', { resource: props.resource });
        }
      },
      {
        label: '添加到TVBox推荐',
        icon: 'Star',
        handler: async () => {
          const result = await tvboxRecommendServer.add(props.resource.id)
          if (result.status) {
            ElMessage.success('已添加到TVBox推荐')
          } else {
            ElMessage.error(result.msg || '添加失败')
          }
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
    ])
  }
  return contentMenuItems
})

</script>
