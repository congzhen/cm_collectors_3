<template>
  <div class="layout-cover-poster-style1">
    <el-scrollbar>
      <ul class="list-ul">
        <li v-for="(item, key) in props.dataList" :key="key">
          <contentStyle1 :resource="item" @click="selectResourcesHandle(item)"></contentStyle1>
        </li>
      </ul>
    </el-scrollbar>
  </div>
</template>
<script lang="ts" setup>
import contentStyle1 from '@/components/content/contentStyle1.vue';
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
.layout-cover-poster-style1 {
  width: 100%;
  height: 100%;
  overflow: hidden;

  .list-ul {
    list-style-type: none;
    display: flex;
    flex-wrap: wrap;
    gap: 0.4em;
    padding-bottom: 1em;
  }
}
</style>
