<template>
  <article
    ref="card"
    class="category-card"
    :style="categoryCardDisplay(displayConfig || {}, source)"
    @pointerenter="open"
    @pointerleave="leave"
    @focusin="open"
    @focusout="leave"
    @keydown.esc="close"
  >
    <button
      class="category-cover"
      type="button"
      :aria-label="`查看${group.name}第${group.part}组`"
      @click="emit('enter', group, false)"
    >
      <template v-if="coverStyle === 'stack' && group.count > 1">
        <img
          v-for="(item, index) in group.preview.slice(1, 3)"
          :key="item.id"
          class="category-back"
          :class="`category-back-${index}`"
          :src="getResourceCoverPoster(item)"
          alt=""
        />
      </template>
      <ResourceGroupCover
        v-if="source"
        class="category-front"
        :class="{ 'is-stack': coverStyle === 'stack' && group.count > 1 }"
        :resource="source"
        :src="getResourceCoverPoster(source)"
        :cover-style="coverStyle"
        :fit="displayConfig?.coverImageFit || 'cover'"
        :composition-fit="displayConfig?.coverImageFit || 'cover'"
      />
      <span v-if="group.parts > 1" class="category-part">{{ group.part }} / {{ group.parts }}</span>
      <span class="category-count">{{ group.count }} 个</span>
    </button>
    <div class="category-caption">
      <el-avatar v-if="group.photo" :src="group.photo" :size="25" /><strong :title="group.name">{{
        group.name
      }}</strong>
    </div>
    <div class="category-meta">
      共 {{ group.total }} 个资源<span v-if="group.parts > 1"> · 第 {{ group.part }} 组</span>
    </div>
    <el-button
      v-if="group.parts > 1"
      class="category-all"
      link
      size="small"
      @click="emit('enter', group, true)"
      >查看全部 ›</el-button
    >
    <el-popover
      :visible="visible"
      :virtual-ref="card"
      virtual-triggering
      :placement="isCardDeck ? 'top' : 'right-start'"
      :width="panelWidth"
      :show-arrow="false"
      :offset="isCardDeck ? deckOffset : -30"
      :popper-options="{
        modifiers: [{ name: 'preventOverflow', options: { padding: 12, altAxis: true } }],
      }"
      :popper-style="{
        maxWidth: 'calc(100vw - 24px)',
        padding: '0',
        border: '0',
        background: 'transparent',
        boxShadow: 'none',
      }"
    >
      <div
        v-if="visible"
        ref="surface"
        class="category-preview"
        :class="{ 'category-preview--deck': isCardDeck }"
        @pointerenter="keep"
        @pointerleave="leave"
        @focusin="keep"
        @focusout="leave"
        @keydown.esc="close"
      >
        <ResourceGroupMagazine
          v-if="expandStyle === 'magazine'"
          :members="members" :adapt-cover-ratio="adaptCoverRatio"
          :source-resource="source"
          :loading="loading"
          :error="error"
          :expected-count="group.count"
          @retry="load"
          @select="select"
        />
        <ResourceGroupFold
          v-else-if="expandStyle === 'fold' || expandStyle === 'spine'"
          :vertical="expandStyle === 'spine'"
          :source-resource="source"
          :members="members" :adapt-cover-ratio="adaptCoverRatio"
          :loading="loading"
          :error="error"
          :expected-count="group.count"
          @retry="load"
          @select="select"
        />
        <ResourceGroupFan
          v-else
          :expand-style="expandStyle === 'strip' ? 'strip' : 'fan'"
          :members="members" :adapt-cover-ratio="adaptCoverRatio"
          :loading="loading"
          :error="error"
          :expected-count="group.count"
          @retry="load"
          @select="select"
        />
      </div>
    </el-popover>
  </article>
</template>
<script setup lang="ts">
import { computed, onBeforeUnmount, provide, ref, watch } from 'vue'
import type { I_resource } from '@/dataType/resource.dataType'
import type { I_config_app } from '@/dataType/config.dataType'
import {
  categoryBrowseServer,
  type CategoryGroup,
  type CategoryQuery,
} from '@/server/categoryBrowse.server'
import { getResourceCoverPoster } from '@/common/photo'
import { playResource } from '@/common/play'
import {
  claimResourceGroupFan,
  releaseResourceGroupFan,
  resourceGroupContextKey,
} from './resourceGroups'
import ResourceGroupCover from './ResourceGroupCover.vue'
import { categoryCardDisplay } from './categoryDisplay'
import ResourceGroupFan from './ResourceGroupFan.vue'
import ResourceGroupFold from './ResourceGroupFold.vue'
import ResourceGroupMagazine from './ResourceGroupMagazine.vue'
const props = defineProps<{
  displayConfig?: I_config_app
  group: CategoryGroup
  query: CategoryQuery
  coverStyle: I_config_app['categoryBrowseCoverStyle']
  hideSingle?: boolean
  adaptCoverRatio?: boolean
  expandStyle: I_config_app['categoryBrowseExpandStyle']
}>()
const emit = defineEmits<{ enter: [group: CategoryGroup, all: boolean] }>()
const key = computed(() => `${props.query.source}:${props.group.categoryId}:${props.group.part}`)
const source = computed(() =>
  props.group.preview[0] ? { ...props.group.preview[0], associationGroupId: key.value } : undefined,
)
provide(resourceGroupContextKey, {
  enabled: computed(() => true),
  pageIds: computed(() => new Set(source.value ? [source.value.id] : [])),
  summaries: computed(
    () =>
      new Map([
        [
          key.value,
          {
            id: key.value,
            filesBasesId: props.query.filesBasesId,
            count: props.group.count,
            preview: props.group.preview,
          },
        ],
      ]),
  ),
})
const card = ref<HTMLElement>(),
  surface = ref<HTMLElement>()
const visible = ref(false),
  loading = ref(false),
  error = ref(''),
  members = ref<I_resource[]>([])
let requestVersion = 0,
  openTimer: ReturnType<typeof setTimeout> | undefined,
  closeTimer: ReturnType<typeof setTimeout> | undefined
let pointer: { x: number; y: number } | undefined
const isCardDeck = computed(() => props.expandStyle === 'fan' || props.expandStyle === 'strip')
const deckOffset = ref(-120)
// Match the resource association group's deck dimensions and overlap positioning.
const panelWidth = computed(() =>
  props.expandStyle === 'fan'
    ? Math.min(720, Math.max(320, 200 + props.group.count * 60))
    : props.expandStyle === 'strip'
      ? Math.min(640, Math.max(280, 208 + Math.max(0, props.group.count - 1) * 40))
      : props.expandStyle === 'spine'
        ? 320
        : props.expandStyle === 'magazine'
          ? 560
          : Math.min(720, Math.max(340, 200 + props.group.count * 44)),
)
const inside = (target: EventTarget | null) =>
  target instanceof Node && (!!card.value?.contains(target) || !!surface.value?.contains(target))
function keep() {
  clearTimeout(closeTimer)
  closeTimer = undefined
}
function track(event: Event) {
  if ('clientX' in event && 'clientY' in event)
    pointer = { x: Number(event.clientX), y: Number(event.clientY) }
}
function pointInside() {
  return (
    !!pointer &&
    [card.value, surface.value].some((el) => {
      const r = el?.getBoundingClientRect()
      return (
        r &&
        pointer!.x >= r.left &&
        pointer!.x <= r.right &&
        pointer!.y >= r.top &&
        pointer!.y <= r.bottom
      )
    })
  )
}
const suppressPreview = computed(() => props.hideSingle === true && props.group.count === 1)
watch(suppressPreview, suppressed => { if (suppressed) close() })
function open(event: PointerEvent | FocusEvent) {
  if (suppressPreview.value) return
  if ('clientX' in event) track(event)
  keep()
  if (visible.value || openTimer) return
  openTimer = setTimeout(() => {
    openTimer = undefined
    if (suppressPreview.value) return
    claimResourceGroupFan(close)
    if (isCardDeck.value) {
      const height = card.value?.getBoundingClientRect().height || 180
      deckOffset.value = -Math.min(220, height * .8)
    }
    visible.value = true
    if (!members.value.length && !loading.value) void load()
  }, 180)
}
function leave(event: PointerEvent | FocusEvent) {
  if ('clientX' in event) track(event)
  if (inside(event.relatedTarget)) return
  clearTimeout(openTimer)
  openTimer = undefined
  keep()
  closeTimer = setTimeout(() => {
    if (!pointInside() && !inside(document.activeElement)) close()
  }, 400)
}
function close() {
  clearTimeout(openTimer)
  openTimer = undefined
  keep()
  visible.value = false
  pointer = undefined
  releaseResourceGroupFan(close)
}
async function load() {
  const version = ++requestVersion
  loading.value = true
  error.value = ''
  try {
    const result = await categoryBrowseServer.members({
      ...props.query,
      categoryId: props.group.categoryId,
      part: props.group.part,
      page: 1,
    })
    if (version !== requestVersion) return
    if (!result.status) throw new Error(result.msg)
    members.value = result.data.dataList
  } catch (e) {
    if (version === requestVersion) error.value = e instanceof Error ? e.message : '加载失败'
  } finally {
    if (version === requestVersion) loading.value = false
  }
}
function select(resource: I_resource) {
  close()
  void playResource(resource)
}
function outside(event: Event) {
  if (!inside(event.target)) close()
}
function scroll(event: Event) {
  if (!inside(event.target)) close()
}
watch(visible, (value) => {
  document[value ? 'addEventListener' : 'removeEventListener']('pointerdown', outside)
  document[value ? 'addEventListener' : 'removeEventListener']('pointermove', track)
  window[value ? 'addEventListener' : 'removeEventListener']('scroll', scroll, true)
  window[value ? 'addEventListener' : 'removeEventListener']('blur', close)
})
watch(
  () => JSON.stringify([props.group, props.query, props.expandStyle]),
  () => {
    requestVersion++
    close()
    members.value = []
    loading.value = false
    error.value = ''
  },
)
onBeforeUnmount(() => {
  requestVersion++
  close()
  document.removeEventListener('pointerdown', outside)
  document.removeEventListener('pointermove', track)
  window.removeEventListener('scroll', scroll, true)
  window.removeEventListener('blur', close)
})
</script>
<style scoped>
.category-card {
  color: var(--el-text-color-primary);
  width: calc(var(--category-cover-width, 162px) + 18px);
  max-width: 100%;
  min-width: 0;
  flex: 0 0 auto;
  padding: 8px;
  box-sizing: border-box;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  background: var(--el-bg-color-overlay);
}
.category-cover {
  position: relative;
  width: 100%;
  height: auto;
  aspect-ratio: var(--category-cover-ratio, 162 / 210);
  padding: 0;
  border: 0;
  border-radius: 7px;
  overflow: hidden;
  background: var(--el-fill-color);
  cursor: pointer;
  display: block;
}
.category-front {
  width: 100%;
  height: 100%;
  display: block;
  position: relative;
  z-index: 1;
}
.category-front.is-stack {
  transform: scale(0.92);
  transform-origin: left bottom;
  border-radius: 6px;
  box-shadow: 2px -2px 5px rgba(0, 0, 0, 0.2);
}
.category-back {
  position: absolute;
  width: 94%;
  height: 94%;
  object-fit: var(--category-cover-fit, cover);
  top: 0;
  right: 0;
  border-radius: 5px;
  opacity: 0.7;
}
.category-back-1 {
  top: 3%;
  right: 3%;
}
.category-part,
.category-count {
  position: absolute;
  z-index: 2;
  border-radius: 5px;
  background: rgba(15, 20, 28, 0.76);
  color: #fff;
  font-size: var(--category-badge-size, 12px);
  padding: 3px 6px;
}
.category-part {
  top: 6px;
  right: 6px;
}
.category-count {
  bottom: 6px;
  right: 6px;
}
.category-caption {
  justify-content: var(--category-caption-align, flex-start);
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 9px;
}
.category-caption strong {
  min-width: 0;
  text-align: var(--category-title-align, left);
  font-size: 14px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.category-meta {
  text-align: var(--category-title-align, left);
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 5px;
}
.category-all {
  margin-top: 5px;
}
.category-preview {
  max-height: calc(100vh - 24px);
  overflow: auto;
}
.category-preview--deck { max-height: none; overflow: visible; }
</style>
