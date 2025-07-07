<template>
  <div class="performer-data-list">
    <performerInfo class="performer-info" v-if="props.showPerformerInfo" :performer="currentShowPerformer">
    </performerInfo>
    <div class="performer-container">
      <performerSearch class="performer-search" :admin="true" @add="addPerformerHandle" @recycleBin="recycleBinHandle"
        @search="changeSearchHandle">
      </performerSearch>
      <div class="performer-list-main" v-loading="loading">
        <el-scrollbar>
          <ul class="performer-list">
            <li v-for="(performer, index) in dataList" :key="index">
              <performerBlock :performer="performer" :tool="true" :admin="true" :attrAge="true" :attrNationality="true"
                @click.stop="clickPerformerHandle(performer)" @edit="editPerformerHandle(performer)"
                @delete="deletePerformerHandle(performer)">
              </performerBlock>
            </li>
          </ul>
        </el-scrollbar>
      </div>
      <div class="performer-paging">
        <el-pagination background layout="total, prev, pager, next, jumper" v-model:current-page="currentPage"
          :total="dataCount" :page-size="pageSize" @change="changePageHandle" />
      </div>
    </div>
  </div>
  <performerFormDrawer ref="performerFormDrawerRef" :performerBasesId="props.performerBasesId"
    @success="getDataListAndCount" />
  <performerRecycleBinDialog ref="performerRecycleBinDialogRef" :performerBasesId="props.performerBasesId"
    @success="getDataListAndCount">
  </performerRecycleBinDialog>
</template>
<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import performerFormDrawer from '@/components/performer/performerFormDrawer.vue';
import performerRecycleBinDialog from '@/components/performer/performerRecycleBinDialog.vue';
import performerSearch from '@/components/performer/performerSearch.vue';
import performerInfo from '@/components/performer/performerInfo.vue';
import performerBlock from '@/components/performer/performerBlock.vue';
import type { I_performer, I_search_performer } from '@/dataType/performer.dataType';
import { performerServer } from '@/server/performer.server';
import { ElMessage } from 'element-plus';
import { messageBoxConfirm } from '../com/feedback/messageBox';

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
const loading = ref(false);
const dataList = ref<I_performer[]>([]);
const dataCount = ref(0);
let fetchCount = true;
const currentPage = ref(1);
const pageSize = ref(75);

let searchCondition: I_search_performer = {
  search: '',
  star: '',
  cup: '',
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

  .performer-info {
    flex-shrink: 0;
    width: 260px;
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
</style>
