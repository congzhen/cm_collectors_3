<template>
  <drawerCommon ref="drawerCommonRef" title="播放列表" :btnSubmit="false">
    <div class="play-list" v-loading="loading">
      <el-empty v-if="dataList.length === 0" description="播放列表为空" />
      <div class="resource" v-for="item, key in dataList" :key="key">
        <div class="resource-title">
          <label>{{ item.title }}</label>
          <el-icon size="16" color="#F56C6C" @click="removePlayList(item.id, key)" class="delete-icon">
            <Delete />
          </el-icon>
        </div>
        <div class="drama-series-container">
          <div class="drama-series" v-for="ds, keyDs in item.dramaSeries" :key="keyDs" @click="playVideo(item, ds.id)">
            <div class="drama-series-info">
              <el-icon size="16" class="play-icon">
                <VideoPlay />
              </el-icon>
              <span class="file-name">{{ getFinalPathSegment(ds.src) }}</span>
            </div>
            <div class="drama-series-actions">
              <el-icon size="14" class="more-options">
                <MoreFilled />
              </el-icon>
            </div>
          </div>
        </div>
      </div>
    </div>
    <template #footerBtn>
      <el-button @click="clearPlayListHandle" type="danger" plain>清空播放列表</el-button>
    </template>
  </drawerCommon>
</template>
<script lang="ts" setup>
import { getPlayListResource, playListClear, playListRemove } from '@/common/playList';
import drawerCommon from '@/components/com/dialog/drawer-common.vue';
import type { I_resource } from '@/dataType/resource.dataType';
import { ref } from 'vue';
import { getFinalPathSegment } from '@/assets/tool'
import { MoreFilled } from '@element-plus/icons-vue'
import { playResource } from '@/common/play';

const drawerCommonRef = ref<InstanceType<typeof drawerCommon>>();

const dataList = ref<I_resource[]>([]);
const loading = ref(false);
const init = () => {
  getPlayList();
}

const getPlayList = async () => {
  try {
    loading.value = true;
    const list = await getPlayListResource();
    dataList.value = list;
  } finally {
    loading.value = false;
  }
}

const playVideo = (resource: I_resource, dramaSeriesId: string) => {
  playResource(resource, dramaSeriesId)
}

const clearPlayListHandle = () => {
  playListClear();
  getPlayList();
}

const removePlayList = async (resourceId: string, index: number) => {
  playListRemove(resourceId)
  dataList.value.splice(index, 1);
}

const open = () => {
  init();
  drawerCommonRef.value?.open();
};

defineExpose({
  open,
});
</script>
<style scoped lang="scss">
.play-list {
  .resource {
    margin-bottom: 15px;
    border-radius: 10px;
    background-color: #f5f5f5;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    overflow: hidden;

    .resource-title {
      font-size: 15px;
      font-weight: bold;
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 12px 15px;
      background-color: #ebebeb;
      color: #333;

      label {
        flex: 1;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        padding-right: 10px;
      }

      .delete-icon {
        cursor: pointer;
        transition: transform 0.2s;

        &:hover {
          transform: scale(1.1);
        }
      }
    }

    .drama-series-container {
      padding: 5px 0;

      .drama-series {
        font-size: 13px;
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 10px 15px;
        cursor: pointer;
        transition: all 0.2s;
        border-left: 3px solid transparent;

        &:hover {
          background-color: #e1e1e1;
          border-left: 3px solid #409eff;
        }

        .drama-series-info {
          flex: 1;
          display: flex;
          align-items: center;
          overflow: hidden;

          .play-icon {
            margin-right: 10px;
            color: #409eff;
          }

          .file-name {
            flex: 1;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
            color: #333;
          }
        }

        .drama-series-actions {
          display: flex;
          align-items: center;

          .more-options {
            opacity: 0;
            transition: opacity 0.2s;
            color: #999;
            cursor: pointer;

            &:hover {
              color: #409eff;
            }
          }
        }

        &:hover .more-options {
          opacity: 1;
        }
      }

      .drama-series:not(:last-child) {
        border-bottom: 1px solid #e4e4e4;
      }
    }
  }
}
</style>
