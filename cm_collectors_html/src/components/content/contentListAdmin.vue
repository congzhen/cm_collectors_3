<template>
  <div class="content-list-admin">
    <div class="btn-container">
      <el-button-group>
        <el-button @click="batchDeleteHandle">批量删除</el-button>
        <el-button @click="batchAddTagHandle">批量添加标签</el-button>
        <el-button @click="batchDeleteTagHandle">批量删除标签</el-button>
      </el-button-group>
    </div>
    <div class="table-container">
      <el-table ref="tableRef" :data="props.dataList" border height="100%" style="width: 100%">
        <el-table-column type="selection" width="55" />
        <el-table-column width="120" label="操作">
          <template #default="{ row }">
            <el-button-group>
              <el-button icon="VideoPlay" @click="playResourceHandle(row)" />
              <el-button icon="Folder" @click="playOpenResourceFolder(row.id)" />
              <el-button icon="Edit" @click="editResourceHandle(row)" />
              <el-button icon="Delete" @click="resourceDeleteHandle(row)" />
            </el-button-group>
          </template>
        </el-table-column>
        <el-table-column width="120" label="封面">
          <template #default="{ row }">
            <el-image :src="getResourceCoverPoster(row)" fit="cover" />
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="180" />
        <el-table-column prop="issueNumber" label="版号、番号、刊号" width="140" />
        <el-table-column prop="issuingDate" label="年份" width="100" />
        <el-table-column prop="country" label="国家" width="120">
          <template #default="{ row }">
            {{ appLang.country(row.country) }}
          </template>
        </el-table-column>
        <el-table-column prop="performers" label="演员" width="260">
          <template #default="{ row }">
            <div class="tag-container">
              <el-tag v-for="performer in row.performers" :key="performer.id" size="small">{{ performer.name }}</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="tags" label="标签" width="300">
          <template #default="{ row }">
            <div class="tag-container">
              <el-tag v-for="tag in row.tags" :key="tag.id" size="small">{{ tag.name }}</el-tag>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <resourceFormDrawer ref="resourceFormDrawerRef" @success="updateResourceSuccessHandle"></resourceFormDrawer>
      <resourceSetTagBatchDialog ref="resourceSetTagBatchDialogRef" @success="updateResourceSuccessHandle">
      </resourceSetTagBatchDialog>
    </div>
  </div>
</template>
<script lang="ts" setup>
import type { I_resource } from '@/dataType/resource.dataType';
import resourceFormDrawer from '@/components/resource/resourceFormDrawer.vue'
import resourceSetTagBatchDialog from '../resource/resourceSetTagBatchDialog.vue';
import { getResourceCoverPoster } from '@/common/photo';
import { playResource, playOpenResourceFolder } from '@/common/play'
import { resourceDelete, resourceBatchDelete } from '@/common/resource'
import { ref, type PropType } from 'vue';
import { ElMessage, type ElTable } from 'element-plus';
import { AppLang } from '@/language/app.lang'
const appLang = AppLang()

const props = defineProps({
  dataList: {
    type: Array as PropType<I_resource[]>,
    default: () => [],
  },
})
const emits = defineEmits(['selectResources', 'updateData']);

const tableRef = ref<InstanceType<typeof ElTable>>()
const resourceFormDrawerRef = ref<InstanceType<typeof resourceFormDrawer>>()
const resourceSetTagBatchDialogRef = ref<InstanceType<typeof resourceSetTagBatchDialog>>()
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const selectResourcesHandle = (item: I_resource) => {
  emits('selectResources', item)
}

const playResourceHandle = (item: I_resource) => {
  playResource(item)
}
const editResourceHandle = (item: I_resource) => {
  resourceFormDrawerRef.value?.open('edit', item)
}
const updateResourceSuccessHandle = (data: I_resource) => {
  emits('updateData', data)
}
const resourceDeleteHandle = (item: I_resource) => {
  resourceDelete(item, () => {
    emits('updateData')
  })
}

const batchDeleteHandle = () => {
  if (!tableRef.value) return;
  const selectedResources = tableRef.value.getSelectionRows();
  resourceBatchDelete(selectedResources, () => {
    emits('updateData')
  })
}
const batchAddTagHandle = () => {
  if (!tableRef.value) return;
  const selectedResources = tableRef.value.getSelectionRows() as I_resource[];
  if (selectedResources.length == 0) {
    ElMessage.error('请选择要添加标签的资源');
  } else {
    resourceSetTagBatchDialogRef.value?.open(selectedResources, 'add')
  }
}
const batchDeleteTagHandle = () => {
  if (!tableRef.value) return;
  const selectedResources = tableRef.value.getSelectionRows() as I_resource[];
  if (selectedResources.length == 0) {
    ElMessage.error('请选择要删除标签的资源');
  } else {
    resourceSetTagBatchDialogRef.value?.open(selectedResources, 'remove')
  }
}

</script>
<style lang="scss" scoped>
.content-list-admin {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;

  .btn-container {
    flex-shrink: 0;
    display: flex;
    gap: 5px;
    padding-bottom: 5px;
  }

  .table-container {
    flex: 1;
    overflow: hidden;

    .tag-container {
      display: flex;
      flex-wrap: wrap;
      gap: 4px;
    }
  }
}
</style>
