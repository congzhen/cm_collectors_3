<template>
  <div class="app-container" v-if="initStatus">
    <router-view v-slot="{ Component }">
      <keep-alive exclude="playMovies,playMoviesMobile,playComic,playComicMobile,playAtlas,playAtlasMobile">
        <component :is="Component" :key="router.currentRoute.value.fullPath" />
      </keep-alive>
    </router-view>
  </div>
  <div v-else-if="startupError" class="mobile-startup-error">
    <p>{{ startupError }}</p>
    <el-button type="primary" @click="init">重新连接</el-button>
  </div>
  <UpdateCenter v-if="initStatus" />
</template>
<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, watch } from 'vue'
import { appStoreData } from '@/storeData/app.storeData'
import { filesBasesStoreData } from '@/storeData/filesBases.storeData'
import { performerBasesStoreData } from '@/storeData/performerBases.storeData';
import { searchStoreData } from './storeData/search.storeData';
import { LoadingService } from '@/assets/loading'
import { ElMessage } from 'element-plus'
import { runRuntimeBridge } from "@/common/runtimeBridge"
import router from '@/router'
import { isMobile, setCloseMobileDisplay } from '@/assets/mobile'
import { isRestorableMobilePath, mobileRestorePath, readMobileSession, saveMobileSession } from '@/common/mobileSession'
import { applyTheme } from '@/common/theme'
import UpdateCenter from '@/components/setting/updateSoft/UpdateCenter.vue'

const initStatus = ref(false)
const startupError = ref('')
let stopSaving: (() => void) | undefined
let mobileSessionEnabled = false

const store = {
  appStoreData: appStoreData(),
  filesBasesStoreData: filesBasesStoreData(),
  performerBasesStoreData: performerBasesStoreData(),
  searchStoreData: searchStoreData(),
}

// 切换主题函数
const init = async () => {
  const saved = readMobileSession()
  let ready = false
  try {
    startupError.value = ''
    LoadingService.show()
    const result = await store.appStoreData.initApp();
    if (result && !result.status) {
      throw new Error(result.message || '连接失败')
    }
    // 根据服务端配置设置是否关闭移动端显示
    setCloseMobileDisplay(store.appStoreData.appConfig.closeMobileDisplay)
    // 如果首次路由守卫已经进入 /mobile，等待服务端配置加载后再纠正回 PC 端。
    if (store.appStoreData.appConfig.closeMobileDisplay && router.currentRoute.value.path === '/mobile') {
      await router.replace('/')
    }
    // 根据存储的主题设置初始化主题
    applyTheme(store.appStoreData.appConfig.theme)
    mobileSessionEnabled = isMobile()
    const firstFilesBases = (mobileSessionEnabled && saved.filesBasesId
      ? store.filesBasesStoreData.filesBasesStatus.find(item => item.id === saved.filesBasesId) : undefined)
      || store.filesBasesStoreData.filesBasesFirst
    if (firstFilesBases) {
      const result = await store.appStoreData.initCurrentFilesBases(firstFilesBases.id)
      if (result && !result.status) {
        throw new Error(result.message || '资源库加载失败')
      }
    }
    store.searchStoreData.init();
    if (mobileSessionEnabled) {
      if (saved.filesBasesId === firstFilesBases?.id && saved.search) store.searchStoreData.applySearchData(saved.search)
      await router.isReady()
      const target = mobileRestorePath(router.currentRoute.value.fullPath, saved.path)
      if (target !== router.currentRoute.value.fullPath) await router.replace(target)
      stopSaving?.()
      stopSaving = watch(() => [router.currentRoute.value.fullPath, store.appStoreData.currentFilesBases.id,
        store.searchStoreData.searchData], savePhoneSession, { deep: true, flush: 'post' })
    }
    ready = true
  } catch (err) {
    console.log(err)
    if (isMobile()) startupError.value = '连接暂不可用，恢复位置已保留，请重新连接。'
    else ElMessage.error('应用初始化失败')
  } finally {
    initStatus.value = ready || !isMobile()
    if (ready) savePhoneSession()
    LoadingService.hide()
  }
}
const savePhoneSession = () => {
  if (!mobileSessionEnabled || !initStatus.value) return
  const path = router.currentRoute.value.fullPath
  if (!isRestorableMobilePath(path)) return
  saveMobileSession({ path, filesBasesId: store.appStoreData.currentFilesBases.id,
    search: store.searchStoreData.searchData })
}

onMounted(async () => {
  document.addEventListener('visibilitychange', savePhoneSession)
  window.addEventListener('pagehide', savePhoneSession)
  await init()
  const runtimeBridgeStatus = await runRuntimeBridge();
  if (runtimeBridgeStatus) {
    store.appStoreData.runtimeBridgeStatus = runtimeBridgeStatus;
  }
})
onBeforeUnmount(() => {
  stopSaving?.()
  document.removeEventListener('visibilitychange', savePhoneSession)
  window.removeEventListener('pagehide', savePhoneSession)
})
</script>
<style lang="scss" scoped>
.mobile-startup-error { padding: 24px; text-align: center; }
.app-container {
  width: calc(100% - 10px);
  height: calc(100% - 10px);
  padding: 5px;
  overflow: hidden;
  background-color: #1f1f1f;
  display: flex;
  flex-direction: column;
}
</style>
