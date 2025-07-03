<template>
  <div class="database-setting">
    <el-tabs tab-position="left" class="setting-tabs" v-model="activeName">
      <el-tab-pane v-for="item, key in store.filesBasesStoreData.filesBases" :key="key" :name="item.id"
        :label="item.name">
        <tagList :id="item.id"></tagList>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import tagList from '@/components/tag/tagList.vue';
import { filesBasesStoreData } from '@/storeData/filesBases.storeData';
const store = {
  filesBasesStoreData: filesBasesStoreData(),
}
const activeName = ref(store.filesBasesStoreData.filesBasesFirst?.id || '');
</script>
<style lang="scss" scoped>
.database-setting {
  :deep(.tag-list) {
    padding: 0 5px;
    overflow-y: auto;
  }
}
</style>
