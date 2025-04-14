<template>
  <div class="performer-block">
    <performerPhoto></performerPhoto>
    <div class="performer-block-name">小栗みゆ</div>
    <div class="performer-block-attr" v-if="attr_C">
      <label v-if="props.attrNationality">日本</label>
      <label v-if="props.attrAge">(28岁)</label>
    </div>
    <div class="performer-block-tool" v-if="props.tool">
      <div class="performer-block-tool-btn">
        <el-icon><VideoCameraFilled /></el-icon>
      </div>
      <div class="performer-block-tool-btn" v-if="props.admin">
        <el-icon><Edit /></el-icon>
      </div>
      <div class="performer-block-tool-btn" v-if="props.admin">
        <el-icon><Delete /></el-icon>
      </div>
    </div>
  </div>
</template>
<script lang="ts" setup>
import { computed } from 'vue'
import performerPhoto from './performerPhoto.vue'
const props = defineProps({
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
  .performer-block-name {
    text-align: center;
    padding: 0.8em 0.5em;
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
  }
}
</style>
