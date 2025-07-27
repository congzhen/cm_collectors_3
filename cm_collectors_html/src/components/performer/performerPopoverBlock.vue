<template>
  <el-popover placement="left" trigger="click" width="380px">
    <template #reference>
      <performerBlock :performer="props.performer"></performerBlock>
    </template>
    <performerDetails :performer="props.performer" :issuing-date="props.issuingDate">
    </performerDetails>
  </el-popover>
</template>
<script lang="ts" setup>
import performerDetails from './performerDetails.vue'
import performerBlock from './performerBlock.vue'
import type { I_performer } from '@/dataType/performer.dataType';
import type { PropType } from 'vue';
const props = defineProps({
  performer: {
    type: Object as PropType<I_performer>,
    required: true,
  },
  issuingDate: {
    type: String,
    default: ''
  }
})
</script>
<style lang="scss" scoped></style>
