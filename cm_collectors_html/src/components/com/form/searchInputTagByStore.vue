<template>
  <searchInputTag v-model="store.searchStoreData.searchData.searchTextSlc" placeholder="请输入搜索内容" clearable
    @clear-click="clearHandle">
    <template #tag>
      <el-tag v-for="(tag, index) in store.searchStoreData.searchData.country.options" :key="index" closable
        :type="getElTagType(store.searchStoreData.searchData.country.logic)"
        @close="closeHandle(store.searchStoreData.searchData.country.options, index)">
        {{ tag }}
      </el-tag>
      <el-tag v-for="(tag, index) in store.searchStoreData.searchData.definition.options" :key="index" closable
        :type="getElTagType(store.searchStoreData.searchData.definition.logic)"
        @close="closeHandle(store.searchStoreData.searchData.definition.options, index)">
        {{ tag }}
      </el-tag>
      <el-tag v-for="(year, index) in store.searchStoreData.searchData.year.options" :key="index" closable
        :type="getElTagType(store.searchStoreData.searchData.year.logic)"
        @close="closeHandle(store.searchStoreData.searchData.year.options, index)">
        {{ getYearText(year) }}
      </el-tag>
      <el-tag v-for="(star, index) in store.searchStoreData.searchData.star.options" :key="index" closable
        :type="getElTagType(store.searchStoreData.searchData.star.logic)"
        @close="closeHandle(store.searchStoreData.searchData.star.options, index)">
        {{ getStarText(star) }}
      </el-tag>
      <el-tag v-for="(performerId, index) in store.searchStoreData.searchData.performer.options" :key="index" closable
        :type="getElTagType(store.searchStoreData.searchData.performer.logic)"
        @close="closeHandle(store.searchStoreData.searchData.performer.options, index)">
        {{ getPerformerText(performerId) }}
      </el-tag>
      <el-tag v-for="(cup, index) in store.searchStoreData.searchData.cup.options" :key="index" closable
        :type="getElTagType(store.searchStoreData.searchData.cup.logic)"
        @close="closeHandle(store.searchStoreData.searchData.cup.options, index)">
        {{ store.appStoreData.cupText(cup) }}
      </el-tag>

      <template v-for="tagClassObj, tagClassIndex in store.searchStoreData.searchData.tag" :key="tagClassIndex">
        <el-tag v-for="(tag, index) in tagClassObj.options" :key="index" closable
          :type="getElTagType(tagClassObj.logic)" @close="closeTagHandle(tag)">
          {{ store.appStoreData.currentTagInfoById(tag)?.name || tag }}
        </el-tag>
      </template>
    </template>
  </searchInputTag>
</template>
<script lang="ts" setup>
import searchInputTag from './searchInputTag.vue';
import { appStoreData } from '@/storeData/app.storeData';
import { searchStoreData } from '@/storeData/search.storeData';
import { cacheData } from '@/cache/index.cache';
import { E_searchLogic } from '@/dataType/search.dataType';
import { appLang } from '@/language/app.lang'
const store = {
  appStoreData: appStoreData(),
  searchStoreData: searchStoreData(),
};
const closeHandle = (options: string[], index: number) => {
  options.splice(index, 1);
}

const closeTagHandle = (tagId: string) => {
  store.searchStoreData.deleteDiyTagById(tagId);
}

const clearHandle = () => {
  store.searchStoreData.clear();
}

const getElTagType = (logic: E_searchLogic) => {
  switch (logic) {
    case E_searchLogic.Single:
      return 'success';
    case E_searchLogic.MultiAnd:
      return 'warning';
    case E_searchLogic.MultiOr:
      return 'primary';
    case E_searchLogic.Not:
      return 'danger';
  }
}

const getYearText = (year: string) => {
  return appLang.year(year);
}

const getStarText = (star: string) => {
  return appLang.stars(star);
}

const getPerformerText = (performerId: string) => {
  if (performerId == store.searchStoreData.notId) {
    return appLang.lang('notPerformer')
  } else {
    return cacheData[performerId] || performerId
  }
}
</script>
