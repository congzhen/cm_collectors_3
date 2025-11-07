<template>
  <el-select v-model="selectVal" clearable :style="{ width: props.width }" placeholder="文件管理挂载路径" @change="changeHandle"
    @clear="handleClear" :multiple="props.multiple" filterable allow-create default-first-option>
    <el-option v-for="item, index in dataset.serverFileManagementRootPath" :key="index" :label="formatDriveLabel(item)"
      :value="item">
      <div class="drive-option">
        <span class="drive-icon">📁</span>
        <span class="drive-label">{{ formatDriveLabel(item) }}</span>
        <span class="drive-value">{{ item }}</span>
      </div>
    </el-option>
  </el-select>
</template>
<script setup lang="ts">
import dataset from '@/assets/dataset';
const selectVal = defineModel<string | string[]>({ type: [String, Array], default: "" as string | string[] });
const props = defineProps({
  width: {
    type: String,
    default: '100%',
  },
  multiple: {
    type: Boolean,
    default: false
  }
})

// 格式化驱动器标签显示
const formatDriveLabel = (drive: string) => {
  if (drive === '/') return 'Linux 根目录';
  return drive;
};

const emit = defineEmits(['change'])

const changeHandle = () => {
  emit('change', selectVal.value || '')
}
const handleClear = () => {
  if (props.multiple) {
    selectVal.value = [];
  } else {
    selectVal.value = '';
  }
}
</script>
<style lang="scss" scoped>
.drive-option {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 0;
}

.drive-icon {
  font-size: 16px;
}

.drive-label {
  font-weight: 500;
  flex: 1;
}

.drive-value {
  color: #999;
  font-size: 12px;
}
</style>
