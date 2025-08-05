<template>
  <dialogCommon ref="dialogCommonRef" title="文件数据库排序" :footer="false">
    <draggable class="fileDatabase-list-draggable" :list="store.filesBasesStoreData.filesBases" item-key="id"
      @end="onDragEnd">
      <template #item="{ element, index }">
        <el-tag class="draggable-item" type="primary" size="large">{{ element.name }}</el-tag>
      </template>
    </draggable>
  </dialogCommon>
</template>
<script lang="ts" setup>
import dialogCommon from '@/components/com/dialog/dialog.common.vue';
import { ref } from 'vue';
import { filesBasesStoreData } from '@/storeData/filesBases.storeData';
import draggable from 'vuedraggable';
import type { I_filesBases_sort } from '@/dataType/filesBases.dataType';
import { filesBasesServer } from '@/server/filesBases.server';
import { ElMessage } from 'element-plus';
import { LoadingService } from '@/assets/loading';
import { debounce } from '@/assets/debounce';
const store = {
  filesBasesStoreData: filesBasesStoreData(),
}

const dialogCommonRef = ref<InstanceType<typeof dialogCommon>>();

const onDragEnd = debounce(async () => {
  LoadingService.show();
  try {
    const sortObj: I_filesBases_sort[] = [];
    store.filesBasesStoreData.filesBases.forEach((item, index) => {
      sortObj.push({
        id: item.id,
        sort: index,
      });
    });
    const result = await filesBasesServer.sort(sortObj);
    if (result.status) {
      ElMessage.success('排序成功');
    } else {
      ElMessage.error(result.msg);
    }
  } catch (error) {
    ElMessage.error('提交失败，请稍后再试');
  } finally {
    LoadingService.hide();
  }
}, 3000)

const open = () => {
  dialogCommonRef.value?.open()
}
const close = () => {
  dialogCommonRef.value?.close()
}
// eslint-disable-next-line no-undef
defineExpose({ open, close })
</script>
<style lang="scss" scoped>
.fileDatabase-list-draggable {
  height: 320px;
  padding: 10px 0px;
  overflow-y: auto;
  display: flex;
  flex-wrap: wrap;
  align-content: flex-start;
  gap: 10px;

  .draggable-item {
    cursor: move;
  }
}
</style>
