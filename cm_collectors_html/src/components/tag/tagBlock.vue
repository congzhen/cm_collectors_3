<template>
  <div class="tag-block">
    <div class="tag-content">
      <tagSpan title="全部"></tagSpan>
      <tagSpan v-for="item, key in dataList" :key="key" :title="item.name"></tagSpan>
    </div>
  </div>
</template>
<script setup lang="ts">
import type { PropType } from 'vue';
import tagSpan from './tagSpan.vue'
import type { E_tagType, I_tagData } from '@/dataType/app.dataType';
const props = defineProps({
  dataList: {
    type: Array as PropType<I_tagData[]>,
    default: () => []
  },
  searchType: {
    type: String as PropType<E_tagType>,
    default: ''
  },
})

</script>
<style lang="scss" scoped>
.tag-block {
  .tag-title {
    font-size: 1.2em;
    font-weight: 500;
    line-height: 2.5em;
  }

  .tag-content {
    display: flex;
    flex-wrap: wrap;
    gap: 0.3em;
  }
}
</style>
