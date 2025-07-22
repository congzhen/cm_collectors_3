<template>
  <div class="tag-block-performer">
    <div class="tag-content">
      <tagSpan :title="store.searchStoreData.allName" @click="performerItemClickHandle(store.searchStoreData.allId)"
        :class="[checkStatus(store.searchStoreData.allId) ? 'check' : '']">
      </tagSpan>
      <tagSpan :title="store.searchStoreData.notPerformer"
        @click="performerItemClickHandle(store.searchStoreData.notId)"
        :class="[checkStatus(store.searchStoreData.notId) ? 'check' : '']">
      </tagSpan>

    </div>
    <div class="tag-performer">
      <performerBlock class="tag-performer-item"
        v-for="performer, key in store.appStoreData.currentTopPreferredPerformers" :key="key" :performer="performer"
        :class="[checkStatus(performer.id) ? 'check' : '']" @click="performerObjectClickHandle(performer)">
      </performerBlock>
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

const store = {
  appStoreData: appStoreData(),
  searchStoreData: searchStoreData()
}
const emits = defineEmits(['performerClick'])

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

  .check {
    background-color: #868686 !important;
    color: #FFF;
  }
}
</style>
