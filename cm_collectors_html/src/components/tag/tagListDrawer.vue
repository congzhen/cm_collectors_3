<template>
  <drawerCommon class="tagListDrawer" ref="drawerCommonRef" width="800px" title="标签" :btnSubmit="false">
    <tagList ref="tagListRef" :id="store.appStoreData.currentFilesBases.id"
      @update-tag-data-completed="updateTagDataCompletedHandle">
    </tagList>
    <template #footerBtn>
      <div>
        <el-button @click="tagClassFormHandle">添加标签分类</el-button>
      </div>
    </template>
  </drawerCommon>
  <tagFormClassDialog ref="tagFormClassDialogRef" @success="addTagClassSuccess"></tagFormClassDialog>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import drawerCommon from '@/components/com/dialog/drawer-common.vue';
import tagList from './tagList.vue';
import tagFormClassDialog from './tagFormClassDialog.vue'
import { appStoreData } from '@/storeData/app.storeData';
import type { I_tagClass } from '@/dataType/tag.dataType';
const store = {
  appStoreData: appStoreData(),
}

const drawerCommonRef = ref<InstanceType<typeof drawerCommon>>();
const tagListRef = ref<InstanceType<typeof tagList>>();
const tagFormClassDialogRef = ref<InstanceType<typeof tagFormClassDialog>>();
const tagClassFormHandle = () => {
  const tagClass: I_tagClass = {
    id: '',
    name: '',
    filesBases_id: store.appStoreData.currentFilesBases.id,
    leftShow: true,
    sort: 0,
    status: true,
  }
  tagFormClassDialogRef.value?.open('add', tagClass)
}

const updateTagDataCompletedHandle = () => {
  store.appStoreData.initTagData(store.appStoreData.currentFilesBases.id)
}

const addTagClassSuccess = () => {
  updateTagDataCompletedHandle();
  tagListRef.value?.init(() => {
    const scrollContainer = document.querySelector('.tagListDrawer .el-drawer__body');
    if (scrollContainer && scrollContainer.scrollHeight > scrollContainer.clientHeight) {
      scrollContainer.scrollTop = 9999999;
    }
  });

}

const open = () => {
  drawerCommonRef.value?.open()
  tagListRef.value?.init();
}
const close = () => {
  drawerCommonRef.value?.close()
}
// eslint-disable-next-line no-undef
defineExpose({ open, close })
</script>
