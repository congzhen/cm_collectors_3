<template>
  <div class="content-list">
    <el-scrollbar>
      <ul class="list-ul">
        <li v-for="(item, key) in props.dataList" :key="key">
          <contentItem :resource="item" @click="selectResourcesHandle(item)"></contentItem>
        </li>
      </ul>
    </el-scrollbar>
  </div>
</template>
<script lang="ts" setup>
import contentItem from '@/components/content/contentItem.vue'
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
.content-list {
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
