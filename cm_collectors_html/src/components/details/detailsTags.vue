<template>
  <div class="details-tag-list">
    <el-tag type="info" effect="plain" size="large" v-for="item, key in tags_C" :key="key"
      @click="setSearchHandle(item)">
      {{ item.name }}
    </el-tag>
  </div>
</template>
<script lang="ts" setup>
import { E_tagType } from '@/dataType/app.dataType';
import type { I_resource } from '@/dataType/resource.dataType';
import type { I_tag } from '@/dataType/tag.dataType';
import { searchStoreData } from '@/storeData/search.storeData';
import { computed, type PropType } from 'vue';
const store = {
  searchStoreData: searchStoreData(),
}
const props = defineProps({
  resource: {
    type: Object as PropType<I_resource>,
    required: true
  },
})

const tags_C = computed(() => {
  return props.resource.tags
})

const setSearchHandle = (tag: I_tag) => {
  console.log(tag);
  store.searchStoreData.setQuery(E_tagType.DiyTag, tag.id, tag.tagClass_id);
}

</script>
<style lang="scss" scoped>
.details-tag-list {
  padding: 5px;
  display: flex;
  flex-wrap: wrap;
  gap: 5px;

  .el-tag {
    cursor: pointer;
    user-select: none;
  }
}
</style>
