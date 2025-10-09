<template>
  <div class="layout-cover-poster-box">
    <el-scrollbar ref="scrollbarRef">
      <ul class="list-ul" :style="{ gap: gap_C }">
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
import { appStoreData } from '@/storeData/app.storeData';
import type { ElScrollbar } from 'element-plus';
import { computed, ref, type PropType } from 'vue';
const store = {
  appStoreData: appStoreData(),
}
const props = defineProps({
  dataList: {
    type: Array as PropType<I_resource[]>,
    default: () => [],
  },
})
const emits = defineEmits(['selectResources']);

const scrollbarRef = ref<InstanceType<typeof ElScrollbar>>();
const gap_C = computed(() => {
  return (store.appStoreData.currentConfigApp.coverPosterGap || 4.8) + 'px';
});
const selectResourcesHandle = (item: I_resource) => {
  emits('selectResources', item)
}
const change = () => {
  scrollbarRef.value?.setScrollTop(0);
};

defineExpose({ change });
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

    // 移动端适配
    @media (max-width: 768px) {
      gap: 20px;

      :deep(.content-style2) {
        width: 100%;
        max-width: 100%;

        .content-cover {
          width: 140px !important;
          height: auto !important;
          align-self: flex-start;
        }

        .content-info {
          width: calc(100% - 152px) !important;
        }
      }
    }
  }
}
</style>
