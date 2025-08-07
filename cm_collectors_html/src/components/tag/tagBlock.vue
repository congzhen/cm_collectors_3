<template>
  <div class="tag-block">
    <div class="tag-content">
      <tagSpan v-for="item, key in props.dataList" :class="[item.status ? 'check' : '']" :key="key" :title="item.name"
        @click="tagClickHandle(item.id)">
      </tagSpan>
    </div>
  </div>
</template>
<script setup lang="ts">
import { type PropType } from 'vue';
import tagSpan from './tagSpan.vue'
import type { I_tagData } from '@/dataType/app.dataType';

const props = defineProps({
  dataList: {
    type: Array as PropType<I_tagData[]>,
    default: () => []
  },
})

const emits = defineEmits(['tagClick'])

const tagClickHandle = (option: string) => {
  emits('tagClick', option)
}

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

    .check {
      background-color: #868686;
      color: #FFF;
    }
  }
}
</style>
