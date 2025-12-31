<template>
  <div class="performer-bases-list-view">
    <HeaderView class="header" :mode="E_headerMode.GoBack"></HeaderView>
    <div class="main">
      <el-tabs v-model="activeName">
        <el-tab-pane v-for="item, key in performerBasesIDS_C" :key="key"
          :label="store.performerBasesStoreData.getNameById(item)" :name="item">
          <performerDataList v-if="activeName == item" :performerBasesId="item"></performerDataList>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>
<script setup lang="ts">
import HeaderView from './HeaderView.vue'
import performerDataList from '@/components/performer/performerDataList.vue'
import { E_headerMode } from '@/dataType/app.dataType'
import { appStoreData } from '@/storeData/app.storeData'
import { filesBasesStoreData } from '@/storeData/filesBases.storeData'
import { performerBasesStoreData } from '@/storeData/performerBases.storeData'
import { ref, computed } from 'vue'

const store = {
  appStoreData: appStoreData(),
  performerBasesStoreData: performerBasesStoreData(),
  filesBasesStoreData: filesBasesStoreData(),
}
const props = defineProps({
  filesBasesId: {
    type: String,
    required: true,
  },
})
const activeName = ref(store.filesBasesStoreData.getMainPerformerBasesIdByFilesBasesId(props.filesBasesId))

const performerBasesIDS_C = computed(() => {
  return store.filesBasesStoreData.getPerformerBasesIdsByFilesBasesId(props.filesBasesId)
})


</script>
<style lang="scss" scoped>
.performer-bases-list-view {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;

  .main {
    flex-grow: 1;
    display: flex;
    overflow: hidden;

    :deep(.el-tabs__header) {
      margin-bottom: 10px;
      --el-tabs-header-height: 28px;

      .el-tabs__item {
        padding: 0 10px;
      }
    }

    .el-tab-pane {
      width: 100%;
      height: 100%;
    }

    .performer-info {
      flex-shrink: 0;
      width: 300px;
      height: 100%;
    }

    .performer-container {
      flex-grow: 1;
      margin-left: 1.5em;
      display: flex;
      flex-direction: column;

      .performer-search {
        flex-shrink: 0;
        display: flex;
      }

      .performer-list-main {
        flex-grow: 1;
        overflow: hidden;
        padding: 0.5em 0;

        .performer-list {
          list-style-type: none;
          display: flex;
          flex-wrap: wrap;
          align-content: flex-start;
          gap: 0.5em;

          li {
            width: 100px;
          }
        }
      }

      .performer-paging {
        flex-shrink: 0;
        padding-top: 5px;
      }
    }
  }
}
</style>
