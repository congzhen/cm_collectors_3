<template>
  <div class="content-right-click-menu" :class="{ 'resource-selectable': selection?.active.value, 'resource-selected': isSelected }"
    :tabindex="selection?.active.value ? 0 : undefined" :role="selection?.active.value ? 'checkbox' : undefined"
    :aria-checked="selection?.active.value ? isSelected : undefined"
    :aria-label="selection?.active.value ? `选择资源：${props.resource.title}` : undefined"
    @click.capture="handleSelectionClick" @keydown.capture="handleSelectionKey" @contextmenu.capture="handleSelectionContextMenu">
    <span v-if="selection?.active.value" class="resource-selection-check" aria-hidden="true">
      <el-icon v-if="isSelected"><Check /></el-icon>
      <span v-if="isSelected">已选中</span>
    </span>
    <contentMenu :title="props.resource.title" :items="contentMenuItems_C">
      <slot></slot>
    </contentMenu>
  </div>
</template>
<script lang="ts" setup>
import { Check } from '@element-plus/icons-vue';
import { playOpenResourceFolder, playResource } from '@/common/play'
import { resourceDelete } from '@/common/resource'
import contentMenu from '@/components/com/tool/rightMenu/contentMenu.vue'
import type { I_resource } from '@/dataType/resource.dataType'
import { computed, inject, type PropType, type Ref } from 'vue'
import { eventBus } from "@/main";
import { appStoreData } from '@/storeData/app.storeData';
import { playListAdd } from '@/common/playList'
import { tvboxRecommendServer } from '@/server/tvboxRecommend.server'
import { ElMessage } from 'element-plus'
import { ElMessageBox } from 'element-plus'
import { videoTranscodeServer } from '@/server/videoTranscode.server'
import { resourceServer } from '@/server/resource.server'
import { searchStoreData } from '@/storeData/search.storeData'
import { E_searchSort } from '@/dataType/search.dataType'
import { resourceSelectionKey } from './resourceSelection';
const selection = inject(resourceSelectionKey, undefined);
const isSelected = computed(() => !!selection?.active.value && selection.selectedIds.value.has(props.resource.id));
const handleSelectionClick = (event: MouseEvent) => {
  if (!selection?.active.value) return;
  event.preventDefault();
  event.stopPropagation();
  selection.toggle(props.resource);
};
const handleSelectionKey = (event: KeyboardEvent) => {
  if (!selection?.active.value || ![' ', 'Enter'].includes(event.key)) return;
  event.preventDefault();
  event.stopPropagation();
  if (!event.repeat) selection.toggle(props.resource);
};
const handleSelectionContextMenu = (event: MouseEvent) => {
  if (!selection?.active.value) return;
  event.preventDefault();
  event.stopPropagation();
};
const store = {
  appStoreData: appStoreData(),
  searchStoreData: searchStoreData(),
}
const props = defineProps({
  resource: {
    type: Object as PropType<I_resource>,
    required: true,
  },
})

const resourceListPage = inject<Ref<I_resource[]> | undefined>('resource-list-page', undefined)
const resourceIndex_C = computed(() => resourceListPage?.value.findIndex(item => item.id === props.resource.id) ?? -1)
const isAddTimeSort_C = computed(() => [E_searchSort.AddTimeAsc, E_searchSort.AddTimeDesc]
  .includes(store.searchStoreData.searchData.sort))
const hasResourceFilter_C = computed(() => {
  const searchData = store.searchStoreData.searchData
  const fixedGroups = [
    searchData.country,
    searchData.definition,
    searchData.videoCodec,
    searchData.year,
    searchData.star,
    searchData.performer,
    searchData.cup,
  ]
  return searchData.searchTextSlc.length > 0
    || fixedGroups.some(group => group.options.length > 0)
    || Object.values(searchData.tag).some(group => group.options.length > 0)
})
const canShowSwapAddTime_C = computed(() => !!resourceListPage
  && isAddTimeSort_C.value
  && !hasResourceFilter_C.value
  && props.resource.pinToTop === 0)
const previousResource_C = computed(() => {
  if (!canShowSwapAddTime_C.value || resourceIndex_C.value <= 0) return undefined
  const item = resourceListPage?.value[resourceIndex_C.value - 1]
  return item?.pinToTop === 0 ? item : undefined
})
const nextResource_C = computed(() => {
  if (!canShowSwapAddTime_C.value || resourceIndex_C.value < 0) return undefined
  const item = resourceListPage?.value[resourceIndex_C.value + 1]
  return item?.pinToTop === 0 ? item : undefined
})

const swapAddTimeHandle = async (otherResource: I_resource | undefined) => {
  if (!otherResource) return
  try {
    const result = await resourceServer.swapAddTime(props.resource.id, otherResource.id)
    if (result?.status) {
      ElMessage.success('交换成功')
      eventBus.emit('resource-add-time-exchanged')
    } else {
      ElMessage.error(result?.msg || '交换添加时间失败')
    }
  } catch {
    ElMessage.error('交换添加时间失败')
  }
}


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
      ...(canShowSwapAddTime_C.value ? [{
        label: '与上项换时间',
        icon: 'Top',
        disabled: !previousResource_C.value,
        handler: () => swapAddTimeHandle(previousResource_C.value),
      }, {
        label: '与下项换时间',
        icon: 'Bottom',
        disabled: !nextResource_C.value,
        handler: () => swapAddTimeHandle(nextResource_C.value),
      }, {
        separator: true,
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } as any] : []),
      {
        label: '编辑',
        icon: 'Edit',
        handler: () => {
          eventBus.emit('edit-resource', { resource: props.resource });
        }
      },
      ...(props.resource.pinToTop == 0
        ? [{
          label: '置顶',
          icon: 'Top',
          handler: () => {
            eventBus.emit('edit-resource-pinToTop', { resource: props.resource, pinToTopStatus: true });
          }
        }]
        : [{
          label: '再次置顶',
          icon: 'Top',
          handler: () => {
            eventBus.emit('edit-resource-pinToTop', { resource: props.resource, pinToTopStatus: true });
          }
        }, {
          label: '取消置顶',
          icon: 'Bottom',
          handler: () => {
            eventBus.emit('edit-resource-pinToTop', { resource: props.resource, pinToTopStatus: false });
          }
        }]
      ),
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
        label: '加入转码与剪辑列表',
        icon: 'VideoCamera',
        handler: async () => {
          try {
            const count = props.resource.dramaSeries?.length || 0
            if (count === 0) {
              ElMessage.warning('当前资源没有可加入的视频')
              return
            }
            await ElMessageBox.confirm(
              `将资源下的 ${count} 个视频加入转码与剪辑列表，系统会自动跳过重复项和不存在的文件。`,
              '加入转码与剪辑列表',
              { type: 'warning' },
            )
            const result = await videoTranscodeServer.add({ resourceIds: [props.resource.id] })
            if (result.status) {
              ElMessage.success(`已加入 ${result.data.added} 个，跳过 ${result.data.skippedDuplicate + result.data.skippedMissing} 个`)
            } else {
              ElMessage.error(result.msg || '加入失败')
            }
          } catch {
            // 用户取消确认。
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

<style scoped lang="scss">
.resource-selectable {
  position: relative;
  isolation: isolate;
  cursor: pointer;
  border-radius: 6px;
  &::after {
    content: ''; position: absolute; inset: 0; border-radius: inherit;
    border: 1px solid rgba(128, 140, 152, .2); pointer-events: none; z-index: 20;
    transition: background-color .18s ease, border-color .18s ease, box-shadow .18s ease;
  }
  &:hover::after, &:focus-visible::after { border-color: color-mix(in srgb, var(--el-color-primary) 65%, transparent); }
  &:focus-visible { outline: 2px solid var(--el-color-primary); outline-offset: 3px; }
  :deep(.play-icon), :deep(.compact-play), :deep(.cinema-gallery-play) { visibility: hidden; }
  :deep(.content-cover .el-image) { scale: 1 !important; }
}
.resource-selectable.resource-selected::after {
  border: 2px solid var(--el-color-primary);
  background: rgba(16, 20, 24, .42);
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--el-color-primary) 16%, transparent);
}
.resource-selection-check {
  position: absolute; top: 8px; right: 8px; width: 24px; height: 24px;
  display: flex; align-items: center; justify-content: center; gap: 6px;
  box-sizing: border-box; border-radius: 6px;
  border: 1px solid rgba(255,255,255,.5); background: rgba(18,24,30,.58);
  color: #eef3f7; z-index: 21; pointer-events: none;
  box-shadow: 0 2px 6px rgba(0,0,0,.15);
  backdrop-filter: blur(8px);
  font-size: 12px; font-weight: 500; white-space: nowrap;
  transition: background-color .18s ease, border-color .18s ease;
}
.resource-selected .resource-selection-check {
  width: auto; height: 28px; padding: 0 8px 0 5px;
  background: rgba(18,24,30,.95); border-color: rgba(255,255,255,.35);
  .el-icon {
    width: 18px; height: 18px; border-radius: 4px;
    background: var(--el-color-primary); color: #fff; font-size: 13px;
  }
}
</style>
