<template>
  <div class="layout-table-style1">
    <el-table ref="tableRef" :data="props.dataList" border height="100%" style="width: 100%" size="small"
      @row-click="selectResourcesHandle" :row-class-name="selectionRowClass">
      <el-table-column v-if="selection?.active.value" width="44" align="center">
        <template #default="{ row }">
          <el-checkbox :model-value="selection.selectedIds.value.has(row.id)" :aria-label="`选择资源：${row.title}`"
            @click.stop @change="selection.toggle(row)" />
        </template>
      </el-table-column>
      <el-table-column v-if="selection?.active.value" label="关联组" width="100">
        <template #default="{ row }"><ResourceGroupMark :group-id="row.associationGroupId" /></template>
      </el-table-column>
      <el-table-column label="-" width="32" align="center">
        <template #default="scope">
          <div v-if="scope.row.coverPoster != ''" @click.stop="playOrSelectResource(scope.row)">
            <el-popover placement="right-start" trigger="hover" :width="(scope.row.coverPosterWidth + 26) + 'px'">
              <el-image
                :style="{ width: scope.row.coverPosterWidth + 'px', height: scope.row.coverPosterHeight + 'px' }"
                :src="getResourceCoverPoster(scope.row)" />
              <template #reference>
                <el-icon class="video-play" size="16">
                  <VideoPlay />
                </el-icon>
              </template>
            </el-popover>
          </div>
          <div class="playDiv" v-else @click.stop="playOrSelectResource(scope.row)">
            <el-icon class="video-play" size="16">
              <VideoPlay />
            </el-icon>
          </div>
        </template>
      </el-table-column>
      <el-table-column v-if="store.appStoreData.currentConfigApp.showVideoDuration" label="时长" width="120"
        align="center">
        <template #default="scope">
          {{ getResourceDurationText(scope.row) || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="title" label="名称" min-width="180" :show-overflow-tooltip="true"></el-table-column>
      <el-table-column prop="issueNumber" label="版号、番号、刊号" width="120" :show-overflow-tooltip="true" />
      <el-table-column label="评星" width="80">
        <template #default="scope">
          <div v-if="scope.row.stars > 0">
            <el-icon v-for="s, key in scope.row.stars" :key="key" color="#F7BA2A">
              <StarFilled />
            </el-icon>
          </div>
          <div v-else>-</div>
        </template>
      </el-table-column>
      <el-table-column prop="performers" label="演员" width="160" :show-overflow-tooltip="true">
        <template #default="{ row }">
          <div class="tag-container">
            <label v-for="performer in row.performers" :key="performer.id">{{ performer.name }}</label>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="tags" label="标签" min-width="180" :show-overflow-tooltip="true">
        <template #default="{ row }">
          <div class="tag-container">
            <label v-for="tag in row.tags" :key="tag.id">{{ tag.name }}</label>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="信息" width="200" :show-overflow-tooltip="true">
        <template #default="scope">
          <el-breadcrumb class="breadcrumb" separator="|">
            <el-breadcrumb-item v-if="scope.row.issuingDate != ''">
              {{ scope.row.issuingDate }}
            </el-breadcrumb-item>
            <el-breadcrumb-item v-if="scope.row.country != ''">
              {{ $t('country.' + scope.row.country) }}
            </el-breadcrumb-item>
            <el-breadcrumb-item v-if="scope.row.definition != ''">
              {{ $t('definition.' + scope.row.definition) }}
            </el-breadcrumb-item>
          </el-breadcrumb>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>
<script lang="ts" setup>
import type { I_resource } from '@/dataType/resource.dataType';
import { ref, inject, type PropType } from 'vue';
import { getResourceCoverPoster } from '@/common/photo';
import { playResource } from '@/common/play';
import type { ElTable } from 'element-plus';
import { appStoreData } from '@/storeData/app.storeData';
import { getResourceDurationText } from '@/common/videoDuration';

import { resourceSelectionKey } from './resourceSelection';
import ResourceGroupMark from './ResourceGroupMark.vue';
const selection = inject(resourceSelectionKey, undefined);
const selectionRowClass = ({ row }: { row: I_resource }) => selection?.active.value && selection.selectedIds.value.has(row.id) ? 'management-selected-row' : '';
const playOrSelectResource = (resource: I_resource) => {
  if (selection?.active.value) selection.toggle(resource);
  else playResource(resource);
};
const store = {
  appStoreData: appStoreData(),
}

const props = defineProps({
  dataList: {
    type: Array as PropType<I_resource[]>,
    default: () => [],
  },
})
const emits = defineEmits(['selectResources']);
const tableRef = ref<InstanceType<typeof ElTable>>();
const selectResourcesHandle = (item: I_resource) => {
  if (selection?.active.value) {
    selection.toggle(item);
    return;
  }
  emits('selectResources', item)
}

const change = () => {
  tableRef.value?.setScrollTop(0);
};

defineExpose({ change });
</script>
<style lang="scss" scoped>
:deep(.management-selected-row) {
  --el-table-tr-bg-color: color-mix(in srgb, var(--el-color-primary) 12%, var(--el-bg-color));
  --el-table-row-hover-bg-color: color-mix(in srgb, var(--el-color-primary) 18%, var(--el-bg-color));
}
:deep(.management-selected-row > td:first-child) { box-shadow: inset 2px 0 var(--el-color-primary); }
.layout-table-style1 {
  width: 100%;
  height: 100%;
  overflow: hidden;

  .video-play {
    cursor: pointer;
  }

  .tag-container {
    display: flex;
    gap: 5px;
  }

  .breadcrumb {
    font-size: 12px;
  }
}
</style>
