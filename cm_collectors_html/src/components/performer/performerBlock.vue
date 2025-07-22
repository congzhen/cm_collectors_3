<template>
  <div class="performer-block" v-if="props.performer">
    <performerPhoto :performer="props.performer"></performerPhoto>
    <div class="performer-block-name">{{ props.performer.name }}</div>
    <div class="performer-block-attr" v-if="attr_C">
      <label v-if="props.attrNationality">{{ props.performer.nationality }}</label>
      <label v-if="props.attrAge && props.performer.birthday != ''" style="padding-left: 2px;">
        ({{ calculateAge(props.performer.birthday) }}岁)
      </label>
    </div>
    <div class="performer-block-tool" v-if="props.tool" @click.stop="emits('search', props.performer)">
      <div class="performer-block-tool-btn">
        <el-icon>
          <VideoCameraFilled />
        </el-icon>
      </div>
      <div class="performer-block-tool-btn displayNone" v-if="props.admin" @click.stop="emits('edit')">
        <el-icon>
          <Edit />
        </el-icon>
      </div>
      <div class="performer-block-tool-btn displayNone" v-if="props.admin" @click.stop="emits('delete')">
        <el-icon>
          <Delete />
        </el-icon>
      </div>
    </div>
  </div>
</template>
<script lang="ts" setup>
import { computed, type PropType } from 'vue'
import { calculateAge } from '@/assets/calculate'
import performerPhoto from './performerPhoto.vue'
import type { I_performer } from '@/dataType/performer.dataType'
const props = defineProps({
  performer: {
    type: Object as PropType<I_performer>,
    required: true,
  },
  attrAge: {
    type: Boolean,
    default: false,
  },
  attrNationality: {
    type: Boolean,
    default: false,
  },
  tool: {
    type: Boolean,
    default: false,
  },
  admin: {
    type: Boolean,
    default: false,
  },
})
const emits = defineEmits(['search', 'edit', 'delete'])
const attr_C = computed(() => {
  return props.attrAge || props.attrNationality
})
</script>
<style lang="scss" scoped>
.performer-block {
  max-width: 200px;
  border-radius: 5px;
  cursor: pointer;
  padding: 3px;
  overflow: hidden;
  background-color: #262727;

  &:hover {
    .performer-block-tool-btn.displayNone {
      display: inline-flex;
      align-items: center;
      justify-content: center;
    }
  }

  .performer-block-name {
    text-align: center;
    padding: 0.8em 0.3em;
    font-size: 0.8em;
    line-height: 1em;
  }

  .performer-block-attr {
    text-align: center;
    font-size: 0.8em;
    height: 1.5em;
    line-height: 1em;
  }

  .performer-block-tool {
    border-top: 1px solid #333;
    display: flex;
    padding-top: 0.3em;

    .performer-block-tool-btn {
      width: 32%;
      line-height: 1.5em;
      text-align: center;
      border-radius: 2px;


      &:hover {
        background-color: #333;
        color: #fff;
      }
    }

    .displayNone {
      display: none;
    }
  }
}
</style>
