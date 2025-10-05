<template>
  <el-select v-model="selectVal" clearable :style="{ width: props.width }" placeholder="刮削器配置文件" @change="changeHandle"
    @clear="handleClear" :multiple="props.multiple" :loading="loading">
    <el-option v-for="item, index in configs" :key="index" :label="item" :value="item"></el-option>
  </el-select>
</template>
<script setup lang="ts">
import { scraperDataServer } from '@/server/scraper.server';
import { ref, onMounted } from 'vue';

const selectVal = defineModel<string | string[]>({ type: [String, Array], default: "" as string | string[] });
const props = defineProps({
  width: {
    type: String,
    default: '100%',
  },
  multiple: {
    type: Boolean,
    default: false
  }
})
const emit = defineEmits(['change'])

const loading = ref(false);
const configs = ref<string[]>([]);

const init = async () => {
  await getConfigs();
}

const getConfigs = async () => {
  try {
    loading.value = true;
    configs.value = [];
    const result = await scraperDataServer.configs()
    if (result && result.status) {
      configs.value = result.data;
    } else if (result && result.msg) {
      console.log(result.msg);
    }
  } catch (error) {
    console.log(error);
  } finally {
    loading.value = false;
  }

}

const changeHandle = () => {
  emit('change', selectVal.value || '')
}
const handleClear = () => {
  if (props.multiple) {
    selectVal.value = [];
  } else {
    selectVal.value = '';
  }
}

onMounted(() => {
  init();
})

</script>
