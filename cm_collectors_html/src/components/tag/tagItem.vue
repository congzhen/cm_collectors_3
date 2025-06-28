<template>
  <div :class="['tag-item', props.status ? '' : 'disable']">
    <label>{{ props.name }}</label>
    <div class="tag-item-tool">
      <div class="tag-item-tool-block">
        <el-icon :size="12" color="#bdbcbc" @click="emits('edit')">
          <Edit />
        </el-icon>
        <el-icon v-if="props.status" :size="12" color="#bdbcbc" @click="emits('delete')">
          <Delete />
        </el-icon>
        <el-icon v-else :size="12" color="#bdbcbc" @click="emits('restore')">
          <RefreshLeft />
        </el-icon>
      </div>
    </div>
  </div>
</template>
<script lang="ts" setup>
const props = defineProps({
  name: {
    type: String,
    required: true,
  },
  status: {
    type: Boolean,
    default: true,
  }
})
const emits = defineEmits(['edit', 'delete', 'restore'])
</script>
<style lang="scss" scoped>
.disable {
  background-color: rgba(255, 255, 255, 0.3);

  label {
    text-decoration: line-through;
  }
}

.tag-item {
  font-size: 12px;
  height: 30px;
  line-height: 30px;
  padding: 0 20px;
  border-radius: 5px;
  border: 1px solid #4c4d4f;
  cursor: pointer;
  -moz-user-select: none;
  -webkit-user-select: none;
  -ms-user-select: none;
  user-select: none;
  position: relative;

  &:hover {
    border: 1px solid #616264;

    label {
      color: #babdc5;
    }

    .tag-item-tool {
      display: block;
    }
  }

  label {
    cursor: pointer;
  }

  .tag-item-tool {
    display: none;
    position: absolute;
    margin-top: -28px;
    right: 2px;

    .tag-item-tool-block {
      display: flex;
      flex-direction: column;
      gap: 2px;

      .el-icon {
        cursor: pointer;

        &:hover {
          color: #fff;
        }
      }
    }
  }
}
</style>
