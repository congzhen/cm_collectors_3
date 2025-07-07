<template>
  <dialogCommon ref="dialogCommonRef" :width="props.width" :title="props.title" :footer="false" @closed="close">
    <el-table class="tableContainer" ref="tableRef" :data="dataList" border :height="props.tableHeight"
      :size="props.size" style="width: 100%" v-loading="loading">
      <el-table-column v-if="props.check" label="-" width="40" align="center">
        <template #default="scope">
          <el-checkbox class="tr-checkbox" v-model="scope.row.check" />
        </template>
      </el-table-column>
      <el-table-column prop="no" label="-" width="60" align="center">
        <template #default="scope">
          {{ getNo(scope.$index) }}
        </template>
      </el-table-column>
      <slot></slot>
    </el-table>
    <el-pagination class="paginationContainer" v-if="props.paging" v-model:current-page="page" :page-size="props.limit"
      background layout="total, prev, pager, next" :total="dataTotal" :size="props.size"
      @change="changePaginationHandle" />
  </dialogCommon>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import dialogCommon from './dialog.common.vue'

const dialogCommonRef = ref<InstanceType<typeof dialogCommon>>()

const props = defineProps({
  loading: {
    type: Boolean,
    default: false,
  },
  width: {
    type: String,
    default: '800px',
  },
  title: {
    type: String,
    default: '',
  },
  dataList: {
    type: Array,
    default: () => []
  },
  dataTotal: {
    type: Number,
    default: 0
  },
  tableHeight: {
    type: String,
    default: '50vh'
  },
  check: {
    type: Boolean,
    default: false
  },
  size: {
    type: String,
    default: 'small'
  },
  paging: {
    type: Boolean,
    default: false
  },
  currentPage: {
    type: Number,
    default: 1
  },
  limit: {
    type: Number,
    default: 30
  },
})

const emits = defineEmits(['changePagination'])

const page = ref(props.currentPage);
const getNo = (index: number) => {
  return index + 1 + (page.value - 1) * props.limit;
}

const changePaginationHandle = (currentPage: number, pageSize: number) => {
  emits('changePagination', currentPage, pageSize)
}


const reload = () => {
  page.value = 1
}

const open = () => {
  dialogCommonRef.value?.open()
}
const close = () => {
  dialogCommonRef.value?.close()
}
// eslint-disable-next-line no-undef
defineExpose({ open, close, reload })
</script>
