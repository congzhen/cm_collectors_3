<template>
  <div class="play-movies-mobile">
    <!-- 顶部导航栏 -->
    <MobileHeader :title="resourceInfo?.title || ''" :show-menu-button="true" @menu-action="handleMenuAction" />

    <!-- 视频播放器 -->
    <div class="video-container">
      <videoPlay ref="videoPlayRef" />
    </div>

    <!-- 剧集选择 -->
    <div class="episode-section" v-if="resourceInfo">
      <div class="section-title">剧集</div>
      <div class="episode-list">
        <resourceDramaSeriesList :drama-series="resourceInfo.dramaSeries" :selected-id="selectedDramaSeriesId"
          :show-mode="store.appStoreData.currentFilesBasesAppConfig.detailsDramaSeriesMode"
          @play-resource-drama-series="playResourceDramaSeriesHandle">
        </resourceDramaSeriesList>
      </div>
    </div>

    <!-- 基本信息 -->
    <div class="info-section" v-if="resourceInfo">
      <div class="section-title">基本信息</div>
      <div class="info-grid">
        <div class="info-item" v-if="resourceInfo.issueNumber">
          <span class="label">番号:</span>
          <span class="value">{{ resourceInfo.issueNumber }}</span>
        </div>
        <div class="info-item" v-if="resourceInfo.issuingDate && resourceInfo.issuingDate != ''">
          <span class="label">年份:</span>
          <span class="value">{{ resourceInfo.issuingDate }}</span>
        </div>
        <div class="info-item" v-if="resourceInfo.country != ''">
          <span class="label">国家:</span>
          <span class="value">{{ appLang.country(resourceInfo.country) }}</span>
        </div>
        <div class="info-item" v-if="resourceInfo.definition != ''">
          <span class="label">清晰度:</span>
          <span class="value">{{ appLang.definition(resourceInfo.definition) }}</span>
        </div>
        <div class="info-item">
          <span class="label">收录时间:</span>
          <span class="value">{{ resourceInfo.addTime }}</span>
        </div>
        <div class="info-item">
          <span class="label">评分:</span>
          <el-rate v-model="resourceInfo.stars" disabled size="small" />
        </div>
      </div>
    </div>

    <!-- 演员信息 -->
    <div class="performer-section" v-if="resourceInfo && resourceInfo.performers.length > 0">
      <div class="section-title">{{ appLang.performer() }}</div>
      <div class="performer-list">
        <div class="performer-item" v-for="performer in resourceInfo.performers" :key="performer.id"
          @click="goToPerformer(performer.id)">
          <performerPhoto class="el-avatar" :performer="performer"></performerPhoto>
          <div class="performer-name">{{ performer.name }}</div>
        </div>
      </div>
    </div>

    <!-- 剧照 -->
    <div class="tag-section" v-if="resourceInfo && store.appStoreData.currentFilesBasesAppConfig.sampleStatus">
      <div class="section-title">剧照</div>
      <div>
        <detailsSampleImages class="sample-list" :resource="resourceInfo" :columns="3"></detailsSampleImages>
      </div>
    </div>

    <!-- 标签 -->
    <div class="tag-section" v-if="resourceInfo && resourceInfo.tags.length > 0">
      <div class="section-title">标签</div>
      <div class="tag-list">
        <el-tag v-for="tag in resourceInfo.tags" :key="tag.id" type="info" effect="plain" size="small">
          {{ tag.name }}
        </el-tag>
      </div>
    </div>

    <!-- 摘要 -->
    <div class="abstract-section" v-if="resourceInfo && resourceInfo.abstract">
      <div class="section-title">摘要</div>
      <div class="abstract-content">{{ resourceInfo.abstract }}</div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, nextTick } from "vue";
import { useRouter } from 'vue-router';
import videoPlay from "@/components/play/videoPlay.vue";
import type { I_resource, I_resourceDramaSeries } from '@/dataType/resource.dataType';
import detailsSampleImages from '@/components/details/detailsSampleImages.vue'
import { resourceServer } from '@/server/resource.server';
import { ElMessage } from 'element-plus';
import performerPhoto from "@/components/performer/performerPhoto.vue";
import { AppLang } from '@/language/app.lang';
import { appStoreData } from "@/storeData/app.storeData";
import resourceDramaSeriesList from '@/components/resource/resourceDramaSeriesList.vue'
import MobileHeader from '../MobileHeaderView.vue'
import { getPlayVideoURL, playUpdate } from "@/common/play";

const appLang = AppLang();
const router = useRouter();

const store = {
  appStoreData: appStoreData(),
}
const videoPlayRef = ref<InstanceType<typeof videoPlay>>();
const resourceInfo = ref<I_resource>();
const selectedDramaSeriesId = ref<string>('');
const loading = ref(false);

const props = defineProps({
  resourceId: {
    type: String,
    required: true,
  },
  dramaSeriesId: {
    type: String,
    default: '',
  },
});

// 跳转到演员页面
const goToPerformer = (performerId: string) => {
  // 这里需要根据实际路由配置调整
  router.push(`/performer/${performerId}`);
};

const init = async () => {
  await getResourceInfo();
  setVideoDramaSeries();
};

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
  loading.value = true;
  let dramaSeriesId = '';
  if (props.dramaSeriesId !== '') {
    dramaSeriesId = props.dramaSeriesId;
  } else if (resourceInfo.value && resourceInfo.value.dramaSeries.length > 0) {
    dramaSeriesId = resourceInfo.value.dramaSeries[0].id;
  }
  if (dramaSeriesId != '') {
    setVideoSource(dramaSeriesId);
  } else {
    noPlayList();
  }
  loading.value = false;
};

const setVideoSource = (dramaSeriesId: string) => {
  selectedDramaSeriesId.value = dramaSeriesId;
  const vp = videoPlayRef.value;
  if (!vp) return;
  vp.setVideoSource(getPlayVideoURL(dramaSeriesId, 'mp4'), 'mp4', () => {
    vp.addTextTrack(
      `/api/video/subtitle/${dramaSeriesId}`,
      '默认字幕',
      'zh',
      true
    );
  });
};

const noPlayList = () => {
  ElMessage({
    showClose: true,
    message: '播放列表为空',
    type: 'error',
  });
};

const playResourceDramaSeriesHandle = (ds: I_resourceDramaSeries) => {
  setVideoSource(ds.id);
  playUpdate(ds.resources_id, ds.id)
};

// 处理菜单操作
const handleMenuAction = (action: string) => {
  switch (action) {
    case 'goBack':
      router.go(-1);
      break;
    case 'goHome':
      router.push('/');
      break;
  }
};

onMounted(async () => {
  nextTick(async () => {
    await init();
  });
});
</script>

<style lang="scss" scoped>
.play-movies-mobile {
  width: 100%;
  height: calc(100vh - 50px);
  background-color: #1f1f1f;
  color: #f3f3f3;
  padding: 0;
  padding-bottom: 20px;
  overflow-y: auto;

  .video-container {
    width: 100%;
    background-color: #000;
  }

  .section-title {
    font-size: 18px;
    font-weight: 500;
    padding: 15px 15px 10px;
    border-bottom: 1px solid #444;
  }

  .episode-section {
    .episode-list {
      display: flex;
      flex-wrap: wrap;
      padding: 10px 15px;

      .episode-item {
        padding: 8px 12px;
        margin: 5px;
        background-color: #333;
        border-radius: 4px;
        font-size: 14px;
        cursor: pointer;

        &.active {
          background-color: #409EFF;
          color: #fff;
        }
      }
    }
  }

  .info-section {
    .info-grid {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 12px;
      padding: 10px 15px;

      .info-item {
        display: flex;
        flex-direction: column;
        font-size: 14px;

        .label {
          color: #909399;
          margin-bottom: 3px;
        }

        .value {
          color: #f3f3f3;
        }
      }
    }
  }

  .performer-section {
    .performer-list {
      display: flex;
      flex-wrap: wrap;
      padding: 10px 15px;
      gap: 15px;

      .performer-item {
        display: flex;
        flex-direction: column;
        align-items: center;
        cursor: pointer;

        .el-avatar {
          width: 60px;
          height: 60px;
          margin-bottom: 5px;
        }

        .performer-name {
          font-size: 13px;
          text-align: center;
          max-width: 70px;
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
        }
      }
    }
  }

  .tag-section {
    .tag-list {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
      padding: 10px 15px;
    }
  }

  .abstract-section {
    .abstract-content {
      padding: 10px 15px;
      font-size: 14px;
      line-height: 1.5;
      color: #ccc;
    }
  }
}
</style>
