<template>
  <div class="setting-data">
    <el-form label-width="auto">
      <el-form-item label="演员集名称">
        <el-input v-model="formData.name" />
      </el-form-item>
    </el-form>
    <div class="setting-data-performer">
      <performerDataList :performerBasesId="props.performerBases.id" :showPerformerInfo="false"></performerDataList>
    </div>
  </div>
</template>
<script setup lang="ts">
import { debounce } from '@/assets/debounce';
import performerDataList from '@/components/performer/performerDataList.vue';
import type { I_performerBases } from '@/dataType/performer.dataType';
import { performerBasesServer } from '@/server/performerBases.server';
import { performerBasesStoreData } from '@/storeData/performerBases.storeData';
import { ElMessage } from 'element-plus';
import { ref, watch, type PropType } from 'vue';
const store = {
  performerBasesStoreData: performerBasesStoreData(),
}
const props = defineProps({
  performerBases: {
    type: Object as PropType<I_performerBases>,
    required: true,
  },
})

const formData = ref<I_performerBases>({ ...props.performerBases });

watch(
  formData,
  (newVal) => {
    update(newVal)
  },
  { deep: true }
);

const update = debounce(async (value: I_performerBases) => {
  const result = await performerBasesServer.update(value)
  if (result.status) {
    store.performerBasesStoreData.update(value)
  } else {
    ElMessage.error(result.msg)
  }
})
</script>
<style lang="scss" scoped>
@use '../settingsModules' as modules;

.setting-data {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;

  .el-form {
    width: 100%;
    padding: 0 20px;
  }

  .setting-data-performer {
    flex: 1;
    overflow: auto;
  }
}

.setting-data {
  padding: 0 0 8px; box-sizing: border-box; min-height: 0;
  > .el-form { @include modules.panel; flex-shrink: 0; padding: 8px 16px; margin-bottom: 8px; }
  > .el-form :deep(.el-form-item) { margin: 0; }
  > .el-form :deep(.el-input) { max-width: 420px; }
  .setting-data-performer { min-height: 0; }
}


.setting-data-performer {
  :deep(.performer-search) { flex-direction: column; gap: 8px; padding-bottom: 8px; }
  :deep(.performer-search-filters), :deep(.performer-search-actions) { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; }
  :deep(.performer-search-actions:empty) { display: none; }
}
</style>
