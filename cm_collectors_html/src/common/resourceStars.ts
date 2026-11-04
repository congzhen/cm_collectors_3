import { computed, inject, onScopeDispose, reactive, watch } from 'vue';
import type { I_resource } from '@/dataType/resource.dataType';
import { isMobile, isMobileDevice } from '@/assets/mobile';
import { appStoreData } from '@/storeData/app.storeData';
import { resourceSelectionKey } from '@/components/content/resourceSelection';

// Track live resource copies without retaining scores across reloads or libraries.
const resources = new Map<symbol, I_resource[]>();
export const pendingResourceStars = reactive(new Set<string>());
export const resourceStarsKey = (resource: I_resource) => `${resource.filesBases_id}:${resource.id}`;
let activeQuickStarsCloser: (() => void) | undefined;
export function claimQuickStarsPanel(close: () => void) {
  if (activeQuickStarsCloser !== close) activeQuickStarsCloser?.();
  activeQuickStarsCloser = close;
}
export function releaseQuickStarsPanel(close: () => void) {
  if (activeQuickStarsCloser === close) activeQuickStarsCloser = undefined;
}

export function trackResourceStars(getResources: () => I_resource[]) {
  const key = Symbol();
  watch(() => [...getResources()], value => resources.set(key, value), { immediate: true, flush: 'sync' });
  onScopeDispose(() => resources.delete(key));
}

export function syncResourceStars(resource: I_resource, stars: number) {
  const key = resourceStarsKey(resource);
  resource.stars = stars;
  for (const list of resources.values()) {
    for (const item of list) {
      if (resourceStarsKey(item) === key) item.stars = stars;
    }
  }
}

export function useQuickResourceStars(resource?: () => I_resource) {
  const store = appStoreData();
  const selection = inject(resourceSelectionKey, undefined);
  const enabled = computed(() => !selection?.active.value && store.currentConfigApp.resourceQuickStars === true && !isMobile() && !isMobileDevice()
    && !/iPhone|iPod|Android.*Mobile/i.test(navigator.userAgent));
  const editable = computed(() => enabled.value && store.displayAdminFn
    && (!resource || resource().filesBases_id === store.currentFilesBases.id));
  return { enabled, editable };
}
