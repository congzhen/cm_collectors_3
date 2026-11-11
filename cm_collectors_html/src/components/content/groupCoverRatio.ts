import { reactive } from 'vue';
import type { I_resource } from '@/dataType/resource.dataType';
import { getResourceCoverPoster } from '@/common/photo';

// Prefer actual image dimensions; metadata provides a stable initial layout.
export function useGroupCoverRatio() {
  const loaded = reactive(new Map<string, number>());
  const ratio = (resource: I_resource) => loaded.get(getResourceCoverPoster(resource))
    || (resource.coverPosterWidth > 0 && resource.coverPosterHeight > 0
      ? resource.coverPosterWidth / resource.coverPosterHeight : 2 / 3);
  const onLoad = (resource: I_resource, event: Event) => {
    const image = event.target as HTMLImageElement;
    if (image.naturalWidth > 0 && image.naturalHeight > 0)
      loaded.set(getResourceCoverPoster(resource), image.naturalWidth / image.naturalHeight);
  };
  return { ratio, onLoad };
}
