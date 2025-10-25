<template>
  <div class="details-tag-list">
    <el-tag type="info" effect="plain" size="large" v-for="item, key in tags_C" :key="key">
      {{ item.name }}
    </el-tag>
  </div>
</template>
<script lang="ts" setup>
import type { I_resource } from '@/dataType/resource.dataType';
import { computed, type PropType } from 'vue';
const props = defineProps({
  resource: {
    type: Object as PropType<I_resource>,
    required: true
  },
})

const tags_C = computed(() => {
  return props.resource.tags
})

</script>
<style lang="scss" scoped>
.details-tag-list {
  padding: 5px;
  display: flex;
  flex-wrap: wrap;
  gap: 5px;
}
</style>
