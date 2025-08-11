<template>
  <div class="database-menu">
    <el-menu :default-active="activeDataBase" mode="horizontal" text-color="#fff" active-text-color="#ffd04b">
      <el-menu-item :index="filesBases.id" v-for="filesBases, key in store.filesBasesStoreData.filesBasesStatus"
        :key="key" @click="changeDataBaseHandle(filesBases)">
        {{ filesBases.name }}
      </el-menu-item>
    </el-menu>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue'
import { filesBasesStoreData } from '@/storeData/filesBases.storeData';
import type { I_filesBases } from '@/dataType/filesBases.dataType';
const store = {
  filesBasesStoreData: filesBasesStoreData(),
}
const emits = defineEmits(['selectFilesBase']);

const activeDataBase = ref(store.filesBasesStoreData.filesBasesFirst?.id || '')

const changeDataBaseHandle = (filesBases: I_filesBases) => {
  if (filesBases.id != activeDataBase.value) {
    activeDataBase.value = filesBases.id;
    emits('selectFilesBase', filesBases);
  }
}

</script>
<style lang="scss" scoped>
.database-menu {
  height: 31px;
  padding-bottom: 5px;

  :deep(.el-menu--horizontal) {
    height: auto;
  }

  :deep(.el-menu--horizontal>.el-menu-item) {
    height: 14px;
    line-height: 16px;
    padding-top: 12px;
    padding-bottom: 16px;

    &:hover {
      background-color: unset;
      color: #409EFF;
      user-select: none;
    }
  }

}
</style>
