<template>
  <tagRightClickMenu :tag="props.tag" @edit="() => emits('edit', props.tag)" @delete="() => emits('delete', props.tag)"
    @disable="() => emits('disable', props.tag)" @enable="() => emits('enable', props.tag)">
    <div :class="['tag-item', props.tag.status ? '' : 'disable']"
      :title="props.showResourceCount ? resourceCountTooltip_C : props.tag.name">
      <label>{{ props.tag.name }}</label>
      <span v-if="props.showResourceCount" class="tag-resource-count">
        {{ displayResourceCount_C }}
      </span>
    </div>
  </tagRightClickMenu>
</template>
<script lang="ts" setup>
import tagRightClickMenu from './tagRightClickMenu.vue';
import type { I_tag } from '@/dataType/tag.dataType';
import { computed, type PropType } from 'vue';
const props = defineProps({
  tag: {
    type: Object as PropType<I_tag>,
    required: true,
  },
  showResourceCount: {
    type: Boolean,
    default: true,
  },
})
const displayResourceCount_C = computed(() => {
  if (props.tag.resourceCount > 999) return '999+';
  return props.tag.resourceCount.toString();
})
const resourceCountTooltip_C = computed(() =>
  `${props.tag.name}：当前库共有 ${props.tag.resourceCount} 个资源使用此标签`
)
const emits = defineEmits(['edit', 'delete', 'disable', 'enable'])
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
  min-height: 30px;
  max-width: 100%;
  padding: 5px 20px;
  box-sizing: border-box;
  border-radius: 5px;
  border: 1px solid #4c4d4f;
  cursor: pointer;
  -moz-user-select: none;
  -webkit-user-select: none;
  -ms-user-select: none;
  user-select: none;
  position: relative;
  display: flex;
  align-items: center;

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
    display: block;
    min-width: 0;
    line-height: 18px;
    overflow-wrap: anywhere;
    word-break: break-word;
    cursor: pointer;
  }

  .tag-resource-count {
    position: absolute;
    top: -5px;
    right: -4px;
    z-index: 3;
    min-width: 13px;
    height: 12px;
    padding: 0 2px;
    box-sizing: border-box;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 3px;
    color: var(--tag-count-text, #67d4d7);
    background-color: var(--tag-count-bg, rgba(55, 198, 202, 0.14));
    font-size: 8px;
    font-weight: 650;
    line-height: 12px;
    pointer-events: none;
  }


}
</style>
