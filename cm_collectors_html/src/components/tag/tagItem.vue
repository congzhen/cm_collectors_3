<template>
  <tagRightClickMenu :tag="props.tag" @edit="() => emits('edit', props.tag)" @delete="() => emits('delete', props.tag)"
    @disable="() => emits('disable', props.tag)" @enable="() => emits('enable', props.tag)">
    <div :class="['tag-item', props.tag.status ? '' : 'disable']">
      <label>{{ props.tag.name }}</label>
    </div>
  </tagRightClickMenu>
</template>
<script lang="ts" setup>
import tagRightClickMenu from './tagRightClickMenu.vue';
import type { I_tag } from '@/dataType/tag.dataType';
import type { PropType } from 'vue';
const props = defineProps({
  tag: {
    type: Object as PropType<I_tag>,
    required: true,
  },
})
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


}
</style>
