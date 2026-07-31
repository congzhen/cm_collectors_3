<template>
  <div class="tool" :class="{ 'tool--labels': props.showLabels }" v-if="props.resource">
    <el-button-group>
      <el-button icon="VideoPlay" @click="playResourceHandle" :style="{ width: buttonWidth }">
        <span v-if="props.showLabels">播放</span>
      </el-button>
      <el-button icon="Folder" v-admin @click="playOpenResourceFolder(props.resource.id)"
        :style="{ width: buttonWidth }">
        <span v-if="props.showLabels">打开文件夹</span>
      </el-button>
      <el-button icon="Edit" v-admin @click="editResourceHandle" :style="{ width: buttonWidth }">
        <span v-if="props.showLabels">编辑</span>
      </el-button>
      <el-button icon="Delete" v-admin @click="resourceDeleteHandle" :style="{ width: buttonWidth }">
        <span v-if="props.showLabels">删除</span>
      </el-button>
    </el-button-group>
  </div>
</template>
<script lang="ts" setup>
import type { I_resource } from '@/dataType/resource.dataType'
import { type PropType, computed } from 'vue'
import { resourceDelete } from '@/common/resource';
import { playResource, playOpenResourceFolder } from '@/common/play'
import { appStoreData } from '@/storeData/app.storeData'
import { eventBus } from '@/main';

const props = defineProps({
  resource: {
    type: Object as PropType<I_resource> | undefined,
    default: undefined
  },
  width: {
    type: String,
    default: '100%'
  },
  showLabels: {
    type: Boolean,
    default: false,
  },
})
const emits = defineEmits(['paly', 'updateResouceSuccess', 'deleteResourceSuccess'])

const store = appStoreData()

// 计算可见按钮数量
const visibleButtonCount = computed(() => {
  let count = 1; // 默认的播放按钮总是可见

  // 检查是否是管理员且已登录
  if (store.isAdminLoginStatus) {
    count += 3; // 文件夹、编辑、删除按钮都可见
  }

  return count;
});

// 计算按钮宽度
const buttonWidth = computed(() => {
  if (props.showLabels) {
    return 'auto';
  }
  // 如果只有一个按钮可见，它占据50%宽度
  if (visibleButtonCount.value === 1) {
    return '70%';
  }
  // 如果有多个按钮可见，平均分配宽度
  return `${100 / visibleButtonCount.value}%`;
});

const playResourceHandle = () => {
  if (!props.resource) return
  playResource(props.resource)
  emits('paly')
}
const editResourceHandle = () => {
  eventBus.emit('edit-resource', { resource: props.resource });
}
const resourceDeleteHandle = () => {
  if (!props.resource) return
  resourceDelete(props.resource, () => {
    emits('deleteResourceSuccess')
  })
}
</script>
<style scoped lang="scss">
.tool {
  width: 100%;
  flex-shrink: 0;
  padding-bottom: 5px;

  .el-button-group {
    width: 100%;
    display: flex;
    justify-content: center;
  }

  &.tool--labels {
    padding-bottom: 0;

    .el-button-group {
      justify-content: flex-start;
      gap: 8px;
    }

    :deep(.el-button) {
      margin-left: 0;
      border-radius: 6px;
    }
  }
}
</style>
