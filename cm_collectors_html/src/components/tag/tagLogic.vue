<template>
  <div class="tag-logic-container" @click.stop>
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
          <el-icon class="single">
            <HelpFilled />
          </el-icon>
          单选
        </el-dropdown-item>
        <el-dropdown-item v-if="logic.length == 0 || props.logic.includes(E_searchLogic.MultiOr)"
          @click.stop="logicClickHandle(E_searchLogic.MultiOr)">
          <el-icon class="multiOr">
            <HelpFilled />
          </el-icon>
          多选(或)
        </el-dropdown-item>
        <el-dropdown-item v-if="logic.length == 0 || props.logic.includes(E_searchLogic.MultiAnd)"
          @click.stop="logicClickHandle(E_searchLogic.MultiAnd)">
          <el-icon class="multiAnd">
            <HelpFilled />
          </el-icon>
          多选(与)
        </el-dropdown-item>
        <el-dropdown-item v-if="logic.length == 0 || props.logic.includes(E_searchLogic.Not)"
          @click.stop="logicClickHandle(E_searchLogic.Not)">
          <el-icon class="not">
            <HelpFilled />
          </el-icon>
          非
        </el-dropdown-item>
      </template>
    </el-dropdown>
  </div>
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
.tag-logic-container {
  width: 80%;
  display: flex;
  height: 3em;
  cursor: auto;
}

.tag-logic {
  width: 100%;
  cursor: pointer;
  display: flex;
  line-height: 3em;

  .el-icon {
    height: 3em;
    line-height: 3em;
  }

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
  color: #409EFF;
}

.not {
  color: #F56C6C;
}

.tag-logic:hover {
  color: var(--el-color-primary);
}

.tag-logic:focus-visible {
  outline: none;
  /* 移除默认焦点环 */
}
</style>
