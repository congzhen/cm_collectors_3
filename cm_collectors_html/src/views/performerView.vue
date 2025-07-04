<template>
  <div class="performer-view">
    <HeaderView class="header" :mode="E_headerMode.GoBack"></HeaderView>
    <div class="main">
      <performerDataList :performerBasesId="props.mainPerformerBasesId"></performerDataList>
    </div>
  </div>
</template>
<script setup lang="ts">
import HeaderView from './HeaderView.vue'
import performerDataList from '@/components/performer/performerDataList.vue'
import { E_headerMode } from '@/dataType/app.dataType'
import { appStoreData } from '@/storeData/app.storeData';
const store = {
  appStoreData: appStoreData(),
}
const props = defineProps({
  mainPerformerBasesId: {
    type: String,
    required: true,
  },
})


</script>
<style lang="scss" scoped>
.performer-view {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;

  .main {
    flex-grow: 1;
    display: flex;
    overflow: hidden;

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
