<template>
  <section class="group-fold" :class="{ 'group-fold--spine': vertical }" :style="vertical ? { '--spine-height': `${Math.min(520, 260 + Math.max(0, (loading || error ? expectedCount || members.length : members.length) - 1) * 28)}px` } : undefined" :aria-busy="loading" :aria-label="vertical ? '纵向封面书脊' : '封面折屏'">
    <div v-if="loading || error || !members.length" class="fold-state">
      <el-icon v-if="loading" class="is-loading"><Loading /></el-icon>
      <span>{{ error || (loading ? '正在展开关联资源…' : '暂无关联资源') }}</span>
      <el-button v-if="error" link @click="emit('retry')">重试</el-button>
    </div>
    <div v-else ref="deckRef" class="fold-screen" @pointermove="pointToSlice" @keydown="navigate" @wheel="browseWheel"
      @pointerdown.left="pressSlice" @pointercancel="pressed = undefined" @click="selectSlice">
      <button v-for="(resource, index) in members" :key="resource.id" class="fold-slice" type="button"
        :class="{ 'is-active': index === activeIndex }" :style="sliceStyle(index)"
        :aria-label="`打开资源：${resource.title}`" :tabindex="index === activeIndex ? 0 : -1"
        @focus="activeIndex = index">
        <el-image :src="getResourceCoverPoster(resource)" fit="cover" loading="lazy">
          <template #error><span class="fold-empty"><el-icon><Picture /></el-icon></span></template>
        </el-image>
        <span class="fold-title">{{ resource.title }}</span>
      </button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue';
import { Loading, Picture } from '@element-plus/icons-vue';
import { getResourceCoverPoster } from '@/common/photo';
import type { I_resource } from '@/dataType/resource.dataType';
const props = defineProps<{ members: I_resource[]; loading: boolean; error: string; sourceResource?: I_resource; vertical?: boolean; expectedCount?: number }>();
const emit = defineEmits<{ select: [resource: I_resource]; retry: [] }>();
const deckRef = ref<HTMLElement>();
const width = ref(480);
const height = ref(480);
const activeIndex = ref(0);
let observer: ResizeObserver | undefined;
let lastPointer: { x: number; y: number } | undefined;
let pressed: I_resource | undefined;
let wheelDelta = 0;
watch(deckRef, element => {
  observer?.disconnect();
  if (!element) return;
  const measure = () => {
    const rect = element.getBoundingClientRect();
    width.value = rect.width || 480;
    height.value = rect.height || 480;
  };
  measure();
  observer = new ResizeObserver(measure);
  observer.observe(element);
}, { flush: 'post' });
onBeforeUnmount(() => observer?.disconnect());
watch(() => [props.members, props.sourceResource?.id, props.vertical], () => {
  const sourceIndex = props.members.findIndex(resource => resource.id === props.sourceResource?.id);
  activeIndex.value = sourceIndex >= 0 ? sourceIndex : Math.floor(props.members.length / 2);
  pressed = undefined; lastPointer = undefined; wheelDelta = 0;
}, { immediate: true });
const geometry = computed(() => {
  const count = props.members.length || 1;
  const size = props.vertical ? height.value : width.value;
  const gap = Math.min(2, size / (count * 8));
  const available = Math.max(0, size - gap * (count - 1));
  const expanded = count === 1 ? available : Math.max(available / count, Math.min(200, available * .66));
  return { gap, expanded, narrow: count === 1 ? 0 : (available - expanded) / (count - 1) };
});
const sliceStyle = (index: number) => {
  const { gap, expanded, narrow } = geometry.value;
  const start = `${index * (narrow + gap) + (index > activeIndex.value ? expanded - narrow : 0)}px`;
  const size = `${index === activeIndex.value ? expanded : narrow}px`;
  return props.vertical ? { top: start, height: size } : { left: start, width: size };
};
const followExistingFocus = () => {
  const deck = deckRef.value;
  if (!deck?.contains(document.activeElement)) return;
  void nextTick(() => {
    if (deck === deckRef.value && deck.contains(document.activeElement)) {
      (deck.children[activeIndex.value] as HTMLElement | undefined)?.focus({ preventScroll: true });
    }
  });
};
const pointToSlice = (event: { clientX: number; clientY: number }) => {
  if (lastPointer?.x === event.clientX && lastPointer.y === event.clientY) return;
  lastPointer = { x: event.clientX, y: event.clientY };
  const rect = deckRef.value?.getBoundingClientRect();
  const size = props.vertical ? rect?.height : rect?.width;
  if (!rect || !size || !props.members.length) return;
  // Fixed equal selection zones do not move when the visible slices expand.
  const coordinate = props.vertical ? event.clientY - rect.top : event.clientX - rect.left;
  const position = coordinate / size * props.members.length - .5;
  if (Math.abs(position - activeIndex.value) < .58) return;
  activeIndex.value = Math.max(0, Math.min(props.members.length - 1, Math.round(position)));
  followExistingFocus();
};
const navigate = async (event: KeyboardEvent) => {
  const previous = props.vertical ? 'ArrowUp' : 'ArrowLeft';
  const next = props.vertical ? 'ArrowDown' : 'ArrowRight';
  if (![previous,next,'Home','End'].includes(event.key)) return;
  event.preventDefault();
  event.stopPropagation();
  const last = props.members.length - 1;
  activeIndex.value = event.key === 'Home' ? 0 : event.key === 'End' ? last
    : Math.max(0, Math.min(last, activeIndex.value + (event.key === previous ? -1 : 1)));
  await nextTick();
  (deckRef.value?.children[activeIndex.value] as HTMLElement | undefined)?.focus({ preventScroll: true });
};
const browseWheel = (event: WheelEvent) => {
  if (props.members.length < 2 || event.ctrlKey || !event.deltaX && !event.deltaY) return;
  event.preventDefault();
  const delta = Math.abs(event.deltaX) > Math.abs(event.deltaY) ? event.deltaX : event.deltaY;
  if (Math.sign(delta) !== Math.sign(wheelDelta)) wheelDelta = 0;
  wheelDelta += delta * (event.deltaMode === 1 ? 16 : event.deltaMode === 2 ? 240 : 1);
  if (Math.abs(wheelDelta) < 36) return;
  activeIndex.value = Math.max(0, Math.min(props.members.length - 1, activeIndex.value + Math.sign(wheelDelta)));
  followExistingFocus();
  wheelDelta = 0;
};
const pressSlice = (event: PointerEvent) => {
  event.preventDefault(); pointToSlice(event); pressed = props.members[activeIndex.value];
};
const selectSlice = (event: MouseEvent) => {
  // Preserve precise pointerdown selection and stationary wheel/keyboard selection.
  if (event.detail && !pressed) pointToSlice(event);
  const button = (event.target as Element).closest('button');
  const focusedIndex = button && deckRef.value ? Array.from(deckRef.value.children).indexOf(button) : activeIndex.value;
  const selected = event.detail ? pressed || props.members[activeIndex.value] : props.members[focusedIndex];
  pressed = undefined;
  if (selected) emit('select', selected);
};
</script>

<style scoped lang="scss">
.group-fold { --fold-gap: #22262c; --fold-border: rgba(255,255,255,.18); --fold-shadow: rgba(0,0,0,.34); padding: 14px; color: var(--el-text-color-primary); }
:global(html.bright) .group-fold { --fold-gap: #e6e9ed; --fold-border: rgba(45,60,75,.16); --fold-shadow: rgba(30,45,60,.2); }
.fold-screen { position: relative; height: 238px; max-height: calc(100vh - 56px); overflow: hidden; border-radius: 12px; background: var(--fold-gap); box-shadow: 0 10px 28px var(--fold-shadow); isolation: isolate; }
.fold-screen::after { content: ''; position: absolute; inset: 0; border: 1px solid var(--fold-border); border-radius: inherit; box-sizing: border-box; pointer-events: none; z-index: 2; }
.fold-slice { position: absolute; top: 0; bottom: 0; padding: 0; border: 0; overflow: hidden; background: var(--fold-gap); cursor: pointer; transition: left .2s cubic-bezier(.2,.7,.2,1), width .2s cubic-bezier(.2,.7,.2,1); }
.fold-slice .el-image { display: block; width: 100%; height: 100%; }
.fold-slice::after { content: ''; position: absolute; inset: 0; background: rgba(0,0,0,.24); transition: opacity .2s; pointer-events: none; }
.fold-slice.is-active::after { opacity: 0; }
.fold-slice:focus-visible { outline: 2px solid var(--el-color-primary); outline-offset: -3px; z-index: 1; }
.fold-title { position: absolute; inset: auto 0 0; padding: 30px 12px 12px; color: white; background: linear-gradient(transparent,rgba(0,0,0,.78)); font-size: 13px; line-height: 20px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; opacity: 0; transition: opacity .15s; pointer-events: none; }
.is-active .fold-title { opacity: 1; }
.fold-empty { display: grid; place-items: center; width: 100%; height: 100%; color: var(--el-text-color-placeholder); font-size: 30px; }
.fold-state { height: 238px; max-height: calc(100vh - 56px); display: flex; gap: 10px; align-items: center; justify-content: center; border-radius: 12px; background: var(--el-bg-color-overlay); box-shadow: 0 10px 28px var(--fold-shadow); font-size: 13px; }
.group-fold--spine .fold-screen, .group-fold--spine .fold-state { height: var(--spine-height, 400px); }
.group-fold--spine .fold-slice { left: 0; right: 0; bottom: auto; width: 100%; text-align: left; transition: top .2s cubic-bezier(.2,.7,.2,1), height .2s cubic-bezier(.2,.7,.2,1); }
.group-fold--spine .fold-title { inset: 0 0 auto; padding: 3px 12px; font-size: 12px; line-height: 20px; background: linear-gradient(90deg,rgba(0,0,0,.62),rgba(0,0,0,.2)); opacity: 0; z-index: 1; }
.group-fold--spine .is-active .fold-title { inset: auto 0 0; padding: 28px 12px 12px; background: linear-gradient(transparent,rgba(0,0,0,.78)); font-size: 13px; opacity: 1; }
@media (prefers-reduced-motion: reduce) { .fold-slice, .group-fold--spine .fold-slice, .fold-slice::after, .fold-title { transition: none; } }
</style>
