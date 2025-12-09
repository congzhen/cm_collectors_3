<template>
  <div class="details-sample-images" v-if="props.resource && sampleImagesUrl.length > 0">
    <div class="images-container" :style="{ gridTemplateColumns: `repeat(${props.columns || 3}, minmax(0, 1fr))` }">
      <div v-for="(image, index) in sampleImagesUrls_C" :key="index" class="image-item">
        <el-image :src="image" fit="cover" class="sample-image" @click="openImageViewer(index)">
          <template #placeholder>
            <div class="image-placeholder">
              <el-icon>
                <Picture />
              </el-icon>
            </div>
          </template>
          <template #error>
            <div class="image-error">
              <el-icon>
                <Picture />
              </el-icon>
            </div>
          </template>
        </el-image>
      </div>
    </div>
    <imageViewer ref="imageViewerRef" :imageList="sampleImagesUrls_C"></imageViewer>
  </div>

</template>

<script setup lang="ts">
import { getSamplePhoto } from '@/common/photo';
import type { I_resource } from '@/dataType/resource.dataType';
import { resourceServer } from '@/server/resource.server';
import { ref, onMounted, computed, type PropType, watch } from 'vue'
import { ElImage, ElIcon } from 'element-plus'
import { Picture } from '@element-plus/icons-vue'
import { appStoreData } from '@/storeData/app.storeData';
import imageViewer from '@/components/play/imageViewer.vue';
const store = {
  appStoreData: appStoreData(),
}
const props = defineProps({
  resource: {
    type: Object as PropType<I_resource> | undefined,
    default: undefined
  },
  columns: {
    type: Number,
    default: 3
  }
})
const emits = defineEmits(['loadImageComplete'])
// 监听 props.resource 变化，重新初始化
watch(() => props.resource, () => {
  init();
}, { deep: true })

const imageViewerRef = ref<InstanceType<typeof imageViewer>>();
const sampleImagesUrl = ref<string[]>([])

//限制剧照最大显示数量
const sampleImagesUrls_C = computed(() => {
  //获取数组中前N个
  return sampleImagesUrl.value.slice(0, store.appStoreData.currentConfigApp.sampleShowMax)
})

const init = async () => {
  sampleImagesUrl.value = [];
  await getSampleImages();
}

const getSampleImages = async () => {
  if (!props.resource) return;
  const result = await resourceServer.sampleImages(props.resource.id, store.appStoreData.currentConfigApp.sampleFolder);
  if (result && result.status) {
    result.data.forEach((imagePath: string) => {
      const url = getSamplePhoto(props.resource, imagePath);
      if (url != '') {
        sampleImagesUrl.value.push(getSamplePhoto(props.resource, imagePath));
      }
    });
    if (sampleImagesUrl.value.length > 0) {
      emits('loadImageComplete', sampleImagesUrl.value);
    }
  }
}
const openImageViewer = (index: number) => {
  imageViewerRef.value?.openImageViewer(index);
}
onMounted(async () => {
  await init();
})
</script>

<style scoped lang="scss">
.details-sample-images {
  width: 100%;
}


.images-container {
  display: grid;
  gap: 12px;
  width: 100%;
}

.image-item {
  aspect-ratio: 16/9;
}

.sample-image {
  width: 100%;
  height: 100%;
  border-radius: 4px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);

  :deep(.el-image__inner) {
    cursor: pointer;
    transition: transform 0.3s ease;
  }

  :deep(.el-image__inner):hover {
    transform: scale(1.05);
  }
}

.image-placeholder,
.image-error {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  height: 100%;
  background-color: var(--el-fill-color-light);
  color: var(--el-text-color-secondary);

  .el-icon {
    font-size: 24px;
  }
}

.image-error {
  background-color: var(--el-color-danger-light-9);
}
</style>
