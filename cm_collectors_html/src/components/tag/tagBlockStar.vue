<template>
  <div class="tag-block-stars">
    <div class="tag-content">
      <tagSpan :title="store.searchStoreData.allName" :tagModeFixed="true"
        @click="starItemClickHandle(store.searchStoreData.allId)"
        :class="[checkStatus(store.searchStoreData.allId) ? 'check' : '']">
      </tagSpan>
      <tagSpan :title="store.searchStoreData.notStar" :tagModeFixed="true" @click="starItemClickHandle('0')"
        :class="[checkStatus('0') ? 'check' : '']">
      </tagSpan>
    </div>
    <ul class="tag-stars">
      <li v-for="(star, index) in starData" :key="index" :class="[checkStatus(star.toString()) ? 'check' : '']">
        <el-rate v-model="starData[index]" disabled @click="starItemClickHandle(star.toString())" />
      </li>
    </ul>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue'
import tagSpan from './tagSpan.vue'
import { searchStoreData } from '@/storeData/search.storeData'
import { E_tagType } from '@/dataType/app.dataType'
const store = {
  searchStoreData: searchStoreData()
}
const emits = defineEmits(['starClick'])
const starData = ref([1, 2, 3, 4, 5])

const starItemClickHandle = (data: string) => {
  emits('starClick', data)
}
const checkStatus = (data: string) => {
  return store.searchStoreData.checkSelected(E_tagType.Star, data);
}

</script>
<style lang="scss" scoped>
.tag-block-stars {
  display: flex;

  .tag-content {
    display: flex;
    flex-wrap: wrap;
    flex-direction: column;
    gap: 0.3em;
  }

  .tag-stars {
    padding-left: 0.3em;
    list-style-type: none;
    cursor: pointer;

    li {
      background-color: #303131;
      padding: 0.4em 0.6em;
      margin-bottom: 0.3em;
      border-radius: 3px;
      height: 1em;
      line-height: 1em;
      overflow: hidden;

      &:hover {
        background-color: var(--el-color-primary-light-9);
      }

      .el-rate {
        height: 1em;
        cursor: pointer;

        :deep(.el-rate__item) {
          cursor: pointer;
        }
      }
    }
  }

  .check {
    background-color: #868686 !important;
    color: #FFF;
  }
}
</style>
