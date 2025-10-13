<template>
  <el-select-v2 v-model="selectVal" clearable :style="{ width: props.width }" placeholder="标签" @change="changeHandle"
    @clear="handleClear" :multiple="props.multiple" filterable :options="options" :loading="loading"
    :filter-method="filterMethod" :props="selectProps">
  </el-select-v2>
</template>
<script setup lang="ts">
import { debounce } from '@/assets/debounce';
import type { I_tag } from '@/dataType/tag.dataType';
import { tagServer } from '@/server/tag.server';
import { ElMessage } from 'element-plus';
import { ref, onMounted, onActivated, type PropType, computed } from 'vue';
import { appStoreData } from '@/storeData/app.storeData';
const store = {
  appStoreData: appStoreData(),
}
const selectVal = defineModel<string | string[]>({ type: [String, Array], default: "" as string | string[] });
const props = defineProps({
  width: {
    type: String,
    default: '100%',
  },
  dataSource: {
    type: String as PropType<'database' | 'store'>,
    default: 'database',
  },
  multiple: {
    type: Boolean,
    default: false
  },
  filesBasesId: {
    type: String,
    default: ''
  },
  tagClassId: {
    type: String,
    default: ''
  }
})
const emit = defineEmits(['change'])
let list: I_tag[] = [];
const options = ref<I_tag[]>([]);
const loading = ref(false);

// 使用固定的对象引用，避免每次渲染时创建新对象导致滚动位置重置
const selectProps = computed(() => ({
  label: 'name',
  value: 'id'
}));


const init = async () => {
  list = [];
  options.value = [];
  await getTagList();
}
const getTagList = async () => {
  loading.value = true;
  if (props.dataSource === 'database') {
    const apiCall = props.tagClassId === ''
      ? tagServer.tagListByFilesBasesId(props.filesBasesId)
      : tagServer.tagListByTagClassId(props.tagClassId);

    const result = await apiCall;
    if (!result.status) {
      ElMessage.error(result.msg)
      return
    }
    list = result.data
    options.value = result.data
  } else {
    if (props.tagClassId != '') {
      list = store.appStoreData.currentTag.filter(item => item.tagClass_id === props.tagClassId);
    } else {
      list = store.appStoreData.currentTag;
    }

    options.value = list;
  }

  loading.value = false;
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

const filterMethod = debounce((query: string) => {
  loading.value = true
  if (query !== '') {
    query = query.toLowerCase()
    options.value = list.filter((item) => {
      return item.name.toLowerCase().includes(query) || item.keyWords.toLowerCase().includes(query)
    })
  } else {
    options.value = list
  }
  loading.value = false
}, 200)


onMounted(async () => {
  await init();
})
onActivated(async () => {
  await init();
})

</script>
<style lang="scss" scoped>
.performer-item {
  display: flex;
  gap: 10px;
}
</style>
