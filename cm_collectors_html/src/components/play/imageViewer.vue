<template>
  <!-- 使用 Element Plus 的图片查看器 -->
  <el-image-viewer v-if="showImageViewer" :url-list="imageList_C" :initial-index="initialIndex"
    @close="closeImageViewer" @switch="handleSwitch" />
</template>

<script lang="ts" setup>
import { ref, computed, type PropType } from 'vue';

const props = defineProps({
  imageList: {
    type: Array as PropType<string[]>,
    required: true,
  },
})

const showImageViewer = ref(false);

const initialIndex = ref(0);
// 用于跟踪当前查看的图片索引
const currentIndex = ref(0);

// 用于存储已加载的图片URL
const loadedImages = ref<Record<number, string>>({});
// 用于存储正在加载的图片URL
const loadingImages = ref<Record<number, boolean>>({});

const imageList_C = computed(() => {
  // 初始化图片列表，优先使用已加载的图片，否则使用占位符
  const result: string[] = [];

  props.imageList.forEach((_, index) => {
    if (loadedImages.value[index]) {
      // 如果图片已加载，使用真实URL
      result[index] = loadedImages.value[index];
    } else {
      // 如果图片未加载，使用占位符
      result[index] = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMzAwIiBoZWlnaHQ9IjMwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSJ0cmFuc3BhcmVudCIvPjx0ZXh0IHg9IjUwJSIgeT0iNTAlIiBmb250LWZhbWlseT0iQXJpYWwsIEhlbHZldGljYSwgc2Fucy1zZXJpZiIgZm9udC1zaXplPSIxOCIgZmlsbD0id2hpdGUiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGRvbWluYW50LWJhc2VsaW5lPSJtaWRkbGUiPkxvYWRpbmcuLi48L3RleHQ+PC9zdmc+';
    }
  });

  // 预加载当前图片和相邻图片
  preloadImage(currentIndex.value);

  // 预加载前后几张图片
  for (let i = Math.max(0, currentIndex.value - 2); i <= Math.min(props.imageList.length - 1, currentIndex.value + 2); i++) {
    preloadImage(i);
  }

  return result;
});

// 预加载指定索引的图片
const preloadImage = (index: number) => {
  // 如果图片已加载或正在加载，则跳过
  if (loadedImages.value[index] || loadingImages.value[index]) {
    return;
  }

  // 标记为正在加载
  loadingImages.value[index] = true;

  // 创建图片对象进行预加载
  const img = new Image();
  img.onload = () => {
    // 加载成功，保存到已加载列表
    loadedImages.value[index] = props.imageList[index];
    loadingImages.value[index] = false;
  };

  img.onerror = () => {
    // 加载失败，使用默认错误图片
    loadedImages.value[index] = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMzAwIiBoZWlnaHQ9IjMwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSJ0cmFuc3BhcmVudCIvPjx0ZXh0IHg9IjUwJSIgeT0iNTAlIiBmb250LWZhbWlseT0iQXJpYWwsIEhlbHZldGljYSwgc2Fucy1zZXJpZiIgZm9udC1zaXplPSIxOCIgZmlsbD0id2hpdGUiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGRvbWluYW50LWJhc2VsaW5lPSJtaWRkbGUiPkltYWdlIEVycm9yPC90ZXh0Pjwvc3ZnPg==';
    loadingImages.value[index] = false;
  };

  // 开始加载
  img.src = props.imageList[index];
};

// 处理图片切换事件
const handleSwitch = (index: number) => {
  currentIndex.value = index;
  // 触发重新计算 imageList_C 来加载新图片
  // 由于 currentIndex 是响应式的，改变它会触发 computed 重新计算
};

// 打开图片查看器
const openImageViewer = (index: number = 0) => {
  initialIndex.value = index;
  currentIndex.value = index; // 初始化当前索引
  showImageViewer.value = true;
};

// 关闭图片查看器
const closeImageViewer = () => {
  showImageViewer.value = false;
};

// 导出函数
defineExpose({
  openImageViewer,
  closeImageViewer,
});
</script>

<style lang="scss" scoped></style>
