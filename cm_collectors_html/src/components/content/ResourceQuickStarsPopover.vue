<template>
  <div v-if="enabled" ref="entry" class="resource-quick-stars-entry" :class="{ 'is-table': table, 'is-open': visible }"
    :style="{ '--quick-stars-bottom': offsetBottom || '6px' }"
    @pointerenter="enter" @pointerleave="leave" @focusin="focusOpen" @focusout="focusLeave"
    @click.stop @dblclick.stop @pointerdown.stop @contextmenu.stop @keydown.stop @keydown.esc="close">
    <el-popover :visible="visible" :placement="table ? 'bottom-end' : 'top-end'" :width="252" :offset="8" :show-arrow="false"
      :persistent="false" popper-class="resource-quick-stars-popper"
      :popper-style="{ padding: '0', borderRadius: '12px', boxShadow: 'var(--el-box-shadow-light)' }"
      :popper-options="{ modifiers: [{ name: 'flip', options: { fallbackPlacements: table ? ['top-end', 'bottom-start', 'top-start'] : ['bottom-end', 'top-start', 'bottom-start'] } }, { name: 'preventOverflow', options: { padding: 8 } }] }">
      <template #reference>
        <button class="quick-stars-trigger" type="button" :aria-label="`评星：${resource.title}，${resource.stars || 0} 星`"
          :aria-expanded="visible" :aria-busy="pending" @click="openImmediately">
          <StarFilled v-if="resource.stars" /><Star v-else />
          <span v-if="resource.stars || table">{{ resource.stars || '未评星' }}</span>
        </button>
      </template>
      <div ref="panel" class="quick-stars-panel" @pointerenter="enterPanel" @pointerleave="leave"
        @focusin="keepOpen" @focusout="focusLeave" @click.stop @dblclick.stop @pointerdown.stop @contextmenu.stop
        @keydown.stop @keydown.esc="close">
        <div class="quick-stars-heading">
          <div class="quick-stars-resource" :title="resource.title">{{ resource.title }}</div>
          <span class="quick-stars-value" :class="{ 'has-rating': resource.stars > 0 }"
            :title="editable ? '当前评星' : '仅管理员可修改评星'">{{ pending ? '保存中' : resource.stars ? `${resource.stars} 星` : '未评星' }}</span>
        </div>
        <ResourceQuickStars :resource="resource" @dragging="dragging = $event" @escape="close" />
      </div>
    </el-popover>
  </div>
</template>
<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { ElPopover } from 'element-plus';
import { Star, StarFilled } from '@element-plus/icons-vue';
import type { I_resource } from '@/dataType/resource.dataType';
import { claimQuickStarsPanel, releaseQuickStarsPanel, pendingResourceStars, resourceStarsKey, useQuickResourceStars } from '@/common/resourceStars';
import ResourceQuickStars from './ResourceQuickStars.vue';

const props = defineProps<{ resource: I_resource; table?: boolean; offsetBottom?: string }>();
const { enabled, editable } = useQuickResourceStars(() => props.resource);
const visible = ref(false);
const dragging = ref(false);
let hovered = false;
const entry = ref<HTMLElement>();
const panel = ref<HTMLElement>();
const pending = computed(() => pendingResourceStars.has(resourceStarsKey(props.resource)));
let openTimer: ReturnType<typeof setTimeout> | undefined;
let closeTimer: ReturnType<typeof setTimeout> | undefined;
const inside = (target: EventTarget | null) => target instanceof Node && (!!entry.value?.contains(target) || !!panel.value?.contains(target));
function keepOpen() { clearTimeout(closeTimer); closeTimer = undefined; }
function enter() { hovered = true; open(); }
function enterPanel() { hovered = true; keepOpen(); }
function leave() { hovered = false; scheduleClose(); }
function focusOpen(event: FocusEvent) { if (event.target instanceof Element && event.target.matches(':focus-visible')) open(); }
function hasKeyboardFocus() { return inside(document.activeElement) && !!document.activeElement?.matches(':focus-visible'); }
function open() {
  if (!enabled.value) return;
  keepOpen();
  if (visible.value || openTimer) return;
  openTimer = setTimeout(() => { openTimer = undefined; if (enabled.value) show(); }, 150);
}
function show() { claimQuickStarsPanel(close); visible.value = true; }
function openImmediately() { clearTimeout(openTimer); openTimer = undefined; keepOpen(); if (enabled.value) show(); }
function close() { clearTimeout(openTimer); clearTimeout(closeTimer); openTimer = undefined; closeTimer = undefined; visible.value = false; releaseQuickStarsPanel(close); }
function scheduleClose() {
  clearTimeout(openTimer); openTimer = undefined;
  keepOpen();
  closeTimer = setTimeout(() => {
    closeTimer = undefined;
    if (!dragging.value && !hovered && !hasKeyboardFocus()) close();
  }, 250);
}
function focusLeave(event: FocusEvent) { if (!inside(event.relatedTarget)) scheduleClose(); }
function outside(event: Event) { if (!inside(event.target)) close(); }
function scroll(event: Event) { if (!inside(event.target)) close(); }
watch(visible, value => {
  document[value ? 'addEventListener' : 'removeEventListener']('pointerdown', outside);
  window[value ? 'addEventListener' : 'removeEventListener']('blur', close);
  window[value ? 'addEventListener' : 'removeEventListener']('scroll', scroll, true);
});
watch(() => [props.resource, resourceStarsKey(props.resource), enabled.value], close);
watch(dragging, value => { if (!value && visible.value) scheduleClose(); });
onBeforeUnmount(() => {
  close(); document.removeEventListener('pointerdown', outside); window.removeEventListener('blur', close); window.removeEventListener('scroll', scroll, true);
});
</script>
<style scoped>
.resource-quick-stars-entry { position: absolute; bottom: var(--quick-stars-bottom, 6px); right: 6px; z-index: 22; opacity: 0; pointer-events: none; transition: opacity .15s; }
.resource-quick-stars-entry:not(.is-table):has(~ .video-duration-badge) { bottom: calc(var(--quick-stars-bottom, 6px) + 28px); }
.resource-quick-stars-entry.is-table { position: static; display: inline-flex; opacity: 1; pointer-events: auto; }
.resource-quick-stars-entry.is-open, .resource-quick-stars-entry:focus-within { opacity: 1; pointer-events: auto; }
.quick-stars-trigger { display: inline-flex; align-items: center; justify-content: center; gap: 4px; height: 28px; min-width: 28px; padding: 0 7px; color: #ffcf70; background: rgba(20, 24, 31, .78); border: 1px solid rgba(255,255,255,.2); border-radius: 8px; cursor: pointer; box-shadow: 0 2px 8px rgba(0,0,0,.16); }
.quick-stars-trigger svg { width: 15px; height: 15px; }
.quick-stars-trigger span { font-size: 12px; line-height: 1; }
.quick-stars-trigger:focus-visible { outline: 2px solid var(--el-color-primary); outline-offset: 2px; }
.is-table .quick-stars-trigger { background: var(--el-fill-color-light); border-color: var(--el-border-color-lighter); color: var(--el-color-warning); box-shadow: none; }
.quick-stars-panel { padding: 12px; background: var(--el-bg-color-overlay); border-radius: 12px; }
.quick-stars-heading { display: flex; align-items: center; gap: 10px; margin: 0 2px 10px; }
.quick-stars-resource { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--el-text-color-regular); font-size: 12px; font-weight: 500; line-height: 20px; }
.quick-stars-value { flex: 0 0 auto; padding: 1px 7px; border-radius: 5px; background: var(--el-fill-color-light); color: var(--el-text-color-secondary); font-size: 11px; line-height: 18px; }
.quick-stars-value.has-rating { color: var(--el-color-warning-dark-2); background: var(--el-color-warning-light-9); }
</style>
