<template>
  <el-select-v2 v-model="selectVal" clearable :style="{ width: props.width }" placeholder="演员" @change="changeHandle"
    @clear="handleClear" :multiple="props.multiple" filterable :options="options" :loading="loading"
    :filter-method="filterMethod" :props="{ label: 'name', value: 'id' }">
    <template #default="{ item }">
      <div class="performer-item">
        <div class="name">{{ item.name }}</div>
        <div class="aliasName" v-if="item.aliasName != ''">({{ item.aliasName }})</div>
      </div>
    </template>
  </el-select-v2>
</template>
<script setup lang="ts">
import { debounce } from '@/assets/debounce';
import type { I_performerBasic } from '@/dataType/performer.dataType';
import { performerServer } from '@/server/performer.server';
import { ElMessage } from 'element-plus';
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
  },
  performerBasesIds: {
    type: Array<string>,
    default: () => []
  }
})
const emit = defineEmits(['change'])
let list: I_performerBasic[] = [];
const options = ref<I_performerBasic[]>([]);
const loading = ref(false);


const init = async () => {
  await getPerformerList();
}
const getPerformerList = async () => {
  loading.value = true;
  const result = await performerServer.basicList(props.performerBasesIds)
  if (!result.status) {
    ElMessage.error(result.msg)
    return
  }
  list = result.data
  options.value = result.data
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
      return item.name.toLowerCase().includes(query) || item.aliasName.toLowerCase().includes(query) || item.keyWords.toLowerCase().includes(query)
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
