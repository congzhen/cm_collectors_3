<template>
  <div class="layout-cover-poster-simple">
    <el-scrollbar ref="scrollbarRef">
      <ul class="list-ul" :class="{ 'mobile-layout': isMobileDevice }" :style="contentLayoutStyle_C">
        <li v-for="(item, key) in props.dataList" :key="key">
          <contentStyleSimpleExpand :resource="item" @click="selectResourcesHandle(item)"></contentStyleSimpleExpand>
        </li>
      </ul>
    </el-scrollbar>
  </div>
</template>
<script lang="ts" setup>
import contentStyleSimpleExpand from './contentStyleSimpleExpand.vue';
import type { I_resource } from '@/dataType/resource.dataType';
import type { ElScrollbar } from 'element-plus';
import { ref, type PropType, onMounted } from 'vue';
import { isMobile } from '@/assets/mobile';

import { contentLayoutStyle_C } from '@/common/content'

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
<style lang="scss" scoped>
.layout-cover-poster-simple {
  width: 100%;
  height: 100%;
  overflow: hidden;

  .list-ul {
    list-style-type: none;
    display: flex;
    flex-wrap: wrap;
    gap: 0.4em;
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
