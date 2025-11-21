<template>
  <div class="cron-jobs">
    <div class="cron-jobs-btn">
      <el-button icon="Plus" type="success" plain @click="createHandle">创建新计划任务</el-button>
    </div>
  </div>
  <createCronJobsDrawer ref="createCronJobsDrawerRef"></createCronJobsDrawer>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import createCronJobsDrawer from './createCronJobsDrawer.vue';

const createCronJobsDrawerRef = ref<InstanceType<typeof createCronJobsDrawer>>();


const createHandle = () => {
  createCronJobsDrawerRef.value?.open();
};

</script>
<style lang="scss" scoped></style>
