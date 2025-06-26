<template>
  <el-collapse-item title="Consistency">
    <template #title>
      <tagLogic :text="props.title"></tagLogic>
    </template>
    <tagBlockPerformer v-if="props.tagType === E_tagType.Performer"></tagBlockPerformer>
    <tagBlockStar v-else-if="props.tagType === E_tagType.Star"></tagBlockStar>
    <tagBlock v-else :dataList="props.dataList" :searchType="props.tagType"></tagBlock>
  </el-collapse-item>
</template>
<script setup lang="ts">
import type { PropType } from 'vue'
import tagLogic from './tagLogic.vue'
import tagBlock from './tagBlock.vue'
import tagBlockPerformer from './tagBlockPerformer.vue'
import tagBlockStar from './tagBlockStar.vue'
import { E_tagType, type I_tagData } from '@/dataType/app.dataType'

const props = defineProps({
  tagType: {
    type: String as PropType<E_tagType>,
    default: E_tagType.DiyTag,
  },
  title: {
    type: String,
    default: '未定义',
  },
  dataList: {
    type: Array as PropType<I_tagData[]>,
    default: () => []
  },
})



</script>
<style lang="scss" scoped>
.title-label {
  padding-left: 0.5em;
  font-size: 1em;
  font-weight: 600;
}
</style>
