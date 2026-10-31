<template>
  <div v-if="grouped && style !== 'stack'" class="el-image group-cover" :class="[`group-cover--${style}`, `group-cover--count-${others.length + 1}`]" :style="style === 'adaptive' ? { '--group-main-width': `${others.length < 3 ? 65 : others.length < 5 ? 60 : 55}%` } : undefined">
    <el-image class="group-cover-main" :src="src" fit="cover" @load="emit('load', $event)" />
    <img class="group-cover-sizer" :src="src" alt="" aria-hidden="true" decoding="async" />
    <div class="group-cover-leaves" aria-hidden="true">
      <div v-for="(member, index) in others" :key="member.id" class="group-cover-leaf" :class="`group-cover-leaf--${index}`">
        <el-image :src="getResourceCoverPoster(member)" fit="cover" loading="lazy" />
      </div>
    </div>
  </div>
  <el-image v-else :src="src" :fit="fit" @load="emit('load', $event)" />
</template>
<script setup lang="ts">
import { computed, inject } from 'vue';
import { appStoreData } from '@/storeData/app.storeData';
import { getResourceCoverPoster } from '@/common/photo';
import type { I_resource } from '@/dataType/resource.dataType';
import { resourceGroupContextKey } from './resourceGroups';
const props = withDefaults(defineProps<{ resource: I_resource; src: string; fit?: string }>(), { fit: 'cover' });
const emit = defineEmits<{ load: [event: Event] }>();
const context = inject(resourceGroupContextKey, undefined);
const style = computed(() => {
  const value = appStoreData().currentConfigApp.resourceGroupCoverStyle;
  return value === 'triptych' || value === 'book' || value === 'adaptive' ? value : 'stack';
});
const grouped = computed(() => !!context?.enabled.value && !!props.resource.associationGroupId && context.pageIds.value.has(props.resource.id));
// Composition follows member count, not card dimensions: seven or more always
// retain the main cover plus six auxiliaries, even in compact layouts.
const memberLimit = computed(() => style.value === 'adaptive' ? 6 : 2);
const others = computed(() => {
  const preview = context?.summaries.value.get(props.resource.associationGroupId || '')?.preview || [];
  const members = preview.filter(member => member.id !== props.resource.id).slice(0, memberLimit.value);
  // A partial search result still needs to look grouped while its summary loads.
  return members.length ? members : [props.resource];
});
</script>
<style scoped>
.group-cover { display: block; position: relative; width: 100%; overflow: hidden; background: var(--el-fill-color-dark); }
.group-cover-sizer { display: block; width: 100%; height: auto; visibility: hidden; pointer-events: none; object-fit: cover; }
.group-cover > .group-cover-main { position: absolute; inset: 0 auto 0 0; display: block; width: 70% !important; height: 100% !important; transform: none !important; scale: 1 !important; }
.group-cover-leaves { position: absolute; inset: 0 0 0 70%; display: flex; gap: 2px; padding-left: 2px; pointer-events: none; }
.group-cover-leaf { position: relative; flex: 1; min-width: 0; overflow: hidden; }
.group-cover-leaf::after { content: ''; position: absolute; inset: 0; background: linear-gradient(90deg, rgba(0,0,0,.22), transparent 45%, rgba(0,0,0,.08)); }
.group-cover-leaf :deep(.el-image) { display: block; width: 100% !important; height: 100% !important; transform: none !important; scale: 1 !important; }
.group-cover-leaf :deep(.el-image__inner) { object-fit: cover !important; }
.group-cover--book > .group-cover-main { width: 90% !important; border-radius: 0 5px 5px 0; }
.group-cover--book .group-cover-leaves { left: 90%; gap: 0; padding-left: 1px; }
.group-cover--book .group-cover-leaf { margin-block: 5px; border-radius: 0 6px 6px 0; box-shadow: inset 1px 0 0 rgba(255,255,255,.35); }
.group-cover--book .group-cover-leaf--1 { margin-block: 10px; }
.group-cover--book .group-cover-leaf::after { background: linear-gradient(90deg, rgba(0,0,0,.42), rgba(255,255,255,.12) 35%, rgba(0,0,0,.2)); border: 1px solid var(--el-border-color); border-left: 0; border-radius: inherit; }
.group-cover--adaptive > .group-cover-main { width: var(--group-main-width) !important; }
.group-cover--adaptive .group-cover-leaves { left: var(--group-main-width); display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); grid-template-rows: repeat(2, minmax(0, 1fr)); }
.group-cover--adaptive .group-cover-leaf { min-height: 0; }
.group-cover--adaptive .group-cover-leaf::after { background: rgba(0,0,0,.08); }
.group-cover--adaptive.group-cover--count-2 .group-cover-leaves { grid-template-columns: 1fr; grid-template-rows: 1fr; }
.group-cover--adaptive.group-cover--count-3 .group-cover-leaves { grid-template-columns: 1fr; }
.group-cover--adaptive.group-cover--count-4 .group-cover-leaf--0 { grid-column: span 2; }
.group-cover--adaptive.group-cover--count-6 .group-cover-leaves,
.group-cover--adaptive.group-cover--count-7 .group-cover-leaves { grid-template-rows: repeat(3, minmax(0, 1fr)); }
.group-cover--adaptive.group-cover--count-6 .group-cover-leaf--4 { grid-column: span 2; }
</style>
