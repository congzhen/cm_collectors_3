<template>
  <section class="group-fan" :class="{ 'is-dense': (loading || error ? expectedCount || members.length : members.length) > 7, 'group-fan--strip': expandStyle === 'strip' }" aria-label="关联组资源" :aria-busy="loading">
    <div v-if="loading || error" class="fan-state">
      <el-icon v-if="loading" class="is-loading"><Loading /></el-icon>
      <span>{{ error || '正在展开关联资源…' }}</span>
      <el-button v-if="error" link @click="emit('retry')">重试</el-button>
    </div>
    <div v-else-if="members.length" class="fan-cards">
      <div ref="deckRef" class="fan-deck" @pointermove="pointToCard" @keydown="navigateCards">
        <button v-for="(resource, index) in members" :key="resource.id" type="button"
          class="fan-card" :class="{ 'is-active': activeIndex === index }" :style="cardStyle(index)"
          :aria-label="`打开资源：${resource.title}`" :title="resource.title"
          :tabindex="activeIndex === index ? 0 : -1"
          @focus="activeIndex = index" @pointerdown.left="pressCard" @pointercancel="cancelCardPress" @click="selectCard($event, resource)">
          <el-image :src="getResourceCoverPoster(resource)" fit="cover" loading="lazy">
            <template #error><span class="fan-cover-empty"><el-icon><Picture /></el-icon></span></template>
          </el-image>
          <span class="fan-card-title">{{ resource.title }}</span>
        </button>
      </div>
    </div>
    <div v-else class="fan-state">暂无关联资源</div>
  </section>
</template>

<script setup lang="ts">
import { nextTick, ref, watch } from 'vue';
import { Loading, Picture } from '@element-plus/icons-vue';
import { getResourceCoverPoster } from '@/common/photo';
import type { I_resource } from '@/dataType/resource.dataType';
const props = defineProps<{ members: I_resource[]; loading: boolean; error: string; expectedCount?: number; expandStyle?: 'fan' | 'strip'; matchedIds?: Set<string> }>();
const emit = defineEmits<{ select: [resource: I_resource]; retry: [] }>();
const activeIndex = ref(0);
let pressedResource: I_resource | undefined;
const cancelCardPress = () => { pressedResource = undefined; };
const deckRef = ref<HTMLElement>();
const navigateCards = async (event: KeyboardEvent) => {
  if (!['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key) || !props.members.length) return;
  event.preventDefault();
  event.stopPropagation();
  const last = props.members.length - 1;
  activeIndex.value = event.key === 'Home' ? 0 : event.key === 'End' ? last
    : Math.max(0, Math.min(last, activeIndex.value + (event.key === 'ArrowRight' ? 1 : -1)));
  await nextTick();
  (deckRef.value?.children[activeIndex.value] as HTMLElement | undefined)?.focus({ preventScroll: true });
};
watch(() => props.members, () => { cancelCardPress(); activeIndex.value = Math.floor(props.members.length / 2); }, { immediate: true });
const pointToCard = (event: { clientX: number }) => {
  const deck = deckRef.value;
  if (!deck) return;
  const bounds = deck.getBoundingClientRect();
  const x = event.clientX - bounds.left;
  const count = props.members.length;
  if (!count) return;
  const firstCard = deck.firstElementChild as HTMLElement;
  // Integer offset/client widths accumulate visible hit errors in very dense groups.
  const deckWidth = bounds.width || deck.clientWidth;
  const cardWidth = Number.parseFloat(getComputedStyle(firstCard).width) || firstCard.offsetWidth;
  const strip = props.expandStyle === 'strip';
  const step = Math.min(strip ? 48 : 74, (deckWidth - cardWidth - (strip ? 48 : 112)) / Math.max(1, count - 1));
  if (step <= 0 || count === 1) { activeIndex.value = 0; return; }
  const position = (x - deckWidth / 2) / step + (count - 1) / 2;
  // A small dead zone prevents tiny pointer movements from toggling adjacent cards.
  if (Math.abs(position - activeIndex.value) < .6) return;
  const nearest = Math.max(0, Math.min(count - 1, Math.round(position)));
  activeIndex.value = nearest;
};
const pressCard = (event: PointerEvent) => {
  event.preventDefault();
  pointToCard(event);
  pressedResource = props.members[activeIndex.value];
};
const selectCard = (event: MouseEvent, resource: I_resource) => {
  // The popup can appear underneath a stationary pointer, before any pointermove.
  // Mouse click coordinates may be rounded; preserve the precise pointerdown target.
  if (event.detail && !pressedResource) pointToCard(event);
  const selected = event.detail ? pressedResource || props.members[activeIndex.value] : resource;
  cancelCardPress();
  if (selected) emit('select', selected);
};
const cardStyle = (index: number) => {
  const count = props.members.length;
  const position = index - (count - 1) / 2;
  const relative = count > 1 ? position / ((count - 1) / 2) : 0;
  return {
    left: `calc(50% + ${position} * min(var(--fan-max-step), (100% - var(--fan-card-width) - var(--fan-edge-space)) / ${Math.max(1, count - 1)}))`,
    '--card-angle': `${props.expandStyle === 'strip' ? 0 : relative * Math.min(14, (count - 1) * 3.5)}deg`,
    '--card-drop': `${props.expandStyle === 'strip' ? 0 : Math.abs(relative) * 16}px`,
    zIndex: activeIndex.value === index ? count + 1 : count - Math.abs(index - activeIndex.value),
  };
};
</script>

<style scoped lang="scss">
.group-fan {
  --fan-max-step: 74px; --fan-edge-space: 112px;
  --fan-card-width: clamp(110px, 28vw, 160px); --fan-card-height: 226px;
  --fan-bg: #282c32; --fan-text: #f4f5f7; --fan-border: rgba(255,255,255,.4);
  --fan-active-border: #e5eaf0; --fan-shadow: rgba(0,0,0,.4); --fan-active-shadow: rgba(0,0,0,.5);
  color: var(--fan-text);
}
:global(html.bright) .group-fan {
  --fan-bg: #fff; --fan-text: #34404c; --fan-border: rgba(67,82,98,.3);
  --fan-active-border: #fff; --fan-shadow: rgba(37,52,70,.2); --fan-active-shadow: rgba(37,52,70,.3);
}
.fan-cards { max-height: calc(100vh - 24px); overflow-y: auto; scrollbar-width: thin; }
.group-fan.is-dense { --fan-card-width: clamp(100px, 24vw, 132px); --fan-card-height: 188px; }
.group-fan--strip { --fan-max-step: 48px; --fan-edge-space: 48px; }
.group-fan--strip .fan-card { border-radius: 8px; box-shadow: 0 6px 14px var(--fan-shadow); }
.group-fan--strip .fan-card:not(.is-active) .fan-card-title { opacity: 0; }
.fan-deck { position: relative; height: calc(var(--fan-card-height) + 84px); }
.fan-card {
  position: absolute; top: 32px; width: var(--fan-card-width); height: var(--fan-card-height); padding: 0;
  box-sizing: border-box;
  border: 1px solid var(--fan-border); border-radius: 10px; overflow: hidden;
  color: var(--fan-text); background: var(--fan-bg); cursor: pointer;
  box-shadow: 0 10px 24px var(--fan-shadow);
  transform: translateX(-50%) translateY(var(--card-drop)) rotate(var(--card-angle));
  transform-origin: 50% 90%;
  transition: transform .22s cubic-bezier(.2,.8,.2,1), border-color .18s, box-shadow .18s;
  .el-image { position: absolute; inset: 0; width: 100%; height: 100%; background: var(--fan-bg); }
  &.is-active { transform: translateX(-50%) translateY(-10px) rotate(0) scale(1.04); border-color: var(--fan-active-border); box-shadow: 0 14px 26px var(--fan-active-shadow); }
  &:focus-visible { outline: 3px solid var(--el-color-primary); outline-offset: 3px; }
}
.fan-card-title { position: absolute; left: 0; right: 0; bottom: 0; padding: 24px 8px 8px; color: white; background: linear-gradient(transparent, rgba(0,0,0,.7)); font-size: 12px; line-height: 18px; text-overflow: ellipsis; overflow: hidden; white-space: nowrap; text-align: left; pointer-events: none; }
.group-fan--strip .fan-card.is-active { transform: translateX(-50%) translateY(-6px) scale(1.02); box-shadow: 0 8px 18px var(--fan-active-shadow); }
.fan-cover-empty { height: 100%; display: grid; place-items: center; font-size: 32px; color: #8b939e; }
.fan-state { display: flex; justify-content: center; align-items: center; gap: 10px; height: calc(var(--fan-card-height) + 84px); max-height: calc(100vh - 24px); box-sizing: border-box; padding: 16px 24px; font-size: 13px; }
.fan-state > span { padding: 12px 16px; border: 1px solid var(--fan-border); border-radius: 12px; background: var(--fan-bg); box-shadow: 0 6px 20px var(--fan-shadow); }
@media (prefers-reduced-motion: reduce) { .fan-card { transition: none; } }
</style>
