<template>
  <div class="performer-search">
    <el-button icon="DocumentAdd" v-admin v-if="props.admin" @click="emits('add')">新增</el-button>
    <el-button icon="Delete" v-admin v-if="props.admin" @click="emits('recycleBin')">回收站</el-button>
    <inputSearch width="280px" placeholder="请输入姓名、别名、首字母" @change="changeSearchHandle" />
    <selectStar width="200px" @change="changeStarHandle" />
    <selectCup v-if="store.appStoreData.currentConfigApp.plugInUnit_Cup" :search-mode="true" width="200px"
      @change="changeCupHandle" />
    <el-select v-model="searchData.sort" style="width: 180px" @change="changeSortHandle">
      <el-option label="最新创建" value="createdAtDesc" />
      <el-option label="影片数从多到少" value="resourceCountDesc" />
      <el-option label="影片数从少到多" value="resourceCountAsc" />
    </el-select>
    <el-button icon="Magnet" v-admin v-if="props.admin" @click="emits('scraper')">刮削</el-button>
    <el-button icon="Picture" v-admin v-if="props.admin" @click="emits('avatarBatch')">批量匹配头像</el-button>
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
  admin: {
    type: Boolean,
    default: false,
  },
})

const searchData = reactive<I_search_performer>({
  search: '',
  star: '',
  cup: '',
  charIndex: '',
  sort: 'createdAtDesc',
})

const emits = defineEmits(['add', 'recycleBin', 'search', 'scraper', 'avatarBatch'])


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

const changeSortHandle = () => {
  emits('search', searchData)
}

</script>
<style lang="scss" scoped>
.performer-search {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5em;

  .el-button+.el-button {
    margin-left: 0px;
  }
}
</style>
