<template>
  <div class="performer-search-shell">
    <div v-if="props.batchMode" class="performer-batch-toolbar">
      <el-button type="primary" icon="Close" @click="emits('batchExit')">退出批量管理</el-button>
      <el-button @click="emits('selectPage')">选择本页</el-button>
      <el-button :loading="props.selectingAll" @click="emits('selectAll')">选择全部筛选结果</el-button>
      <el-button @click="emits('invertPage')">反选本页</el-button>
      <el-button @click="emits('clearSelection')">清空选择</el-button>
      <span class="toolbar-divider" />
      <el-button icon="Star" @click="emits('batchStars')">批量评星</el-button>
      <el-button icon="Switch" @click="emits('batchMigrate')">迁移演员集</el-button>
      <el-button icon="PriceTag" @click="emits('batchTags')">设置演员标签</el-button>
      <el-button type="danger" plain icon="Delete" @click="emits('batchDelete')">批量删除</el-button>
    </div>
    <div class="performer-search">
      <div class="performer-search-filters">
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
      </div>
      <div class="performer-search-actions">
      <el-button icon="DocumentAdd" v-admin v-if="props.admin && !props.batchMode" @click="emits('add')">新增</el-button>
      <el-button icon="Delete" v-admin v-if="props.admin && !props.batchMode" @click="emits('recycleBin')">回收站</el-button>
      <el-button icon="Magnet" v-admin v-if="props.admin && !props.batchMode" @click="emits('scraper')">刮削</el-button>
      <el-button icon="Picture" v-admin v-if="props.admin && !props.batchMode" @click="emits('avatarBatch')">批量匹配头像</el-button>
      <el-button v-if="props.admin && !props.batchMode" v-admin icon="Operation" @click="emits('batchEnter')">批量管理</el-button>
      </div>
    </div>
    <div v-if="props.batchMode" class="selection-status">
      已选择 <strong>{{ props.selectionCount }}</strong> 位演员 · 当前筛选结果 {{ props.total }} 位
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
import { reactive, type PropType } from 'vue';
import type { I_search_performer, T_performerSort } from '@/dataType/performer.dataType';
import type { I_performerTag, PerformerTagMatchMode } from '@/dataType/performerTag.dataType';

const store = { appStoreData: appStoreData() };
const props = defineProps({
  admin: { type: Boolean, default: false },
  performerBasesId: { type: String, default: '' },
  batchMode: { type: Boolean, default: false },
  selectionCount: { type: Number, default: 0 },
  total: { type: Number, default: 0 },
  selectingAll: { type: Boolean, default: false },
  initialSort: { type: String as PropType<T_performerSort>, default: 'createdAtDesc' },
});
const searchData = reactive<I_search_performer>({ search: '', star: '', cup: '', charIndex: '', tagIds: [], tagMatchMode: 'any', sort: props.initialSort });
const activeTags = reactive<I_performerTag[]>([]);
const emits = defineEmits([
  'add', 'recycleBin', 'search', 'scraper', 'avatarBatch', 'batchEnter', 'batchExit',
  'selectPage', 'selectAll', 'invertPage', 'clearSelection', 'batchStars', 'batchMigrate',
  'batchTags', 'batchDelete',
]);
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
.performer-search-filters, .performer-search-actions { display: contents; }
.performer-search-actions > .el-button:nth-child(-n + 2) { order: -1; }
.performer-search-shell { display: flex; flex-direction: column; gap: 6px; }
.performer-search { display: flex; flex-wrap: wrap; gap: 0.5em; }
.performer-search .el-button + .el-button { margin-left: 0; }
.performer-batch-toolbar { display: flex; align-items: center; flex-wrap: wrap; gap: 8px; padding: 8px; border: 1px solid var(--el-border-color); border-radius: 6px; background: var(--el-bg-color-overlay); }
.performer-batch-toolbar .el-button + .el-button { margin-left: 0; }
.toolbar-divider { width: 1px; height: 24px; margin: 0 2px; background: var(--el-border-color); }
.selection-status { padding: 7px 10px; border: 1px solid var(--el-color-primary-light-5); border-radius: 5px; background: var(--el-color-primary-light-9); color: var(--el-color-primary); font-size: 13px; }
.active-tag-filter { display: flex; align-items: center; flex-wrap: wrap; gap: 6px; min-height: 24px; color: var(--el-text-color-secondary); }
:global(html.dark) .performer-batch-toolbar { border-color: #414243; background: #202225; }
:global(html.dark) .selection-status { border-color: #245b86; background: #18222c; color: #79bbff; }
:global(html.bright) .performer-batch-toolbar { border-color: #dcdfe6; background: #ffffff; }
:global(html.bright) .selection-status { border-color: #a0cfff; background: #ecf5ff; color: #337ecc; }
</style>
