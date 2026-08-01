<template>
  <div class="tag-block-performer"
    :class="{ 'tag-block-performer--photo': store.appStoreData.currentConfigApp.performerPhoto,
      'tag-block-performer--modern': props.appearanceStyle === 'modern' }">
    <div class="tag-content">
      <tagSpan :title="store.searchStoreData.allName" @click="performerItemClickHandle(store.searchStoreData.allId)"
        :class="[checkStatus(store.searchStoreData.allId) ? 'check' : '']">
      </tagSpan>
      <tagSpan :title="$t('notPerformer')" @click="performerItemClickHandle(store.searchStoreData.notId)"
        :class="[checkStatus(store.searchStoreData.notId) ? 'check' : '']">
      </tagSpan>

    </div>
    <div class="tag-performer" v-if="store.appStoreData.currentConfigApp.performerPhoto">
      <performerBlock class="tag-performer-item"
        v-for="performer, key in store.appStoreData.currentTopPreferredPerformers" :key="key" :performer="performer"
        :modern="props.appearanceStyle === 'modern'"
        :class="[checkStatus(performer.id) ? 'check' : '']" @click="performerObjectClickHandle(performer)"
        :style="{ width: width_C }">
      </performerBlock>
    </div>
    <div class="tag-performer" v-else>
      <tagSpan v-for="performer, key in store.appStoreData.currentTopPreferredPerformers"
        :class="[checkStatus(performer.id) ? 'check' : '']" :key="key" :title="performer.name"
        @click="performerObjectClickHandle(performer)">
      </tagSpan>
    </div>
  </div>
</template>
<script setup lang="ts">
import tagSpan from './tagSpan.vue'
import performerBlock from '../performer/performerBlock.vue'
import { appStoreData } from '@/storeData/app.storeData';
import { searchStoreData } from '@/storeData/search.storeData'
import { E_tagType } from '@/dataType/app.dataType'
import type { I_performer } from '@/dataType/performer.dataType';
import { cacheData } from '@/cache/index.cache'
import { computed, type PropType } from 'vue';

const store = {
  appStoreData: appStoreData(),
  searchStoreData: searchStoreData()
}
const emits = defineEmits(['performerClick'])
const props = defineProps({
  appearanceStyle: {
    type: String as PropType<'classic' | 'modern'>,
    default: 'classic',
  },
})


const width_C = computed(() => {
  let rowNumRatio = 25;
  try {
    rowNumRatio = 100 / store.appStoreData.currentConfigApp.tagFixedModeRowShowNum;
  } catch {
    rowNumRatio = 25;
  }
  return `calc(${rowNumRatio}% - ${store.appStoreData.currentConfigApp.tagFixedModeRowShowNum * 0.1}em)`
  //return `calc(${rowNumRatio}% - 0.9em)`
})


const performerItemClickHandle = (data: string) => {
  emits('performerClick', data)
}

const performerObjectClickHandle = (performer: I_performer) => {
  cacheData[performer.id] = performer.name;
  store.searchStoreData.setQuery(E_tagType.Performer, performer.id);
}

const checkStatus = (data: string) => {
  return store.searchStoreData.checkSelected(E_tagType.Performer, data);
}


</script>
<style lang="scss" scoped>
.tag-block-performer {
  .tag-content {
    display: flex;
    flex-wrap: wrap;
    gap: 0.3em;
  }

  .tag-performer {
    padding: 0.5em 0;
    display: flex;
    flex-wrap: wrap;
    gap: 0.3em;

    .tag-performer-item {
      width: calc(25% - 0.9em);

      &:hover {
        background-color: var(--el-color-primary-light-9);
      }
    }
  }

  &:not(.tag-block-performer--modern) .check {
    background-color: #f2b75b !important;
    color: #3f2b08;
    box-shadow: inset 0 0 0 1px #d4932f, 0 2px 7px rgba(125, 83, 18, 0.18);
  }
}
</style>
