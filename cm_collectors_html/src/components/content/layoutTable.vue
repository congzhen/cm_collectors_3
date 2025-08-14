<template>
  <div class="layout-table-style1">
    <el-table ref="tableRef" :data="props.dataList" border height="100%" style="width: 100%" size="small"
      @row-click="selectResourcesHandle">
      <el-table-column label="-" width="32" align="center">
        <template #default="scope">
          <div v-if="scope.row.coverPoster != ''" @click.stop="playResource(scope.row)">
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
          <div class="playDiv" v-else @click.stop="playResource(scope.row)">
            <el-icon class="video-play" size="16">
              <VideoPlay />
            </el-icon>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="title" label="名称" min-width="180" :show-overflow-tooltip="true"></el-table-column>
      <el-table-column prop="issueNumber" label="版号、番号、刊号" width="120" :show-overflow-tooltip="true" />
      <el-table-column label="评分" width="80">
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
import { ref, type PropType } from 'vue';
import { getResourceCoverPoster } from '@/common/photo';
import { playResource } from '@/common/play';
import type { ElTable } from 'element-plus';
const props = defineProps({
  dataList: {
    type: Array as PropType<I_resource[]>,
    default: () => [],
  },
})
const emits = defineEmits(['selectResources']);
const tableRef = ref<InstanceType<typeof ElTable>>();
const selectResourcesHandle = (item: I_resource) => {
  emits('selectResources', item)
}

const change = () => {
  tableRef.value?.setScrollTop(0);
};

defineExpose({ change });
</script>
<style lang="scss" scoped>
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
