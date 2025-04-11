<template>
  <div class="performer-search">
    <el-button icon="DocumentAdd" v-if="props.addBtn" @click="emits('add')">新增</el-button>
    <inputSearch width="280px" />
    <selectStar width="200px" />
    <selectCup width="200px" />
  </div>
</template>
<script setup lang="ts">
import inputSearch from '../com/form/inputSearch.vue'
import selectStar from '../com/form/selectStar.vue'
import selectCup from '../com/form/selectCup.vue'
const props = defineProps({
  addBtn: {
    type: Boolean,
    default: false,
  },
})
const emits = defineEmits(['add'])
</script>
<style lang="scss" scoped>
.performer-search {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5em;
}
</style>
