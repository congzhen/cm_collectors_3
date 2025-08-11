<template>
  <el-collapse-item title="Consistency">
    <template #title>
      <tagLogic :text="props.title" :logic="props.logic" :selectedLogic="selectedLogic()"
        @logicClick="logicClickHandle">
      </tagLogic>
    </template>
    <tagBlockPerformer v-if="props.tagType === E_tagType.Performer" @performerClick="tagClickHandle">
    </tagBlockPerformer>
    <tagBlockStar v-else-if="props.tagType === E_tagType.Star" @starClick="tagClickHandle"></tagBlockStar>
    <tagBlock v-else :dataList="props.dataList" @tagClick="tagClickHandle"></tagBlock>
  </el-collapse-item>
</template>
<script setup lang="ts">
import type { PropType } from 'vue'
import tagLogic from './tagLogic.vue'
import tagBlock from './tagBlock.vue'
import tagBlockPerformer from './tagBlockPerformer.vue'
import tagBlockStar from './tagBlockStar.vue'
import { E_tagType, type I_tagData } from '@/dataType/app.dataType'
import { E_searchLogic } from '@/dataType/search.dataType'
import { searchStoreData } from '@/storeData/search.storeData'
const store = {
  searchStoreData: searchStoreData(),
}

const props = defineProps({
  tagType: {
    type: String as PropType<E_tagType>,
    default: E_tagType.DiyTag,
  },
  title: {
    type: String,
    default: '未定义',
  },
  logic: {
    type: Array as PropType<E_searchLogic[]>,
    default: () => [E_searchLogic.Single, E_searchLogic.MultiOr, E_searchLogic.MultiAnd, E_searchLogic.Not],
  },
  dataList: {
    type: Array as PropType<I_tagData[]>,
    default: () => []
  },
  diyTagClassId: {
    type: String,
    default: '',
  }
})

const logicClickHandle = (logic: E_searchLogic) => {
  store.searchStoreData.setLogic(props.tagType, logic, props.diyTagClassId);
}

const tagClickHandle = (option: string) => {
  store.searchStoreData.setQuery(props.tagType, option, props.diyTagClassId);
}


const selectedLogic = (): E_searchLogic => {
  return store.searchStoreData.getLogic(props.tagType, props.diyTagClassId);
}

</script>
<style lang="scss" scoped>
.title-label {
  padding-left: 0.5em;
  font-size: 1em;
  font-weight: 600;
}
</style>
