<template>
  <div class="performer-data-list">
    <performerInfo class="performer-info" v-if="props.showPerformerInfo" :performer="currentShowPerformer">
    </performerInfo>
    <div class="performer-container">
      <div class="performer-index">
        <span v-for="item, index in indexChars" :key="index" :class="{ 'select-index': item === selectIndex }"
          @click="selectCharIndexHandle(item)">
          {{ item }}
        </span>
      </div>
      <div class="performer-container-main">
        <performerSearch class="performer-search" :admin="true" @add="addPerformerHandle" @recycleBin="recycleBinHandle"
          @search="changeSearchHandle" @scraper="scraperHandle">
        </performerSearch>
        <div class="performer-list-main" v-loading="loading">
          <el-scrollbar>
            <ul class="performer-list">
              <li v-for="(performer, index) in dataList" :key="index">
                <performerRightClickMenu :performer="performer" @search="searchPerformerHandle"
                  @edit="editPerformerHandle" @delete="deletePerformerHandle">
                  <performerBlock :performer="performer" :tool="true" :admin="true" :attrAge="true"
                    :attrNationality="true" @search="searchPerformerHandle"
                    @click.stop="clickPerformerHandle(performer)" @edit="editPerformerHandle(performer)"
                    @delete="deletePerformerHandle(performer)">
                  </performerBlock>
                </performerRightClickMenu>
              </li>
            </ul>
          </el-scrollbar>
        </div>
        <div class="performer-paging">
          <el-pagination background layout="total, prev, pager, next, jumper" v-model:current-page="currentPage"
            :total="dataCount" :page-size="pageSize" @change="changePageHandle" size="small" />
        </div>
      </div>
    </div>
  </div>
  <performerFormDrawer ref="performerFormDrawerRef" :performerBasesId="props.performerBasesId"
    @success="getDataListAndCount" />
  <performerRecycleBinDialog ref="performerRecycleBinDialogRef" :performerBasesId="props.performerBasesId"
    @success="getDataListAndCount">
  </performerRecycleBinDialog>
  <scraperPerformerDialog ref="scraperPerformerDialogRef" @success="getDataListAndCount">
  </scraperPerformerDialog>
</template>
<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import performerFormDrawer from '@/components/performer/performerFormDrawer.vue';
import performerRecycleBinDialog from '@/components/performer/performerRecycleBinDialog.vue';
import scraperPerformerDialog from '../importResource/scraperPerformerDialog.vue';
import performerSearch from '@/components/performer/performerSearch.vue';
import performerInfo from '@/components/performer/performerInfo.vue';
import performerBlock from '@/components/performer/performerBlock.vue';
import type { I_performer, I_search_performer } from '@/dataType/performer.dataType';
import { performerServer } from '@/server/performer.server';
import { ElMessage } from 'element-plus';
import { messageBoxConfirm } from '../../common/messageBox';
import { searchStoreData } from '@/storeData/search.storeData';
import { useRouter } from 'vue-router';
import performerRightClickMenu from './performerRightClickMenu.vue';
const router = useRouter()
const store = {
  searchStoreData: searchStoreData(),
}

const props = defineProps({
  performerBasesId: {
    type: String,
    default: '',
  },
  showPerformerInfo: {
    type: Boolean,
    default: true,
  },
})
const performerFormDrawerRef = ref<InstanceType<typeof performerFormDrawer>>();
const performerRecycleBinDialogRef = ref<InstanceType<typeof performerRecycleBinDialog>>();
const scraperPerformerDialogRef = ref<InstanceType<typeof scraperPerformerDialog>>();
const loading = ref(false);
const dataList = ref<I_performer[]>([]);
const dataCount = ref(0);
let fetchCount = true;
const currentPage = ref(1);
const pageSize = ref(90);
const indexChars = ref(['ALL', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'])
const selectIndex = ref('ALL');
let searchCondition: I_search_performer = {
  search: '',
  star: '',
  cup: '',
  charIndex: '',
}

const currentShowPerformer = ref<I_performer | undefined>(undefined);

const init = async () => {
  await getDataListAndCount(true);
  if (dataList.value.length > 0) {
    currentShowPerformer.value = dataList.value[0];
  }
}

const getDataListAndCount = async (fetchCountStatus: boolean = true) => {
  fetchCount = fetchCountStatus;
  await getDataList();
}
const getDataList = async () => {
  loading.value = true;
  searchCondition.charIndex = selectIndex.value;
  const result = await performerServer.dataList(props.performerBasesId, fetchCount, currentPage.value, pageSize.value, searchCondition);
  if (result && result.status) {
    dataList.value = result.data.dataList;
    if (fetchCount) {
      dataCount.value = result.data.total;
      fetchCount = false;
    }
  } else {
    ElMessage.error(result.msg);
  }
  loading.value = false;
}

const changePageHandle = () => {
  getDataList();
}

const clickPerformerHandle = (data: I_performer) => {
  currentShowPerformer.value = data;
}

const searchPerformerHandle = (data: I_performer) => {
  store.searchStoreData.setQueryPerformer(data.id, data.name)
  router.push(`/`)
}

const addPerformerHandle = () => {
  performerFormDrawerRef.value?.open('add')
}
const editPerformerHandle = (data: I_performer) => {
  performerFormDrawerRef.value?.open('edit', data)
}

const deletePerformerHandle = (performer: I_performer) => {
  messageBoxConfirm({
    text: '确定要删除吗？',
    successCallBack: async () => {
      const result = await performerServer.updateStatus(performer.id, false);
      if (result && result.status) {
        getDataListAndCount()
      } else {
        ElMessage.error(result.msg);
      }
    },
    failCallBack: () => {
      //console.log('取消删除')
    },
  })
}

const recycleBinHandle = () => {
  performerRecycleBinDialogRef.value?.open()
}

const changeSearchHandle = (search: I_search_performer) => {
  searchCondition = search;
  fetchCount = true;
  getDataList();
}
const selectCharIndexHandle = (charIndex: string) => {
  selectIndex.value = charIndex;
  fetchCount = true;
  getDataList();
}

const scraperHandle = () => {
  scraperPerformerDialogRef.value?.open(props.performerBasesId)
}


onMounted(async () => {
  await init()
})

</script>
<style lang="scss" scoped>
.performer-data-list {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  gap: 5px;

  .performer-info {
    flex-shrink: 0;
    width: 260px;
    height: 100%;
  }

  .performer-container {
    flex: 1;
    overflow: hidden;
    display: flex;
    gap: 5px;

    .performer-index {
      width: 40px;
      display: flex;
      flex-direction: column;
      gap: 3px;

      span {
        display: block;
        width: 100%;
        height: 21px;
        line-height: 21px;
        text-align: center;
        background-color: #262727;
        border-radius: 2px;
        cursor: pointer;
        /*禁止选择 */
        -webkit-user-select: none;
        -moz-user-select: none;
        -ms-user-select: none;
        user-select: none;
      }

      span:hover {
        background-color: #3d3f3f;
      }

      .select-index {
        background-color: #E67F23;
      }

      .select-index:hover {
        background-color: #E67F23;
      }
    }

    .performer-container-main {
      flex: 1;
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
