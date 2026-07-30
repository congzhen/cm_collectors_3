<template>
  <el-popover v-model:visible="visible" trigger="hover" :show-after="400" :hide-after="200" :width="390"
    placement="bottom-start" @before-enter="handleBeforeEnter">
    <template #reference>
      <div class="search-favorite-trigger" title="搜索收藏">
        <el-icon>
          <CollectionTag />
        </el-icon>
      </div>
    </template>

    <div class="search-favorite-panel" v-loading="loading">
      <div class="panel-header">
        <div>
          <strong>搜索收藏</strong>
          <span>{{ store.appStoreData.currentFilesBases.name || '当前文件库' }}</span>
        </div>
        <el-button type="primary" size="small" :disabled="!hasCurrentConditions" @click="createFavorite">
          保存当前搜索
        </el-button>
      </div>

      <div v-if="!hasCurrentConditions" class="current-hint">
        先在搜索框或左侧栏设置条件，再保存收藏。
      </div>

      <el-scrollbar v-if="favorites.length > 0" max-height="360px">
        <div class="favorite-list">
          <div v-for="favorite in favorites" :key="favorite.id" class="favorite-item">
            <button class="favorite-main" type="button" @click="applyFavorite(favorite)">
              <span class="favorite-summary favorite-summary--title">
                {{ favoriteConditionText(favorite) }}
              </span>
              <span v-if="favorite.invalidConditions > 0" class="favorite-invalid">
                已忽略 {{ favorite.invalidConditions }} 个失效条件
              </span>
            </button>
            <div class="favorite-actions">
              <el-tooltip content="使用" placement="top">
                <el-button link type="primary" icon="Search" @click="applyFavorite(favorite)" />
              </el-tooltip>
              <el-tooltip content="用当前条件覆盖" placement="top">
                <el-button link icon="Refresh" :disabled="!hasCurrentConditions" @click="overwriteFavorite(favorite)" />
              </el-tooltip>
              <el-tooltip content="删除" placement="top">
                <el-button link type="danger" icon="Delete" @click="deleteFavorite(favorite)" />
              </el-tooltip>
            </div>
          </div>
        </div>
      </el-scrollbar>

      <el-empty v-else-if="!loading" description="当前文件库还没有搜索收藏" :image-size="54" />
    </div>
  </el-popover>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { appStoreData } from '@/storeData/app.storeData';
import { searchStoreData } from '@/storeData/search.storeData';
import { E_searchLogic, type I_searchData, type I_searchGroup } from '@/dataType/search.dataType';
import type { I_searchFavorite } from '@/dataType/searchFavorite.dataType';
import { searchFavoriteServer } from '@/server/searchFavorite.server';
import { cacheData } from '@/cache/index.cache';
import { AppLang } from '@/language/app.lang';

const store = {
  appStoreData: appStoreData(),
  searchStoreData: searchStoreData(),
};
const appLang = AppLang();

const visible = ref(false);
const loading = ref(false);
const favorites = ref<I_searchFavorite[]>([]);
let loadedFilesBasesId = '';
let requestVersion = 0;

const currentFilesBasesId = computed(() => store.appStoreData.currentFilesBases.id || '');

const cloneCurrentSearchData = () => (
  JSON.parse(JSON.stringify(store.searchStoreData.searchData)) as I_searchData
);

const groupHasOptions = (group?: I_searchGroup) => Array.isArray(group?.options) && group.options.length > 0;

const hasCurrentConditions = computed(() => {
  const data = store.searchStoreData.searchData;
  if (data.searchTextSlc.length > 0) return true;
  if (
    groupHasOptions(data.country) ||
    groupHasOptions(data.definition) ||
    groupHasOptions(data.videoCodec) ||
    groupHasOptions(data.year) ||
    groupHasOptions(data.star) ||
    groupHasOptions(data.performer) ||
    groupHasOptions(data.cup)
  ) {
    return true;
  }
  return Object.values(data.tag).some(groupHasOptions);
});

watch(currentFilesBasesId, () => {
  requestVersion++;
  loadedFilesBasesId = '';
  favorites.value = [];
  loading.value = false;
  visible.value = false;
});

const loadFavorites = async (force = false) => {
  const filesBasesId = currentFilesBasesId.value;
  if (!filesBasesId || loading.value || (!force && loadedFilesBasesId === filesBasesId)) return;
  const version = ++requestVersion;
  loading.value = true;
  const result = await searchFavoriteServer.list(filesBasesId);
  if (version !== requestVersion || filesBasesId !== currentFilesBasesId.value) return;
  loading.value = false;
  if (result.status) {
    favorites.value = result.data || [];
    loadedFilesBasesId = filesBasesId;
  } else {
    ElMessage.error(result.msg || '读取搜索收藏失败');
  }
};

const handleBeforeEnter = () => {
  loadFavorites();
};

const createFavorite = async () => {
  if (!hasCurrentConditions.value) return;
  const result = await searchFavoriteServer.create({
    filesBasesId: currentFilesBasesId.value,
    searchData: cloneCurrentSearchData(),
  });
  if (result.status) {
    favorites.value.push(result.data);
    loadedFilesBasesId = currentFilesBasesId.value;
    ElMessage.success('搜索收藏已保存');
  } else {
    ElMessage.error(result.msg || '保存搜索收藏失败');
  }
};

const applyFavorite = (favorite: I_searchFavorite) => {
  store.searchStoreData.applySearchData(favorite.searchData);
  visible.value = false;
  if (favorite.invalidConditions > 0) {
    ElMessage.warning(`已应用收藏，并忽略 ${favorite.invalidConditions} 个失效条件`);
  } else {
    ElMessage.success('搜索收藏已应用');
  }
};

const overwriteFavorite = async (favorite: I_searchFavorite) => {
  if (!hasCurrentConditions.value) return;
  try {
    await ElMessageBox.confirm(
      '确定用当前搜索条件覆盖这条收藏吗？',
      '覆盖搜索收藏',
      { type: 'warning', confirmButtonText: '覆盖', cancelButtonText: '取消' },
    );
  } catch {
    return;
  }
  const result = await searchFavoriteServer.update(favorite.id, {
    searchData: cloneCurrentSearchData(),
  });
  if (result.status) {
    Object.assign(favorite, result.data);
    ElMessage.success('搜索收藏已更新');
  } else {
    ElMessage.error(result.msg || '更新搜索收藏失败');
  }
};

const deleteFavorite = async (favorite: I_searchFavorite) => {
  try {
    await ElMessageBox.confirm(
      '确定删除这条搜索收藏吗？',
      '删除搜索收藏',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' },
    );
  } catch {
    return;
  }
  const result = await searchFavoriteServer.delete(favorite.id);
  if (result.status) {
    favorites.value = favorites.value.filter(item => item.id !== favorite.id);
    ElMessage.success('搜索收藏已删除');
  } else {
    ElMessage.error(result.msg || '删除搜索收藏失败');
  }
};

const groupConditionText = (
  group: I_searchGroup | undefined,
  formatOption: (option: string) => string,
) => {
  if (!group?.options?.length) return '';
  const options = group.options.map(formatOption).filter(Boolean);
  if (options.length === 0) return '';
  switch (group.logic) {
    case E_searchLogic.MultiAnd:
      return options.join(' 且 ');
    case E_searchLogic.MultiOr:
      return options.join(' 或 ');
    case E_searchLogic.Not:
      return `非（${options.join('、')}）`;
    default:
      return options.join('、');
  }
};

const favoriteConditionText = (favorite: I_searchFavorite) => {
  const data = favorite.searchData;
  const labels = favorite.optionLabels || {};
  const parts: string[] = [];
  if (data.searchTextSlc?.length) {
    parts.push(`关键词：${data.searchTextSlc.join('、')}`);
  }
  const appendGroup = (
    label: string,
    group: I_searchGroup | undefined,
    formatOption: (option: string) => string,
  ) => {
    const text = groupConditionText(group, formatOption);
    if (text) parts.push(`${label}：${text}`);
  };
  appendGroup('国家', data.country, option => appLang.country(option));
  appendGroup('清晰度', data.definition, option => appLang.definition(option));
  appendGroup('编码', data.videoCodec, option => option.toUpperCase());
  appendGroup('年份', data.year, option => appLang.year(option));
  appendGroup('评分', data.star, option => appLang.stars(option));
  appendGroup('演员', data.performer, option => (
    option === store.searchStoreData.notId
      ? appLang.lang('notPerformer')
      : String(labels[option] || cacheData[option] || option)
  ));
  appendGroup('杯型', data.cup, option => store.appStoreData.cupText(option));
  Object.entries(data.tag || {}).forEach(([classId, group]) => {
    if (!group.options?.length) return;
    const className = store.appStoreData.currentTagClass.find(item => item.id === classId)?.name || '标签';
    appendGroup(className, group, option => (
      labels[option] || store.appStoreData.currentTagInfoById(option)?.name || option
    ));
  });
  return parts.join('；') || '无有效筛选条件';
};
</script>

<style lang="scss" scoped>
.search-favorite-trigger {
  width: 30px;
  height: 30px;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--el-border-color);
  border-radius: var(--el-border-radius-base);
  background: var(--el-fill-color-blank);
  color: var(--el-text-color-regular);
  cursor: pointer;
  transition:
    color 0.2s,
    border-color 0.2s,
    background-color 0.2s;

  &:hover {
    color: var(--el-color-primary);
    border-color: var(--el-color-primary-light-5);
    background: var(--el-color-primary-light-9);
  }
}

.search-favorite-panel {
  min-height: 100px;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--el-border-color-lighter);

  > div {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  strong {
    color: var(--el-text-color-primary);
  }

  span {
    color: var(--el-text-color-secondary);
    font-size: 12px;
  }
}

.current-hint {
  margin-top: 10px;
  padding: 8px 10px;
  border-radius: 4px;
  background: var(--el-fill-color-light);
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.favorite-list {
  display: grid;
  gap: 8px;
  padding-top: 10px;
}

.favorite-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 8px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 5px;

  &:hover {
    border-color: var(--el-border-color);
    background: var(--el-fill-color-light);
  }
}

.favorite-main {
  min-width: 0;
  flex: 1;
  padding: 0;
  border: 0;
  background: transparent;
  color: inherit;
  text-align: left;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.favorite-summary {
  display: block;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.45;
  overflow-wrap: anywhere;
}

.favorite-summary--title {
  color: var(--el-text-color-primary);
  font-size: 13px;
  font-weight: 500;
}

.favorite-invalid {
  color: var(--el-color-warning);
  font-size: 11px;
}

.favorite-actions {
  flex-shrink: 0;
  display: flex;
  align-items: center;

  :deep(.el-button) {
    margin-left: 2px;
    padding: 3px;
  }
}
</style>
