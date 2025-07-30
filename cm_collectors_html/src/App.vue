<template>
  <div class="app-container" v-if="initStatus">
    <router-view v-slot="{ Component }">
      <keep-alive exclude="playMovies,playComic,playAtlas">
        <component :is="Component" :key="$route.fullPath" />
      </keep-alive>
    </router-view>
  </div>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { appStoreData } from '@/storeData/app.storeData'
import { filesBasesStoreData } from '@/storeData/filesBases.storeData'
import { performerBasesStoreData } from '@/storeData/performerBases.storeData';
import { searchStoreData } from './storeData/search.storeData';
import { LoadingService } from '@/assets/loading'
import { ElMessage } from 'element-plus'
const initStatus = ref(false)

const store = {
  appStoreData: appStoreData(),
  filesBasesStoreData: filesBasesStoreData(),
  performerBasesStoreData: performerBasesStoreData(),
  searchStoreData: searchStoreData(),
}

const init = async () => {
  try {
    LoadingService.show()
    const result = await store.appStoreData.initApp();
    if (result && !result.status) {
      ElMessage.error(result.message);
      return
    }
    const firstFilesBases = store.filesBasesStoreData.filesBasesFirst
    if (firstFilesBases) {
      const result = await store.appStoreData.initCurrentFilesBases(firstFilesBases.id)
      if (result && !result.status) {
        ElMessage.error(result.message);
        return
      }
    }
    store.searchStoreData.init();
  } catch (err) {
    console.log(err)
  } finally {
    initStatus.value = true
    LoadingService.hide()
  }
}

onMounted(async () => {
  await init()
})
</script>
<style lang="scss" scoped>
.app-container {
  width: calc(100vw - 10px);
  height: calc(100vh - 10px);
  padding: 5px;
  overflow: hidden;
  background-color: #1f1f1f;
  display: flex;
  flex-direction: column;
}
</style>
