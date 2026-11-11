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

.database-setting {
  box-sizing: border-box;
  padding: 0;
  > .database-setting-btn { width: auto; flex-shrink: 0; margin: 0 0 8px; padding: 8px 16px; background: var(--el-bg-color); border: 1px solid var(--el-border-color-lighter); border-radius: 8px; }
  > .setting-tabs { flex: 1; min-height: 0; height: auto; }
  :deep(.tag-list) { box-sizing: border-box; height: 100%; padding: 0 0 16px; }
  :deep(.tag-list-header) { flex-wrap: wrap; gap: 10px; }
}

</style>
