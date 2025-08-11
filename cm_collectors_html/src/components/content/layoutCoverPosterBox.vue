<template>
  <div class="layout-cover-poster-box">
    <el-scrollbar>
      <ul class="list-ul">
        <li v-for="(item, key) in props.dataList" :key="key">
          <contentStyle2 :resource="item" @click="selectResourcesHandle(item)"></contentStyle2>
        </li>
      </ul>
    </el-scrollbar>
  </div>
</template>
<script lang="ts" setup>
import contentStyle2 from '@/components/content/contentStyle2.vue';
import type { I_resource } from '@/dataType/resource.dataType';
import type { PropType } from 'vue';
const props = defineProps({
  dataList: {
    type: Array as PropType<I_resource[]>,
    default: () => [],
  },
})
const emits = defineEmits(['selectResources']);

const selectResourcesHandle = (item: I_resource) => {
  emits('selectResources', item)
}

</script>
<style lang="scss" scoped>
.layout-cover-poster-box {
  width: 100%;
  height: 100%;
  overflow: hidden;

  .list-ul {
    list-style-type: none;
    display: flex;
    flex-wrap: wrap;
    gap: 20px;
    padding-bottom: 1em;
  }
}
</style>
