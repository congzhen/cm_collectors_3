<template>
  <el-popover v-model:visible="visible" :width="660" trigger="click" placement="bottom-start">
    <template #reference>
      <el-button icon="PriceTag" :type="props.modelValue.length ? 'primary' : 'default'">
        标签筛选<span v-if="props.modelValue.length">（{{ props.modelValue.length }}）</span>
      </el-button>
    </template>
    <div class="performer-tag-filter">
      <div class="filter-header">
        <strong>演员标签筛选</strong>
        <el-button link type="primary" @click="openManagement">管理标签</el-button>
      </div>
      <el-scrollbar max-height="360px">
        <div v-if="groups_C.length" class="filter-groups">
          <div v-for="group in groups_C" :key="group.id" class="filter-group">
            <span class="group-name">{{ group.name }}</span>
            <el-checkbox-group v-model="draftTagIds" class="group-tags">
              <el-checkbox-button v-for="tag in group.tags" :key="tag.id" :value="tag.id">{{ tag.name }}</el-checkbox-button>
            </el-checkbox-group>
          </div>
        </div>
        <el-empty v-else description="还没有演员标签" :image-size="70" />
      </el-scrollbar>
      <div class="filter-footer">
        <el-radio-group v-model="draftMatchMode" size="small">
          <el-radio-button value="any">满足任一</el-radio-button>
          <el-radio-button value="all">同时满足</el-radio-button>
        </el-radio-group>
        <div>
          <el-button @click="clear">清空</el-button>
          <el-button type="primary" @click="apply">应用筛选</el-button>
        </div>
      </div>
    </div>
  </el-popover>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { performerTagServer } from '@/server/performerTag.server';
import type { I_performerTag, I_performerTagClass, PerformerTagMatchMode } from '@/dataType/performerTag.dataType';

const props = defineProps<{ performerBasesId: string; modelValue: string[]; matchMode: PerformerTagMatchMode }>();
const emits = defineEmits<{ change: [value: { tagIds: string[]; tagMatchMode: PerformerTagMatchMode; tags: I_performerTag[] }] }>();
const router = useRouter();
const visible = ref(false);
const classes = ref<I_performerTagClass[]>([]);
const tags = ref<I_performerTag[]>([]);
const draftTagIds = ref<string[]>([]);
const draftMatchMode = ref<PerformerTagMatchMode>('any');
const groups_C = computed(() => classes.value.map(item => ({ ...item, tags: tags.value.filter(tag => tag.performerTagClass_id === item.id) })).filter(item => item.tags.length));

const load = async () => {
  if (!props.performerBasesId) return;
  const result = await performerTagServer.data(props.performerBasesId);
  if (result?.status) {
    classes.value = result.data.tagClasses;
    tags.value = result.data.tags;
  }
};
const apply = () => {
  emits('change', { tagIds: [...draftTagIds.value], tagMatchMode: draftMatchMode.value, tags: tags.value.filter(tag => draftTagIds.value.includes(tag.id)) });
  visible.value = false;
};
const clear = () => {
  draftTagIds.value = [];
  emits('change', { tagIds: [], tagMatchMode: draftMatchMode.value, tags: [] });
};
const openManagement = () => {
  visible.value = false;
  router.push(`/performer/tags/${props.performerBasesId}`);
};
watch(() => props.performerBasesId, load, { immediate: true });
watch(visible, value => {
  if (value) {
    draftTagIds.value = [...props.modelValue];
    draftMatchMode.value = props.matchMode;
  }
});
</script>

<style scoped lang="scss">
.performer-tag-filter {
  .filter-header, .filter-footer { display: flex; align-items: center; justify-content: space-between; gap: 16px; }
  .filter-header { padding-bottom: 12px; border-bottom: 1px solid var(--el-border-color-lighter); }
  .filter-groups { padding: 6px 0; }
  .filter-group { display: grid; grid-template-columns: 72px 1fr; gap: 10px; padding: 8px 0; }
  .group-name { color: var(--el-text-color-secondary); line-height: 28px; }
  .group-tags { display: flex; flex-wrap: wrap; gap: 7px; }
  :deep(.el-checkbox-button__inner) { border: 1px solid var(--el-border-color); border-radius: 3px; padding: 7px 13px; box-shadow: none; }
  .filter-footer { padding-top: 12px; border-top: 1px solid var(--el-border-color-lighter); }
}
</style>
