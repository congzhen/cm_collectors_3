<template>
  <div class="play-movies-container">
    <HeaderView class="header" :mode="E_headerMode.GoBack" :title="resourceInfo?.title || ''"></HeaderView>
    <div class="main-container" v-loading="loading">
      <div class="main" v-if="resourceInfo">
        <div class="main-left">
          <div>
            <videoPlay ref="videoPlayRef" />
          </div>
          <div class="info-base">
            <div v-if="resourceInfo.issueNumber">
              版号、番号、刊号: {{ resourceInfo.issueNumber }}
            </div>
            <div v-if="resourceInfo.issuingDate && resourceInfo.issuingDate != ''">
              年份: {{ resourceInfo.issuingDate }}
            </div>
            <div v-if="resourceInfo.country != ''">
              国家: {{ appLang.country(resourceInfo.country) }}
            </div>
            <div v-if="resourceInfo.definition != ''">
              清晰度: {{ appLang.definition(resourceInfo.definition) }}
            </div>
            <div> 收录时间: {{ resourceInfo.addTime }}</div>
            <el-rate v-model="resourceInfo.stars" disabled />
          </div>
          <el-alert class="tagAlert" title="演员" type="success" :closable="false" />
          <div class="performer-list">
            <div class="performer-item" v-for="performer, key in resourceInfo.performers" :key="key">
              <performerDetails :performer="performer" :issuing-date="resourceInfo.issuingDate" :performerBtn="false"
                :roundAvatar="true">
              </performerDetails>
            </div>
          </div>
          <el-alert class="tagAlert" title="标签" type="warning" :closable="false" />
          <div class="tag-list">
            <el-tag type="info" effect="plain" size="large" v-for="item, key in resourceInfo.tags" :key="key">
              {{ item.name }}
            </el-tag>
          </div>
          <el-alert class="tagAlert" title="摘要" type="info" :closable="false" />
          <div class="abstract">
            {{ resourceInfo.abstract }}
          </div>
          <div class="c-height"></div>
        </div>
        <div class="main-right">
          <el-image :src="getResourceCoverPoster(resourceInfo)" fit="cover" />
          <div class="title">{{ resourceInfo.title }}</div>
          <resourceDramaSeriesList :drama-series="resourceInfo.dramaSeries" :selected-id="selectedDramaSeriesId"
            :show-mode="store.appStoreData.currentFilesBasesAppConfig.detailsDramaSeriesMode"
            @play-resource-drama-series="playResourceDramaSeriesHandle">
          </resourceDramaSeriesList>
          <div class="c-height"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import HeaderView from '../HeaderView.vue'
import videoPlay from "@/components/play/videoPlay.vue";
import resourceDramaSeriesList from '@/components/resource/resourceDramaSeriesList.vue'
import performerDetails from '@/components/performer/performerDetails.vue'
import { E_headerMode } from '@/dataType/app.dataType'
import type { I_resource, I_resourceDramaSeries } from '@/dataType/resource.dataType';
import { resourceServer } from '@/server/resource.server';
import { ElMessage } from 'element-plus';
import { ref, onMounted, onActivated } from "vue";
import { getResourceCoverPoster } from '@/common/photo';
import { appStoreData } from '@/storeData/app.storeData';
import { appLang } from '@/language/app.lang'
const store = {
  appStoreData: appStoreData(),
}
const props = defineProps({
  resourceId: {
    type: String,
    required: true,
  },
  dramaSeriesId: {
    type: String,
    default: '',
  },
})
const videoPlayRef = ref<InstanceType<typeof videoPlay>>();
const resourceInfo = ref<I_resource>();
const selectedDramaSeriesId = ref<string>('');
const loading = ref(false);

const init = async () => {
  await getResourceInfo();
  setVideoDramaSeries();
}

const getResourceInfo = async () => {
  loading.value = true;
  const result = await resourceServer.info(props.resourceId);
  if (!result || !result.status) {
    ElMessage.error(result.msg);
    return;
  }
  resourceInfo.value = result.data;
  loading.value = false;
};

const setVideoDramaSeries = () => {
  let dramaSeriesId = '';
  if (props.dramaSeriesId !== '') {
    dramaSeriesId = props.dramaSeriesId;
  } else if (resourceInfo.value && resourceInfo.value.dramaSeries.length > 0) {
    dramaSeriesId = resourceInfo.value.dramaSeries[0].id;
  }
  if (dramaSeriesId != '') {
    setVideoSource(dramaSeriesId);
  } else {
    noPlayList()
  }
}

const setVideoSource = (dramaSeriesId: string) => {
  selectedDramaSeriesId.value = dramaSeriesId;
  const vp = videoPlayRef.value;
  if (!vp) return;
  vp.setVideoSource('/api/video/mp4/' + dramaSeriesId, 'mp4', () => {
    vp.addTextTrack(
      `/api/video/subtitle/${dramaSeriesId}`,
      '默认字幕',
      'zh',
      true // 设为默认字幕
    )
    //vp.play();
  });
}

const noPlayList = () => {
  ElMessage({
    showClose: true,
    message: '播放列表为空',
    type: 'error',
  })
}

const playResourceDramaSeriesHandle = (ds: I_resourceDramaSeries) => {
  setVideoSource(ds.id)
}

onMounted(async () => {
  await init();
});
</script>
<style lang="scss" scoped>
.play-movies-container {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;

  .main-container {
    flex: 1;
    overflow: hidden;

    .main {
      width: calc(100% - 80px);
      height: calc(100% - 20px);
      padding: 10px 40px;
      display: flex;
      gap: 20px;
      overflow-y: auto;

      .main-left {
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 10px;

        .tagAlert {
          flex-shrink: 0;
        }

        .info-base {
          display: flex;
          flex-wrap: wrap;
          gap: 20px;
          font-size: 14px;
          color: #909399;
          line-height: 32px;
        }

        .performer-list {
          display: flex;
          flex-wrap: wrap;
          gap: 10px;

          .performer-item {
            width: 300px;
            border-radius: 10px;
            padding: 15px;
            background-color: #434343;
          }
        }

        .tag-list {
          display: flex;
          flex-wrap: wrap;
          gap: 10px;
        }

        .abstract {
          font-size: 16px;
          color: #909399;
          line-height: 1.5;
          text-indent: 2em;
        }
      }

      .main-right {
        width: 360px;
        flex-shrink: 0;
        display: flex;
        flex-direction: column;
        gap: 10px;

        .title {
          font-size: 14px;
        }
      }
    }
  }

  .c-height {
    padding-bottom: 50px;
  }

}
</style>
