<template>
  <dialogCommon ref="dialogCommonRef" title="标签排序" :footer="false">
    <draggable class="fileDatabase-list-draggable" :list="tags" item-key="id" @end="onDragEnd">
      <template #item="{ element }">
        <el-tag class="draggable-item" type="primary" size="large">{{ element.name }}</el-tag>
      </template>
    </draggable>
  </dialogCommon>
</template>
<script lang="ts" setup>
import dialogCommon from '@/components/com/dialog/dialog-common.vue';
import { ref } from 'vue';
import draggable from 'vuedraggable';
import { debounce } from '@/assets/debounce';
import type { I_tag } from '@/dataType/tag.dataType';

const emits = defineEmits(['success'])

const dialogCommonRef = ref<InstanceType<typeof dialogCommon>>();
const tags = ref<I_tag[]>([]);
const init = (_tags: I_tag[]) => {
  tags.value = _tags;
}

const onDragEnd = debounce(async () => {
  const ids = tags.value.map(tag => tag.id);
  emits('success', ids)
}, 500)

const open = (_tags: I_tag[]) => {
  init(_tags);
  dialogCommonRef.value?.open()
}
const close = () => {
  dialogCommonRef.value?.close()
}
// eslint-disable-next-line no-undef
defineExpose({ open, close })
</script>
<style lang="scss" scoped>
.fileDatabase-list-draggable {
  height: 320px;
  padding: 10px 0px;
  overflow-y: auto;
  display: flex;
  flex-wrap: wrap;
  align-content: flex-start;
  gap: 10px;

  .draggable-item {
    cursor: move;
  }
}
</style>
