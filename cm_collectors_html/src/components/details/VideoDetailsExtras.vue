<template>
  <section v-if="resource.mode === 'movies'" class="video-extras">
    <MomentCollection :resource-id="resource.id" />
  </section>
</template>
<script setup lang="ts">
import type { I_resource } from '@/dataType/resource.dataType'
import MomentCollection from '@/components/play/MomentCollection.vue'
defineProps<{ resource: I_resource }>()
</script>
<style scoped>
.video-extras {
  min-width: 0;
}
</style>
