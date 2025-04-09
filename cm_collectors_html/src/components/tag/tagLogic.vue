<template>
  <el-dropdown placement="bottom-end" @click.stop>
    <div class="tag-logic" @click.stop>
      <el-icon>
        <HelpFilled />
      </el-icon>
      <label class="tag-logic-text">
        {{ props.text }}
      </label>
    </div>
    <template #dropdown>
      <el-dropdown-item> 单选 </el-dropdown-item>
      <el-dropdown-item> 多选(或) </el-dropdown-item>
      <el-dropdown-item> 多选(与) </el-dropdown-item>
      <el-dropdown-item> 非 </el-dropdown-item>
    </template>
  </el-dropdown>
</template>
<script setup lang="ts">
const props = defineProps({
  text: {
    type: String,
    default: '',
  },
})
</script>
<style lang="scss" scoped>
.tag-logic {
  cursor: pointer;
  display: flex;
  .tag-logic-text {
    padding-left: 0.2em;
    cursor: pointer;
  }
}
.tag-logic:hover {
  color: var(--el-color-primary);
}
.tag-logic:focus-visible {
  outline: none; /* 移除默认焦点环 */
}
</style>
