<template>
  <section class="group-magazine" :class="{ 'is-adaptive': adaptCoverRatio, 'is-single': members.length === 1 }" aria-label="杂志拼版" :aria-busy="loading"
    :style="[sourceStyle, { '--single-ratio': members.length === 1 ? ratio(members[0]) : 1, '--magazine-height': `${(loading || error ? expectedCount || members.length : members.length) > 5 ? 460 : 300}px` }]">
    <div v-if="loading || error || !members.length" class="magazine-state">
      <el-icon v-if="loading" class="is-loading"><Loading /></el-icon>
      <span>{{ error || (loading ? '正在展开关联资源…' : '暂无关联资源') }}</span>
      <el-button v-if="error" link @click="emit('retry')">重试</el-button>
    </div>
    <div v-else ref="scrollRef" class="magazine-scroll" @keydown="navigate">
      <div v-if="adaptCoverRatio && ordered.length > 1" class="magazine-adaptive">
        <div v-for="(row, rowIndex) in adaptiveRows" :key="rowIndex" class="magazine-adaptive-row">
          <button v-for="{ resource, index } in row" :key="resource.id" class="magazine-tile"
            :class="{ 'is-hero': index === 0, 'is-active': activeIndex === index }" type="button"
            :style="{ flexGrow: ratio(resource), aspectRatio: String(ratio(resource)) }"
            :aria-label="`打开资源：${resource.title}`" :tabindex="activeIndex === index ? 0 : -1"
            @pointerenter="activeIndex = index" @focus="activeIndex = index" @click="emit('select', resource)">
            <el-image :src="getResourceCoverPoster(resource)" fit="contain" @load="onLoad(resource, $event)" loading="lazy">
              <template #error><span class="magazine-empty"><el-icon><Picture /></el-icon></span></template>
            </el-image>
            <span class="magazine-title">{{ resource.title }}</span>
          </button>
        </div>
      </div>
      <div v-else class="magazine-lead" :class="`magazine-lead--${Math.min(ordered.length, 5)}`">
        <button v-for="(resource, index) in ordered.slice(0, 5)" :key="resource.id" class="magazine-tile"
          :class="{ 'is-hero': index === 0, 'is-active': activeIndex === index }" type="button"
          :aria-label="`打开资源：${resource.title}`" :tabindex="index === activeIndex ? 0 : -1"
          @pointerenter="activeIndex = index" @focus="activeIndex = index" @click="emit('select', resource)">
          <el-image :src="getResourceCoverPoster(resource)" :fit="adaptCoverRatio ? 'contain' : 'cover'" @load="onLoad(resource, $event)" loading="lazy"><template #error><span class="magazine-empty"><el-icon><Picture /></el-icon></span></template></el-image>
          <span class="magazine-title">{{ resource.title }}</span>
        </button>
      </div>
      <div v-if="!adaptCoverRatio && ordered.length > 5" class="magazine-rest">
        <button v-for="(resource, index) in ordered.slice(5)" :key="resource.id" class="magazine-tile"
          :class="{ 'is-active': activeIndex === index + 5 }" :style="tailStyle(index)" type="button"
          :aria-label="`打开资源：${resource.title}`" :tabindex="index + 5 === activeIndex ? 0 : -1"
          @pointerenter="activeIndex = index + 5" @focus="activeIndex = index + 5" @click="emit('select', resource)">
          <el-image :src="getResourceCoverPoster(resource)" :fit="adaptCoverRatio ? 'contain' : 'cover'" @load="onLoad(resource, $event)" loading="lazy"><template #error><span class="magazine-empty"><el-icon><Picture /></el-icon></span></template></el-image>
          <span class="magazine-title">{{ resource.title }}</span>
        </button>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useGroupPreviewSize } from './groupPreviewSize';
import { useGroupCoverRatio } from './groupCoverRatio';
import { computed, nextTick, ref, watch } from 'vue';
import { Loading, Picture } from '@element-plus/icons-vue';
import { getResourceCoverPoster } from '@/common/photo';
import type { I_resource } from '@/dataType/resource.dataType';
const props = defineProps<{ sourceElement?: HTMLElement; adaptCoverRatio?: boolean; members: I_resource[]; loading: boolean; error: string; sourceResource?: I_resource; expectedCount?: number }>();
const { ratio, onLoad } = useGroupCoverRatio();
const emit = defineEmits<{ select: [resource: I_resource]; retry: []; layoutWidth: [width: number | undefined] }>();
const { previewHeight, viewportWidth } = useGroupPreviewSize(() => props.sourceElement);
const sourceSize = computed(() => {
  if (!props.adaptCoverRatio || !props.sourceElement) return undefined;
  const singleRatio = props.members.length === 1 ? ratio(props.members[0]) : 2.8;
  const desiredWidth = props.members.length > 1
    ? Math.min(...adaptiveRows.value.map(row => previewHeight.value * row.reduce((sum, item) => sum + ratio(item.resource), 0) + (row.length - 1) * 3)) + 28
    : previewHeight.value * singleRatio + 28;
  const width = Math.min(viewportWidth.value - 24, desiredWidth);
  return { width, height: (width - 28) / singleRatio };
});
const sourceStyle = computed(() => sourceSize.value ? { '--magazine-source-height': `${sourceSize.value.height}px` } : undefined);
const scrollRef = ref<HTMLElement>();
const activeIndex = ref(0);
const ordered = computed(() => {
  const index = props.members.findIndex(resource => resource.id === props.sourceResource?.id);
  return index > 0 ? [props.members[index], ...props.members.slice(0, index), ...props.members.slice(index + 1)] : props.members;
});
// Equal-height rows fill the available width using the actual cover ratios.
// Keep source order and let portrait-heavy rows contain more covers.
const adaptiveRows = computed(() => {
  const rows: { resource: I_resource; index: number }[][] = [];
  let row: { resource: I_resource; index: number }[] = [];
  let sum = 0;
  ordered.value.forEach((resource, index) => {
    row.push({ resource, index });
    sum += ratio(resource);
    if (sum >= 2.8 || row.length >= 5) {
      rows.push(row); row = []; sum = 0;
    }
  });
  if (row.length) {
    if (sum < 1.4 && rows.length) rows[rows.length - 1].push(...row);
    else rows.push(row);
  }
  return rows;
});
watch(() => sourceSize.value?.width, value => emit('layoutWidth', value), { immediate: true });
watch(ordered, () => { activeIndex.value = 0; if (scrollRef.value) scrollRef.value.scrollTop = 0; });
const tailStyle = (index: number) => {
  const count = ordered.value.length - 5;
  const remainder = count % 3;
  return { gridColumn: `span ${remainder && index >= count - remainder ? 6 / remainder : 2}` };
};
const navigate = async (event: KeyboardEvent) => {
  if (!['ArrowLeft','ArrowRight','ArrowUp','ArrowDown','Home','End'].includes(event.key)) return;
  event.preventDefault(); event.stopPropagation();
  const last = ordered.value.length - 1;
  activeIndex.value = event.key === 'Home' ? 0 : event.key === 'End' ? last
    : Math.max(0, Math.min(last, activeIndex.value + (event.key === 'ArrowLeft' || event.key === 'ArrowUp' ? -1 : 1)));
  await nextTick();
  const tile = scrollRef.value?.querySelectorAll<HTMLElement>('.magazine-tile')[activeIndex.value];
  tile?.focus({ preventScroll: true });
  if (tile && scrollRef.value) {
    const container = scrollRef.value;
    const box = tile.getBoundingClientRect();
    const viewport = container.getBoundingClientRect();
    if (box.top < viewport.top) container.scrollTop += box.top - viewport.top;
    else if (box.bottom > viewport.bottom) container.scrollTop += box.bottom - viewport.bottom;
  }
};
</script>

<style scoped lang="scss">
.group-magazine { --magazine-bg: #25292f; --magazine-border: rgba(255,255,255,.16); --magazine-shadow: rgba(0,0,0,.32); padding: 14px; color: var(--el-text-color-primary); }
:global(html.bright) .group-magazine { --magazine-bg: #e8ecf0; --magazine-border: rgba(40,55,70,.14); --magazine-shadow: rgba(30,45,60,.2); }
.magazine-scroll, .magazine-state { height: min(var(--magazine-height, 300px), calc(100vh - 56px)); overflow: auto; border-radius: 12px; background: var(--magazine-bg); box-shadow: 0 10px 28px var(--magazine-shadow); }
.magazine-scroll { container-type: inline-size; overscroll-behavior: contain; scrollbar-width: thin; outline: 1px solid var(--magazine-border); outline-offset: -1px; }
.magazine-state { display: flex; align-items: center; justify-content: center; gap: 10px; font-size: 13px; }
.magazine-lead { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); grid-template-rows: repeat(2, minmax(0, 1fr)); height: 300px; gap: 3px; }
.magazine-lead--1 { grid-template-columns: 1fr; grid-template-rows: 1fr; }
.magazine-lead--2 { grid-template-columns: 1.2fr .8fr; grid-template-rows: 1fr; }
.magazine-lead--1 .is-hero, .magazine-lead--2 .is-hero { grid-column: auto; grid-row: auto; }
.is-hero { grid-column: span 2; grid-row: span 2; }
.magazine-lead--3 .magazine-tile:not(.is-hero), .magazine-lead--4 .magazine-tile:nth-child(2) { grid-column: span 2; }
.magazine-rest { display: grid; grid-template-columns: repeat(6, minmax(0, 1fr)); grid-auto-rows: 144px; gap: 3px; margin-top: 3px; }
.magazine-adaptive { display: grid; gap: 3px; }
.is-adaptive:not(.is-single) .magazine-scroll { height: auto; max-height: min(var(--magazine-height, 300px), calc(100vh - 56px)); }
.magazine-adaptive-row { display: flex; align-items: flex-start; gap: 3px; }
.magazine-adaptive-row .magazine-tile { flex-basis: 0; flex-shrink: 1; }
.magazine-adaptive-row .magazine-tile .el-image { position: absolute; inset: 0; }
.magazine-tile { display: block; position: relative; min-width: 0; min-height: 0; padding: 0; border: 0; overflow: hidden; cursor: pointer; color: white; background: var(--magazine-bg); text-align: left; }
.magazine-tile .el-image { display: block; width: 100%; height: 100%; }
.magazine-tile :deep(.el-image__inner) { transition: transform .2s ease, filter .2s ease; filter: brightness(.88); }
.magazine-tile.is-hero :deep(.el-image__inner) { filter: none; }
.magazine-tile:hover :deep(.el-image__inner), .magazine-tile:focus-visible :deep(.el-image__inner) { transform: scale(1.035); filter: brightness(1.04); }
.magazine-tile:focus-visible { outline: 2px solid var(--el-color-primary); outline-offset: -3px; z-index: 1; }
.magazine-title { position: absolute; inset: auto 0 0; padding: 28px 10px 10px; background: linear-gradient(transparent,rgba(0,0,0,.76)); font-size: 12px; line-height: 18px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; opacity: 0; transition: opacity .16s; pointer-events: none; }
.is-hero .magazine-title { font-size: 14px; }
.is-hero .magazine-title, .magazine-tile:hover .magazine-title, .magazine-tile:focus-visible .magazine-title { opacity: 1; }
.magazine-empty { display: grid; place-items: center; width: 100%; height: 100%; font-size: 28px; color: var(--el-text-color-placeholder); }
@container (max-width: 380px) {
  .magazine-lead { height: auto; grid-template-columns: repeat(2, minmax(0, 1fr)); grid-template-rows: none; grid-auto-rows: 150px; }
  .magazine-lead .is-hero { grid-column: span 2; grid-row: auto; }
  .magazine-lead--1 { grid-auto-rows: 300px; }
  .magazine-lead--2 .magazine-tile:not(.is-hero) { grid-column: span 2; }
  .magazine-lead--3 .magazine-tile:not(.is-hero) { grid-column: span 1; }
}
@media (prefers-reduced-motion: reduce) { .magazine-tile :deep(.el-image__inner), .magazine-title { transition: none; } }
.is-adaptive.is-single { box-sizing: border-box; width: min(100%, calc(var(--magazine-source-height, 300px) * var(--single-ratio) + 28px)); }
.is-adaptive .magazine-tile:hover :deep(.el-image__inner), .is-adaptive .magazine-tile:focus-visible :deep(.el-image__inner) { transform: none; }
/* Preserve the original Popper box and anchor; align to its source-facing edge when flipped. */
:global(.el-popper[data-popper-placement^="left"] .group-magazine.is-adaptive.is-single) { margin-left: auto; }
.is-adaptive.is-single .magazine-scroll, .is-adaptive.is-single .magazine-lead { height: var(--magazine-source-height, 300px); }
.is-adaptive.is-single .magazine-lead { grid-template-rows: minmax(0, 1fr); grid-auto-rows: auto; }
</style>
