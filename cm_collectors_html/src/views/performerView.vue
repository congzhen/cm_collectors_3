<template>
  <div class="performer-view">
    <HeaderView class="header" :mode="E_headerMode.GoBack"></HeaderView>
    <div class="main">
      <performerInfo class="performer-info"></performerInfo>
      <div class="performer-container">
        <performerSearch
          class="performer-search"
          :addBtn="true"
          @add="addPerformerHandle"
        ></performerSearch>
        <div class="performer-list-main">
          <el-scrollbar>
            <ul class="performer-list">
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>

              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>

              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>

              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>

              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>

              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
              <li><performerBlock :attrAge="true" :attrNationality="true"></performerBlock></li>
            </ul>
          </el-scrollbar>
        </div>
        <div class="performer-paging">
          <el-pagination background layout="total, prev, pager, next, jumper" :total="1000" />
        </div>
      </div>
    </div>
  </div>
  <performerFormDrawer ref="performerFormDrawerRef" />
</template>
<script setup lang="ts">
import HeaderView from './HeaderView.vue'
import performerInfo from '@/components/performer/performerInfo.vue'
import performerSearch from '@/components/performer/performerSearch.vue'
import performerBlock from '@/components/performer/performerBlock.vue'
import performerFormDrawer from '@/components/performer/performerFormDrawer.vue'
import { E_headerMode } from '@/dataType/app.dataType'
import { useRouter } from 'vue-router'
import { ref } from 'vue'
const router = useRouter()

const performerFormDrawerRef = ref<InstanceType<typeof performerFormDrawer>>()

const addPerformerHandle = () => {
  performerFormDrawerRef.value?.open()
}
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
