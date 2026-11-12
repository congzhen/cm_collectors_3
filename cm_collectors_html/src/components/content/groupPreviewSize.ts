import { computed, onBeforeUnmount, ref, watch } from 'vue';

// All adaptive previews follow the rendered source cover, including responsive resizing.
export function useGroupPreviewSize(source: () => HTMLElement | undefined) {
  const sourceHeight = ref(238);
  const viewportWidth = ref(window.innerWidth);
  const viewportHeight = ref(window.innerHeight);
  let observer: ResizeObserver | undefined;
  const measure = () => {
    sourceHeight.value = source()?.getBoundingClientRect().height || 238;
    viewportWidth.value = window.innerWidth;
    viewportHeight.value = window.innerHeight;
  };
  watch(source, element => {
    observer?.disconnect(); measure();
    if (element && typeof ResizeObserver !== 'undefined') { observer = new ResizeObserver(measure); observer.observe(element); }
  }, { immediate: true });
  window.addEventListener('resize', measure);
  onBeforeUnmount(() => { observer?.disconnect(); window.removeEventListener('resize', measure); });
  const previewHeight = computed(() => Math.min(sourceHeight.value * .9, Math.max(1, viewportHeight.value - 56)));
  return { previewHeight, viewportWidth, viewportHeight };
}
