<template>
  <div class="app-container" v-if="initStatus">
    <router-view v-slot="{ Component }">
      <keep-alive>
        <component :is="Component" />
      </keep-alive>
    </router-view>
  </div>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { appDataServer } from './server/app.server';
import { appStoreData } from './storeData/app.storeData';
import { LoadingService } from './assets/loading';
const initStatus = ref(false);
const init = async () => {
  LoadingService.show();
  const addDataResult = await appDataServer.init();
  if (addDataResult && addDataResult.status) {
    appStoreData().init(addDataResult.data);
    LoadingService.hide();
  } else {
    alert(addDataResult.msg)
    return
  }
  initStatus.value = true;
}


onMounted(async () => {
  await init();
});

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
