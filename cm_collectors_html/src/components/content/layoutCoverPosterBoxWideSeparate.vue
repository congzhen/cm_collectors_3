<template>
  <div class="layout-cover-poster-box-wide-separate">
    <el-scrollbar ref="scrollbarRef">
      <ul class="list-ul" :class="{ 'mobile-layout': isMobileDevice }">
        <li v-for="(item, key) in props.dataList" :key="key">
          <contentStyle3 :resource="item" @click="selectResourcesHandle(item)"></contentStyle3>
        </li>
      </ul>
    </el-scrollbar>
  </div>
</template>
<script lang="ts" setup>
import { isMobile } from '@/assets/mobile';
import contentStyle3 from '@/components/content/contentStyle3.vue';
import type { I_resource } from '@/dataType/resource.dataType';
import type { ElScrollbar } from 'element-plus';
import { onMounted, ref, type PropType } from 'vue';

const props = defineProps({
  dataList: {
    type: Array as PropType<I_resource[]>,
    default: () => [],
  },
})
const emits = defineEmits(['selectResources']);

const scrollbarRef = ref<InstanceType<typeof ElScrollbar>>();

const isMobileDevice = ref(false);
const selectResourcesHandle = (item: I_resource) => {
  emits('selectResources', item)
}
const change = () => {
  scrollbarRef.value?.setScrollTop(0);
};
onMounted(() => {
  isMobileDevice.value = isMobile();
});
defineExpose({ change });
</script>
<style scoped lang="scss">
.layout-cover-poster-box-wide-separate {
  width: 100%;
  height: 100%;
  overflow: hidden;

  .list-ul {
    list-style-type: none;
    display: flex;
    flex-wrap: wrap;
    gap: 2em;
    padding-bottom: 1em;

    &.mobile-layout {
      display: grid;
      grid-template-columns: repeat(2, 1fr);
      gap: 0.8em;
      padding-bottom: 0.8em;
    }
  }
}

.mobile-layout li {
  /*
  aspect-ratio: 158 / 214;
  */
}

.mobile-layout li :deep(.content-style1) {
  width: 100% !important;
  height: 100% !important;
}
</style>
