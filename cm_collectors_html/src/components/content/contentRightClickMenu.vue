<template>
  <div ref="groupAnchor" class="content-right-click-menu" :class="{ 'resource-group-stack': isGroupStack, 'resource-group-composite': isGroupStack && groupCoverStyle !== 'stack', 'resource-selectable': selection?.active.value, 'resource-selected': isSelected }"
    :tabindex="selection?.active.value ? 0 : undefined" :role="selection?.active.value ? 'checkbox' : undefined"
    :aria-checked="selection?.active.value ? isSelected : undefined"
    :aria-label="selection?.active.value ? `选择资源：${props.resource.title}` : undefined"
    @pointerenter="openGroupFan" @pointermove="trackGroupPointer" @pointerleave="leaveGroupFan" @focusin="openGroupFan" @focusout="leaveGroupFan" @keydown.esc="closeGroupFan"
    @click.capture="handleSelectionClick" @keydown.capture="handleSelectionKey" @contextmenu.capture="handleSelectionContextMenu">
    <ResourceGroupMark v-if="selection?.active.value && props.resource.associationGroupId"
      class="resource-management-group" :group-id="props.resource.associationGroupId" />
    <span v-if="selection?.active.value" class="resource-selection-check" aria-hidden="true">
      <el-icon v-if="isSelected"><Check /></el-icon>
      <span v-if="isSelected">已选中</span>
    </span>
    <template v-if="isGroupStack && groupCoverStyle === 'stack'">
      <div v-for="(cover, index) in stackCovers" :key="index" class="resource-group-back" :class="`resource-group-back--${index}`" aria-hidden="true">
        <img v-if="cover" :src="cover" alt="" loading="lazy" decoding="async" />
      </div>
    </template>
    <contentMenu :title="props.resource.title" :items="contentMenuItems_C">
      <div v-if="isGroupStack" class="resource-group-front"><slot /></div>
      <slot v-else></slot>
    </contentMenu>
    <el-popover v-if="isGroupStack" :visible="groupVisible" :virtual-ref="(groupExpandStyle === 'spine' || groupExpandStyle === 'magazine') && groupPointerAnchor ? groupPointerAnchor : groupAnchor" virtual-triggering
      trigger="hover" :placement="(groupExpandStyle === 'spine' || groupExpandStyle === 'magazine') ? 'right' : 'top'" :width="groupFanWidth" :show-arrow="false" :show-after="250" :hide-after="250" :offset="groupFanOffset"
      :popper-options="{ modifiers: [{ name: 'preventOverflow', options: { altAxis: true, padding: 12 } }] }"
      :popper-style="{ maxWidth: 'calc(100vw - 16px)', padding: '0', border: '0', background: 'transparent', boxShadow: 'none' }" @after-leave="releaseGroupFanView">
      <div ref="groupSurface" @pointerenter="keepGroupFan" @pointerleave="leaveGroupFan" @focusin="keepGroupFan" @focusout="leaveGroupFan" @keydown.esc="closeGroupFan">
      <ResourceGroupPanel v-if="groupPanelReady" :files-bases-id="props.resource.filesBases_id" :group-id="props.resource.associationGroupId!"
        :source-resource="props.resource" :expand-style="groupExpandStyle" :display-active="groupViewActive" :expected-count="groupMemberCount" :matched-ids="groupContext?.pageIds.value" @select="selectGroupMember" />
      </div>
    </el-popover>
  </div>
</template>
<script lang="ts" setup>
import { Check } from '@element-plus/icons-vue';
import { playOpenResourceFolder, playResource } from '@/common/play'
import { resourceDelete } from '@/common/resource'
import contentMenu from '@/components/com/tool/rightMenu/contentMenu.vue'
import type { I_resource } from '@/dataType/resource.dataType'
import { computed, inject, onBeforeUnmount, ref, shallowRef, watch, type PropType, type Ref } from 'vue'
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
import { claimResourceGroupFan, releaseResourceGroupFan, resourceGroupContextKey, resourceGroupKeepOpenKey, resourceGroupSelectKey } from './resourceGroups';
import ResourceGroupPanel from './ResourceGroupPanel.vue';
import ResourceGroupMark from './ResourceGroupMark.vue';
import { getResourceCoverPoster } from '@/common/photo';
const props = defineProps({
  resource: {
    type: Object as PropType<I_resource>,
    required: true,
  },
})
const groupContext = inject(resourceGroupContextKey, undefined);
const groupAction = inject(resourceGroupSelectKey, playResource);
const keepOpenAfterSelect = inject(resourceGroupKeepOpenKey, false);
const groupAnchor = ref<HTMLElement>();
const groupPointerAnchor = shallowRef<{ getBoundingClientRect: () => DOMRect; contextElement?: HTMLElement }>();
const groupVisible = ref(false);
const groupViewActive = ref(false);
const releaseGroupFanView = () => { if (!groupVisible.value) groupViewActive.value = false; };
const groupSurface = ref<HTMLElement>();
let groupOpenTimer: ReturnType<typeof setTimeout> | undefined;
let groupCloseTimer: ReturnType<typeof setTimeout> | undefined;
let groupPointer: { x: number; y: number } | undefined;
let groupAnchorPosition: { top: number; left: number } | undefined;
const pointerInsideGroupFan = () => !!groupPointer && [groupAnchor.value, groupSurface.value].some(element => {
  const rect = element?.getBoundingClientRect();
  return rect && rect.width > 0 && rect.height > 0 && groupPointer!.x >= rect.left && groupPointer!.x <= rect.right
    && groupPointer!.y >= rect.top && groupPointer!.y <= rect.bottom;
});
const finishGroupLeave = () => {
  groupCloseTimer = undefined;
  // Mounting/repositioning the teleported popup can emit leave without mouse movement.
  if (pointerInsideGroupFan() || insideGroupFan(document.activeElement)) return;
  closeGroupFan();
};
const scheduleGroupLeave = () => {
  if (!groupCloseTimer) groupCloseTimer = setTimeout(finishGroupLeave, 400);
};
const trackGroupPointer = (event: PointerEvent) => {
  if (!isGroupStack.value) return;
  groupPointer = { x: event.clientX, y: event.clientY };
  if (!groupVisible.value) return;
  if (pointerInsideGroupFan()) keepGroupFan();
  else scheduleGroupLeave();
};
const keepGroupFan = () => { clearTimeout(groupCloseTimer); groupCloseTimer = undefined; };
const closeGroupFan = () => {
  clearTimeout(groupOpenTimer);
  clearTimeout(groupCloseTimer);
  groupOpenTimer = undefined;
  groupCloseTimer = undefined;
  groupVisible.value = false;
  groupPointer = undefined;
  groupAnchorPosition = undefined;
  releaseResourceGroupFan(closeGroupFan);
};
const insideGroupFan = (target: EventTarget | null) => target instanceof Node
  && (!!groupAnchor.value?.contains(target) || !!groupSurface.value?.contains(target));
const leaveGroupFan = (event: PointerEvent | FocusEvent) => {
  if ('clientX' in event) groupPointer = { x: event.clientX, y: event.clientY };
  if (insideGroupFan(event.relatedTarget)) return;
  clearTimeout(groupOpenTimer);
  groupOpenTimer = undefined;
  keepGroupFan();
  scheduleGroupLeave();
};
const openGroupFan = (event: PointerEvent | FocusEvent) => {
  if (!isGroupStack.value) return;
  if ('clientX' in event) groupPointer = { x: event.clientX, y: event.clientY };
  keepGroupFan();
  if (groupVisible.value || groupOpenTimer) return;
  groupOpenTimer = setTimeout(() => {
    groupOpenTimer = undefined;
    if (!isGroupStack.value) return;
    claimResourceGroupFan(closeGroupFan);
    prepareGroupFan();
    groupVisible.value = true;
  }, 180);
};
const outsideGroupFan = (event: PointerEvent) => { if (!insideGroupFan(event.target)) closeGroupFan(); };
const scrollGroupFan = (event: Event) => {
  if (insideGroupFan(event.target)) return;
  const rect = groupAnchor.value?.getBoundingClientRect();
  // Ignore scroll notifications caused by the popup's first layout/focus setup.
  if (rect && groupAnchorPosition && (Math.abs(rect.top - groupAnchorPosition.top) > 1
    || Math.abs(rect.left - groupAnchorPosition.left) > 1)) closeGroupFan();
};
watch(groupVisible, visible => {
  if (visible) {
    document.addEventListener('pointerdown', outsideGroupFan);
    document.addEventListener('pointermove', trackGroupPointer);
  } else {
    document.removeEventListener('pointerdown', outsideGroupFan);
    document.removeEventListener('pointermove', trackGroupPointer);
  }
  window[visible ? 'addEventListener' : 'removeEventListener']('blur', closeGroupFan);
  window[visible ? 'addEventListener' : 'removeEventListener']('scroll', scrollGroupFan, true);
  window[visible ? 'addEventListener' : 'removeEventListener']('resize', prepareGroupFan);
});
onBeforeUnmount(() => {
  closeGroupFan();
  document.removeEventListener('pointerdown', outsideGroupFan);
  document.removeEventListener('pointermove', trackGroupPointer);
  window.removeEventListener('blur', closeGroupFan);
  window.removeEventListener('scroll', scrollGroupFan, true);
  window.removeEventListener('resize', prepareGroupFan);
});
const groupPanelReady = ref(false);
const groupFanOffset = ref(-120);
const groupMemberCount = computed(() => groupContext?.summaries.value.get(props.resource.associationGroupId || '')?.count);
const groupCoverStyle = computed(() => {
  const value = appStoreData().currentConfigApp.resourceGroupCoverStyle;
  return value === 'triptych' || value === 'book' || value === 'adaptive' ? value : 'stack';
});
const groupExpandStyle = computed(() => {
  const style = appStoreData().currentConfigApp.resourceGroupExpandStyle;
  return style === 'strip' || style === 'fold' || style === 'spine' || style === 'magazine' ? style : 'fan';
});
const groupFanWidth = computed(() => groupExpandStyle.value === 'magazine' ? 560 : groupExpandStyle.value === 'spine' ? 320 : groupExpandStyle.value === 'fold'
  ? Math.min(720, Math.max(340, 248 + Math.max(0, (groupMemberCount.value || 5) - 1) * 68))
  : groupExpandStyle.value === 'strip'
  ? Math.min(640, Math.max(280, 208 + Math.max(0, (groupMemberCount.value || 5) - 1) * 40))
  : Math.min(720, Math.max(320, 200 + (groupMemberCount.value || 5) * 60)));
const prepareGroupFan = () => {
  const rect = groupAnchor.value?.getBoundingClientRect();
  const height = rect?.height || 180;
  groupAnchorPosition = rect ? { top: rect.top, left: rect.left } : undefined;
  if ((groupExpandStyle.value === 'spine' || groupExpandStyle.value === 'magazine') && rect) {
    // Freeze the opening point so the panel is close to the pointer without chasing it.
    const x = groupPointer?.x ?? rect.left + rect.width / 2;
    const y = groupPointer?.y ?? rect.top + rect.height / 2;
    const openingPoint = new DOMRect(x, y, 0, 0);
    groupPointerAnchor.value = { getBoundingClientRect: () => openingPoint, contextElement: groupAnchor.value };
  }
  // Let the floating cards overlap their source stack instead of sitting above it.
  groupFanOffset.value = (groupExpandStyle.value === 'spine' || groupExpandStyle.value === 'magazine') ? -24 : -Math.min(220, height * .8);
  groupPanelReady.value = true;
  groupViewActive.value = true;
};
const isGroupStack = computed(() => !!groupContext?.enabled.value && !!props.resource.associationGroupId && groupContext.pageIds.value.has(props.resource.id));
const stackCovers = computed(() => {
  const preview = groupContext?.summaries.value.get(props.resource.associationGroupId || '')?.preview || [];
  const others = preview.filter(resource => resource.id !== props.resource.id);
  return [0, 1].map(index => getResourceCoverPoster(others[index] || props.resource));
});
const selectGroupMember = (resource: I_resource) => {
  if (keepOpenAfterSelect) keepGroupFan();
  else closeGroupFan();
  void groupAction(resource);
};
watch(() => [props.resource.id, props.resource.associationGroupId, isGroupStack.value], () => {
  closeGroupFan();
  groupPanelReady.value = false;
});
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
.resource-management-group { position: absolute; top: 8px; left: 8px; z-index: 21; max-width: calc(100% - 48px); box-shadow: 0 1px 5px rgba(0,0,0,.2); }
.resource-selected .resource-management-group { max-width: calc(100% - 94px); }
.resource-group-stack { position: relative; isolation: isolate; overflow: hidden; border-radius: 6px; }
.resource-group-back {
  position: absolute; width: 92%; height: 92%; top: 0; right: 0; overflow: hidden;
  border: 1px solid var(--el-border-color); border-radius: 6px; box-sizing: border-box;
  background: var(--el-fill-color-dark); box-shadow: 0 1px 4px rgba(0,0,0,.18); pointer-events: none;
  img { width: 100%; height: 100%; object-fit: cover; opacity: .8; }
}
.resource-group-back--1 { top: 4%; right: 4%; }
.resource-group-front {
  position: relative; z-index: 1; width: 100%; height: 100%; transform: scale(.92); transform-origin: left bottom;
  border-radius: 6px; background: var(--el-bg-color); box-shadow: 0 2px 6px rgba(0,0,0,.2);
}

.resource-group-composite .resource-group-front { transform: none; box-shadow: none; }

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
