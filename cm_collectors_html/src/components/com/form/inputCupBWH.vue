<template>
  <div class="BWH">
    <div class="BWH_k">
      <el-input
        v-model="objectVal.bust"
        :maxlength="3"
        :formatter="(value: string) => value.replace(/[^\d]/g, '')"
      >
        <template #prepend>B</template>
      </el-input>
    </div>
    <div class="BWH_k">
      <el-input
        v-model="objectVal.waist"
        :maxlength="3"
        :formatter="(value: string) => value.replace(/[^\d]/g, '')"
      >
        <template #prepend>W</template>
      </el-input>
    </div>
    <div>
      <el-input
        v-model="objectVal.hip"
        :maxlength="3"
        :formatter="(value: string) => value.replace(/[^\d]/g, '')"
      >
        <template #prepend>H</template>
      </el-input>
    </div>
  </div>
</template>
<script setup lang="ts">
import type { I_cupBWH } from '@/dataType/app.dataType'
const objectVal = defineModel<I_cupBWH>({
  default: {
    bust: '',
    waist: '',
    hip: '',
  },
})
const props = defineProps({
  width: {
    type: String,
    default: '100%',
  },
})
</script>
<style lang="scss" scoped>
.BWH {
  display: flex;
  .BWH_k {
    margin-right: 5px;
  }
  :deep(.el-input-group__prepend) {
    padding: 0 7px;
  }
}
</style>
