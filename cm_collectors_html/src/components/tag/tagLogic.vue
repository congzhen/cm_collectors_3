<template>
  <el-dropdown placement="bottom-end" @click.stop>
    <div class="tag-logic" @click.stop>
      <el-icon :class="[props.selectedLogic]">
        <HelpFilled />
      </el-icon>
      <label class="tag-logic-text">
        {{ props.text }}
      </label>
    </div>
    <template #dropdown>
      <el-dropdown-item v-if="logic.length == 0 || props.logic.includes(E_searchLogic.Single)"
        @click.stop="logicClickHandle(E_searchLogic.Single)">
        单选
      </el-dropdown-item>
      <el-dropdown-item v-if="logic.length == 0 || props.logic.includes(E_searchLogic.MultiOr)"
        @click.stop="logicClickHandle(E_searchLogic.MultiOr)">
        多选(或)
      </el-dropdown-item>
      <el-dropdown-item v-if="logic.length == 0 || props.logic.includes(E_searchLogic.MultiAnd)"
        @click.stop="logicClickHandle(E_searchLogic.MultiAnd)">
        多选(与)
      </el-dropdown-item>
      <el-dropdown-item v-if="logic.length == 0 || props.logic.includes(E_searchLogic.Not)"
        @click.stop="logicClickHandle(E_searchLogic.Not)">
        非
      </el-dropdown-item>
    </template>
  </el-dropdown>
</template>
<script setup lang="ts">
import { E_searchLogic } from '@/dataType/search.dataType';
import type { PropType } from 'vue';

const props = defineProps({
  text: {
    type: String,
    default: '',
  },
  logic: {
    type: Array as PropType<E_searchLogic[]>,
    default: [],
  },
  selectedLogic: {
    type: String as PropType<E_searchLogic>,
    default: E_searchLogic.Single,
  }
})
const emits = defineEmits(['logicClick']);

const logicClickHandle = (option: E_searchLogic) => {
  emits('logicClick', option);
}

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

.single {
  color: #67C23A;
}

.multiAnd {
  color: #E6A23C;
}

.multiOr {
  color: #F56C6C;
}

.not {
  color: #909399;
}

.tag-logic:hover {
  color: var(--el-color-primary);
}

.tag-logic:focus-visible {
  outline: none;
  /* 移除默认焦点环 */
}
</style>
