<template>
  <div class="performer-search-shell">
    <div class="performer-search">
      <el-button icon="DocumentAdd" v-admin v-if="props.admin" @click="emits('add')">新增</el-button>
      <el-button icon="Delete" v-admin v-if="props.admin" @click="emits('recycleBin')">回收站</el-button>
      <inputSearch width="280px" placeholder="请输入姓名、别名、首字母" @change="changeSearchHandle" />
      <selectStar width="200px" @change="changeStarHandle" />
      <selectCup v-if="store.appStoreData.currentConfigApp.plugInUnit_Cup" :search-mode="true" width="200px" @change="changeCupHandle" />
      <performerTagFilter :performer-bases-id="props.performerBasesId" :model-value="searchData.tagIds"
        :match-mode="searchData.tagMatchMode" @change="changeTagHandle" />
      <el-select v-model="searchData.sort" style="width: 180px" @change="emitSearch">
        <el-option label="最新创建" value="createdAtDesc" />
        <el-option label="姓名正序" value="nameAsc" />
        <el-option label="姓名倒序" value="nameDesc" />
        <el-option label="影片数从多到少" value="resourceCountDesc" />
        <el-option label="影片数从少到多" value="resourceCountAsc" />
      </el-select>
      <el-button icon="Magnet" v-admin v-if="props.admin" @click="emits('scraper')">刮削</el-button>
      <el-button icon="Picture" v-admin v-if="props.admin" @click="emits('avatarBatch')">批量匹配头像</el-button>
    </div>
    <div v-if="activeTags.length" class="active-tag-filter">
      <span>已选标签：</span>
      <el-tag v-for="tag in activeTags" :key="tag.id" size="small" closable @close="removeTag(tag.id)">{{ tag.name }}</el-tag>
      <el-button link type="primary" @click="clearTags">清空</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import inputSearch from '../com/form/inputSearch.vue';
import selectStar from '../com/form/selectStar.vue';
import selectCup from '../com/form/selectCup.vue';
import performerTagFilter from './performerTagFilter.vue';
import { appStoreData } from '@/storeData/app.storeData';
import { reactive } from 'vue';
import type { I_search_performer } from '@/dataType/performer.dataType';
import type { I_performerTag, PerformerTagMatchMode } from '@/dataType/performerTag.dataType';

const store = { appStoreData: appStoreData() };
const props = defineProps({ admin: { type: Boolean, default: false }, performerBasesId: { type: String, default: '' } });
const searchData = reactive<I_search_performer>({ search: '', star: '', cup: '', charIndex: '', tagIds: [], tagMatchMode: 'any', sort: 'createdAtDesc' });
const activeTags = reactive<I_performerTag[]>([]);
const emits = defineEmits(['add', 'recycleBin', 'search', 'scraper', 'avatarBatch']);
const emitSearch = () => emits('search', searchData);
const changeSearchHandle = (val: string) => { searchData.search = val; emitSearch(); };
const changeStarHandle = (val: string) => { searchData.star = val; emitSearch(); };
const changeCupHandle = (val: string) => { searchData.cup = val; emitSearch(); };
const changeTagHandle = (value: { tagIds: string[]; tagMatchMode: PerformerTagMatchMode; tags: I_performerTag[] }) => {
  searchData.tagIds = value.tagIds;
  searchData.tagMatchMode = value.tagMatchMode;
  activeTags.splice(0, activeTags.length, ...value.tags);
  emitSearch();
};
const removeTag = (id: string) => {
  searchData.tagIds = searchData.tagIds.filter(item => item !== id);
  const index = activeTags.findIndex(item => item.id === id);
  if (index >= 0) activeTags.splice(index, 1);
  emitSearch();
};
const clearTags = () => { searchData.tagIds = []; activeTags.splice(0); emitSearch(); };
</script>

<style lang="scss" scoped>
.performer-search-shell { display: flex; flex-direction: column; gap: 6px; }
.performer-search { display: flex; flex-wrap: wrap; gap: 0.5em; }
.performer-search .el-button + .el-button { margin-left: 0; }
.active-tag-filter { display: flex; align-items: center; flex-wrap: wrap; gap: 6px; min-height: 24px; color: var(--el-text-color-secondary); }
</style>
