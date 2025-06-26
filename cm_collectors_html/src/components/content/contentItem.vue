<template>
  <contentStyle1 :resource="props.resource"></contentStyle1>
</template>
<script setup lang="ts">
import type { I_resource } from '@/dataType/resource.dataType';
import contentStyle1 from './contentStyle1.vue'
import type { PropType } from 'vue';
const props = defineProps({
  resource: {
    type: Object as PropType<I_resource>,
    required: true,
  },
})

</script>
<style lang="scss" scoped></style>
