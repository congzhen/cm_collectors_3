<template>
  <div class="content-tag" :style="{
    backgroundColor: props.bgColor,
    color: props.color,
    border: '1px solid ' + props.bgColor,
    borderRadius: props.borderRadius,
    fontSize: props.fontSize,
  }">
    {{ props.title }}
  </div>
</template>
<script lang="ts" setup>
const props = defineProps({
  title: {
    type: String,
    required: true,
  },
  bgColor: {
    type: String,
    default: 'rgba(244, 54, 16, 0.4)',
  },
  color: {
    type: String,
    default: '#FFFFFF',
  },
  borderRadius: {
    type: String,
    default: '0.3em',
  },
  fontSize: {
    type: String,
    default: '12px',
  },
})
</script>
<style lang="scss" scoped>
.content-tag {
  padding: 0 5px;
  font-size: 12px;
  line-height: 1.3em;
}
</style>
