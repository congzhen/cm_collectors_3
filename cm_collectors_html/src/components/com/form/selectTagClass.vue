<template>
  <el-select-v2 v-model="selectVal" clearable :style="{ width: props.width }" placeholder="标签分类" @change="changeHandle"
    @clear="handleClear" :multiple="props.multiple" filterable :options="options" :loading="loading"
    :filter-method="filterMethod" :props="{ label: 'name', value: 'id' }">
  </el-select-v2>
</template>
<script setup lang="ts">
import { debounce } from '@/assets/debounce';
import type { I_tagClass } from '@/dataType/tag.dataType';
import { tagServer } from '@/server/tag.server';
import { ElMessage } from 'element-plus';
import { ref, onMounted, type PropType } from 'vue';
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
    required: true
  }
})
const emit = defineEmits(['change'])
let list: I_tagClass[] = [];
const options = ref<I_tagClass[]>([]);
const loading = ref(false);


const init = async () => {
  await getTagClassList();
}
const getTagClassList = async () => {
  loading.value = true;
  if (props.dataSource === 'database') {
    const result = await tagServer.tagClassListByFilesBasesId(props.filesBasesId)
    if (!result.status) {
      ElMessage.error(result.msg)
      return
    }
    list = result.data
    options.value = result.data
  } else {
    list = store.appStoreData.currentTagClass;
    options.value = store.appStoreData.currentTagClass;
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
      return item.name.toLowerCase().includes(query)
    })
  } else {
    options.value = list
  }
  loading.value = false
}, 200)


onMounted(async () => {
  await init();
})

</script>
<style lang="scss" scoped>
.performer-item {
  display: flex;
  gap: 10px;
}
</style>
