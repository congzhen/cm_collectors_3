<template>
  <div v-loading="loading" class="performer-tag-select-list">
    <el-form-item v-for="group in groups_C" :key="group.id" :label="group.name">
      <el-select-v2
        class="performer-tag-select"
        :model-value="selectedByClass(group.id)"
        :options="group.options"
        multiple
        filterable
        clearable
        :placeholder="`请选择${group.name}`"
        @update:model-value="updateClassSelection(group.id, $event)"
      />
    </el-form-item>
    <div v-if="!loading && groups_C.length === 0" class="performer-tag-empty">
      暂无可用的演员标签分类
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { performerTagServer } from '@/server/performerTag.server';
import type { I_performerTag, I_performerTagClass } from '@/dataType/performerTag.dataType';
import { ElMessage } from 'element-plus';

interface SelectOption {
  value: string;
  label: string;
}

const props = defineProps<{ modelValue: string[]; performerBasesId: string }>();
const emits = defineEmits<{ 'update:modelValue': [value: string[]] }>();
const classes = ref<I_performerTagClass[]>([]);
const tags = ref<I_performerTag[]>([]);
const loading = ref(false);

const selectedIdSet_C = computed(() => new Set(props.modelValue));

const groups_C = computed(() => classes.value
  .filter(tagClass => tagClass.status || tags.value.some(tag =>
    tag.performerTagClass_id === tagClass.id && selectedIdSet_C.value.has(tag.id)))
  .map(tagClass => ({
    id: tagClass.id,
    name: tagClass.status ? tagClass.name : `${tagClass.name}（已停用）`,
    options: tags.value
      .filter(tag => tag.performerTagClass_id === tagClass.id &&
        (tagClass.status ? tag.status || selectedIdSet_C.value.has(tag.id) : selectedIdSet_C.value.has(tag.id)))
      .map<SelectOption>(tag => ({
        value: tag.id,
        label: tag.status ? tag.name : `${tag.name}（已停用）`,
      })),
  })));

const selectedByClass = (classId: string) => {
  const classTagIds = new Set(tags.value
    .filter(tag => tag.performerTagClass_id === classId)
    .map(tag => tag.id));
  return props.modelValue.filter(id => classTagIds.has(id));
};

const updateClassSelection = (classId: string, selectedIds: string[]) => {
  const classTagIds = new Set(tags.value
    .filter(tag => tag.performerTagClass_id === classId)
    .map(tag => tag.id));
  const selectionsFromOtherClasses = props.modelValue.filter(id => !classTagIds.has(id));
  emits('update:modelValue', [...selectionsFromOtherClasses, ...selectedIds]);
};

const load = async () => {
  if (!props.performerBasesId) {
    classes.value = [];
    tags.value = [];
    return;
  }

  loading.value = true;
  try {
    const result = await performerTagServer.data(props.performerBasesId, true);
    if (result?.status) {
      classes.value = result.data.tagClasses;
      tags.value = result.data.tags;
    } else {
      ElMessage.error(result?.msg || '加载演员标签失败');
    }
  } finally {
    loading.value = false;
  }
};

watch(() => props.performerBasesId, load, { immediate: true });
</script>

<style lang="scss" scoped>
.performer-tag-select-list {
  min-height: 32px;
}

.performer-tag-select {
  width: 100%;
}

.performer-tag-empty {
  padding: 6px 0 14px 80px;
  color: var(--el-text-color-placeholder);
  font-size: var(--el-font-size-small);
}
</style>
