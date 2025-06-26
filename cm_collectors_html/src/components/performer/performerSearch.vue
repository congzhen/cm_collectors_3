<template>
  <div class="performer-search">
    <el-button icon="DocumentAdd" v-if="props.addBtn" @click="emits('add')">新增</el-button>
    <inputSearch width="280px" placeholder="请输入姓名、别名、首字母" @change="changeSearchHandle" />
    <selectStar width="200px" @change="changeStarHandle" />
    <selectCup v-if="store.appStoreData.currentConfigApp.plugInUnit_Cup" width="200px" @change="changeCupHandle" />
  </div>
</template>
<script setup lang="ts">
import inputSearch from '../com/form/inputSearch.vue'
import selectStar from '../com/form/selectStar.vue'
import selectCup from '../com/form/selectCup.vue'
import { appStoreData } from '@/storeData/app.storeData';
import { reactive } from 'vue'
import type { I_search_performer } from '@/dataType/performer.dataType';
const store = {
  appStoreData: appStoreData(),
}
const props = defineProps({
  addBtn: {
    type: Boolean,
    default: false,
  },
})

const searchData = reactive<I_search_performer>({
  search: '',
  star: '',
  cup: '',
})

const emits = defineEmits(['add', 'search'])


const changeSearchHandle = (val: string) => {
  searchData.search = val
  emits('search', searchData)
}
const changeStarHandle = (val: string) => {
  searchData.star = val
  emits('search', searchData)
}
const changeCupHandle = (val: string) => {
  searchData.cup = val
  emits('search', searchData)
}

</script>
<style lang="scss" scoped>
.performer-search {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5em;
}
</style>
