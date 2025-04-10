<template>
  <el-popover placement="left" trigger="click" width="380px">
    <template #reference>
      <performerBlock></performerBlock>
    </template>
    <performerDetails></performerDetails>
  </el-popover>
</template>
<script lang="ts" setup>
import performerDetails from './performerDetails.vue'
import performerBlock from './performerBlock.vue'
</script>
<style lang="scss" scoped></style>
