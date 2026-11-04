<template>
  <div v-if="enabled" class="resource-quick-stars" :class="{ 'is-readonly': !editable, 'is-pending': pending }"
    :aria-label="`评星：${shownStars} 星`" :aria-busy="pending"
    @click.stop @dblclick.stop @pointerdown.stop @pointerup.stop @contextmenu.stop @keydown.stop @keydown.esc="escape" @dragstart.prevent>
    <div ref="starsElement" class="quick-stars-track" @pointerdown="start" @pointermove="move"
      @pointerup="finish" @pointercancel="cancel" @lostpointercapture="lostCapture" @pointerleave="leave">
      <button v-for="star in 5" :key="star" type="button" :disabled="!editable || pending"
        :aria-label="`评为 ${star} 星`" :aria-pressed="resource.stars === star"
        :title="editable ? `评为 ${star} 星（可按住拖动）` : `${resource.stars || 0} 星`"
        :class="{ filled: star <= shownStars }" @click="keyboardClick($event, star)"
        @focus="preview = star" @blur="leave">
        <StarFilled v-if="star <= shownStars" /><Star v-else />
      </button>
    </div>
    <button v-if="editable" class="quick-stars-clear" type="button" title="清除评星" aria-label="清除评星"
      :disabled="pending || !resource.stars" @click="save(0)">清除</button>
    <span class="quick-stars-status" role="status">{{ pending ? '保存中' : feedback }}</span>
  </div>
</template>
<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { Star, StarFilled } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';
import type { I_resource } from '@/dataType/resource.dataType';
import { resourceServer } from '@/server/resource.server';
import { pendingResourceStars, resourceStarsKey, syncResourceStars, trackResourceStars, useQuickResourceStars } from '@/common/resourceStars';

const props = defineProps<{ resource: I_resource }>();
const emit = defineEmits<{ dragging: [active: boolean]; escape: [] }>();
const { enabled, editable } = useQuickResourceStars(() => props.resource);
const pending = computed(() => pendingResourceStars.has(resourceStarsKey(props.resource)));
const preview = ref<number>();
const pointer = ref<number>();
const starsElement = ref<HTMLElement>();
const feedback = ref('');
const savingPreview = ref(false);
let saveVersion = 0;
const shownStars = computed(() => preview.value ?? props.resource.stars ?? 0);
trackResourceStars(() => [props.resource]);

function valueAt(event: PointerEvent) {
  const rect = starsElement.value?.getBoundingClientRect();
  if (!rect || rect.width <= 0 || !Number.isFinite(event.clientX)) return undefined;
  return Math.max(1, Math.min(5, Math.ceil((event.clientX - rect.left) / rect.width * 5)));
}
function start(event: PointerEvent) {
  if (!editable.value || pending.value || event.pointerType !== 'mouse' || event.button !== 0) return;
  const stars = valueAt(event);
  if (stars === undefined) return;
  event.preventDefault();
  pointer.value = event.pointerId;
  emit('dragging', true);
  window.addEventListener('blur', cancel);
  starsElement.value?.setPointerCapture(event.pointerId);
  preview.value = stars;
}
function move(event: PointerEvent) {
  if (!editable.value || pending.value || event.pointerType !== 'mouse') return;
  if (pointer.value !== undefined && pointer.value !== event.pointerId) return;
  preview.value = valueAt(event);
}
function finish(event: PointerEvent) {
  if (pointer.value !== event.pointerId) return;
  const stars = valueAt(event);
  cancel();
  if (stars !== undefined) void save(stars);
}
function cancel() {
  const id = pointer.value;
  pointer.value = undefined;
  if (id !== undefined) emit('dragging', false);
  window.removeEventListener('blur', cancel);
  preview.value = undefined;
  if (id !== undefined && starsElement.value?.hasPointerCapture(id)) starsElement.value.releasePointerCapture(id);
}
function lostCapture() { if (pointer.value !== undefined) cancel(); }
function escape() { cancel(); emit('escape'); }
function leave() { if (pointer.value === undefined && !savingPreview.value) preview.value = undefined; }
function keyboardClick(event: MouseEvent, stars: number) {
  // Mouse clicks already commit on pointerup. Keyboard activation has detail=0.
  if (event.detail === 0) void save(stars);
}
async function save(stars: number) {
  if (pending.value) return;
  if (!editable.value || stars === props.resource.stars) { preview.value = undefined; return; }
  const resource = props.resource;
  const key = resourceStarsKey(resource);
  const version = ++saveVersion;
  savingPreview.value = true;
  pendingResourceStars.add(key);
  preview.value = stars;
  feedback.value = '';
  try {
    const result = await resourceServer.batchSetStars([resource.id], stars);
    if (!result?.status) throw new Error(result?.msg || '评星保存失败');
    syncResourceStars(resource, stars);
    if (version === saveVersion) feedback.value = stars ? `已保存 ${stars} 星` : '已清除评星';
  } catch (error) {
    if (version === saveVersion) ElMessage.error(error instanceof Error ? error.message : '评星保存失败，请重试');
  } finally {
    pendingResourceStars.delete(key);
    if (version === saveVersion) { savingPreview.value = false; preview.value = undefined; }
  }
}
watch(() => [props.resource, resourceStarsKey(props.resource), enabled.value, editable.value], () => {
  saveVersion++;
  savingPreview.value = false;
  cancel();
  feedback.value = '';
});
onBeforeUnmount(() => { saveVersion++; cancel(); });
</script>
<style scoped>
.resource-quick-stars { display: flex; align-items: center; box-sizing: border-box; gap: 8px; width: 100%; min-width: 0; min-height: 46px; padding: 4px 6px; background: var(--el-fill-color-light); border: 1px solid var(--el-border-color-lighter); border-radius: 9px; user-select: none; cursor: default; }
.quick-stars-track { display: flex; flex: 1; min-width: 0; }
button { display: inline-flex; justify-content: center; align-items: center; padding: 0; border: 0; background: transparent; color: var(--el-text-color-placeholder); cursor: pointer; height: 36px; min-width: 0; }
.quick-stars-track button { flex: 1; border-radius: 5px; }
button svg { width: 23px; height: 23px; max-width: 100%; pointer-events: none; }
button.filled { color: #e7a038; }
.quick-stars-track button:hover:not(:disabled) { background: var(--el-color-warning-light-9); }
button:focus-visible { outline: 2px solid var(--el-color-primary); outline-offset: 1px; border-radius: 3px; }
button:disabled { cursor: default; }
.quick-stars-clear { flex: 0 0 34px; height: 22px; padding-left: 8px; border-left: 1px solid var(--el-border-color); color: var(--el-text-color-secondary); font-size: 11px; }
.quick-stars-clear:hover:not(:disabled) { color: var(--el-color-primary); }
.quick-stars-clear:disabled { color: var(--el-text-color-placeholder); cursor: default; }
.is-pending { opacity: .65; }
.quick-stars-status { position: absolute; width: 1px; height: 1px; overflow: hidden; clip-path: inset(50%); }
</style>
