<template>
  <div class="play-movies-mobile" v-loading="loading">
    <!-- 顶部导航栏 -->
    <MobileHeader :title="resourceInfo?.title || ''" :show-menu-button="true" @menu-action="handleMenuAction" />

    <!-- 视频播放器 -->
    <div class="video-container">
      <mobileVideoPlayer v-if="resourceInfo && selectedDramaSeriesId" :resource-id="resourceId"
        :drama-series-id="selectedDramaSeriesId" :title="resourceInfo.title" :poster="getResourceCoverPoster(resourceInfo)" />
      <div v-if="resourceError" class="resource-error"><p>{{ resourceError }}</p><el-button @click="init">重试</el-button></div>
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
        <div class="info-item" v-if="resourceInfo.score > 0">
          <span class="label">评分:</span>
          <span class="value">{{ resourceInfo.score }}</span>
        </div>
        <div class="info-item">
          <span class="label">评星:</span>
          <el-rate v-model="resourceInfo.stars" disabled size="small" />
        </div>
      </div>
    </div>

    <!-- 导演信息 -->
    <div class="performer-section" v-if="resourceInfo && resourceInfo.directors.length > 0">
      <div class="section-title">{{ appLang.director() }}</div>
      <div class="performer-list">
        <div class="performer-item" v-for="performer in resourceInfo.directors" :key="performer.id">
          <performerPhoto class="el-avatar" :performer="performer"></performerPhoto>
          <div class="performer-name">{{ performer.name }}</div>
        </div>
      </div>
    </div>
    <!-- 演员信息 -->
    <div class="performer-section" v-if="resourceInfo && resourceInfo.performers.length > 0">
      <div class="section-title">{{ appLang.performer() }}</div>
      <div class="performer-list">
        <div class="performer-item" v-for="performer in resourceInfo.performers" :key="performer.id">
          <performerPhoto class="el-avatar" :performer="performer"></performerPhoto>
          <div class="performer-name">{{ performer.name }}</div>
        </div>
      </div>
    </div>

    <VideoDetailsExtras v-if="resourceInfo" :resource="resourceInfo" play-mode="page" />

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
      <div class="abstract-content" v-html="abstract_C"></div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import VideoDetailsExtras from "@/components/details/VideoDetailsExtras.vue";
import { ref, onMounted, onBeforeUnmount, nextTick, computed, defineAsyncComponent } from "vue";
import { useRouter } from 'vue-router';
const mobileVideoPlayer = defineAsyncComponent(() => import('@/components/play/mobileVideoPlayer.vue'));
import type { I_resource, I_resourceDramaSeries } from '@/dataType/resource.dataType';
import detailsSampleImages from '@/components/details/detailsSampleImages.vue'
import { resourceServer } from '@/server/resource.server';
import { ElMessage } from 'element-plus';
import performerPhoto from "@/components/performer/performerPhoto.vue";
import { AppLang } from '@/language/app.lang';
import { appStoreData } from "@/storeData/app.storeData";
import resourceDramaSeriesList from '@/components/resource/resourceDramaSeriesList.vue'
import MobileHeader from '../MobileHeaderView.vue'
import { playUpdate } from "@/common/play";
import { getResourceCoverPoster } from '@/common/photo';

const appLang = AppLang();
const router = useRouter();

const store = {
  appStoreData: appStoreData(),
}
const resourceError = ref('');
const resourceInfo = ref<I_resource>();
const selectedDramaSeriesId = ref<string>('');
const loading = ref(false);
let videoRequestVersion = 0

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
const abstract_C = computed(() => {
  if (!resourceInfo.value) return ''
  //将props.resource.abstract中的换行符号转换为html的换行符号
  return resourceInfo.value.abstract.replace(/\n/g, '<br>')
})

// 跳转到演员页面
const goToPerformer = (performerId: string) => {
  // 这里需要根据实际路由配置调整
  router.push(`/performer/${performerId}`);
};

const init = async () => {
  const loaded = await getResourceInfo();
  if (loaded) setVideoDramaSeries();
};

const getResourceInfo = async () => {
  const request = ++videoRequestVersion;
  loading.value = true;
  resourceError.value = '';
  try {
    const result = await resourceServer.info(props.resourceId);
    if (request !== videoRequestVersion) return false;
    if (!result || !result.status) {
      resourceError.value = result?.msg || '资源信息加载失败，请检查连接或资源是否仍存在';
      return false;
    }
    resourceInfo.value = result.data;
    return true
  } catch (error) {
    if (request !== videoRequestVersion) return false;
    console.error(error)
    resourceError.value = '资源信息加载失败，请检查网络后重试';
    return false
  } finally {
    if (request === videoRequestVersion) loading.value = false;
  }
};

const setVideoDramaSeries = () => {
  let dramaSeriesId = '';
  if (props.dramaSeriesId !== '' && resourceInfo.value?.dramaSeries.some(item => item.id === props.dramaSeriesId)) {
    dramaSeriesId = props.dramaSeriesId;
  } else if (resourceInfo.value && resourceInfo.value.dramaSeries.length > 0) {
    dramaSeriesId = resourceInfo.value.dramaSeries[0].id;
  }
  if (dramaSeriesId != '') {
    void setVideoSource(dramaSeriesId);
  } else {
    noPlayList();
  }
};

const setVideoSource = async (dramaSeriesId: string) => {
  selectedDramaSeriesId.value = dramaSeriesId;
  const path = `/play/moviesMobile/${encodeURIComponent(props.resourceId)}/${encodeURIComponent(dramaSeriesId)}`;
  if (router.currentRoute.value.path !== path) await router.replace(path);
};

const noPlayList = () => {
  ElMessage({
    showClose: true,
    message: '播放列表为空',
    type: 'error',
  });
};

const playResourceDramaSeriesHandle = (ds: I_resourceDramaSeries) => {
  void setVideoSource(ds.id);
  void playUpdate(ds.resources_id, ds.id)
};

// 处理菜单操作
const handleMenuAction = (action: string) => {
  switch (action) {
    case 'goBack':
      if (window.history.state?.back) router.go(-1);
      else router.replace('/mobile');
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

onBeforeUnmount(() => {
  videoRequestVersion++
})
</script>

<style lang="scss" scoped>
.play-movies-mobile {
  --mobile-play-bg: #121619;
  --mobile-play-surface: #1b2024;
  --mobile-play-surface-raised: #272d32;
  --mobile-play-border: rgba(255, 255, 255, 0.12);
  --mobile-play-text: #edf1f3;
  --mobile-play-muted: #9ba5ad;

  width: 100%;
  height: 100%;
  min-height: 0;
  box-sizing: border-box;
  background-color: var(--mobile-play-bg);
  color: var(--mobile-play-text);
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
    margin-bottom: 6px;
    border-bottom: 1px solid var(--mobile-play-border);
  }

  .episode-section {
    .episode-list {
      display: flex;
      flex-wrap: wrap;
      padding: 10px 15px;

      .episode-item {
        padding: 8px 12px;
        margin: 5px;
        background-color: var(--mobile-play-surface-raised);
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
          color: var(--mobile-play-muted);
          margin-bottom: 3px;
        }

        .value {
          color: var(--mobile-play-text);
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
      color: var(--mobile-play-muted);
    }
  }
}

:global(html.bright) .play-movies-mobile {
  --mobile-play-bg: #f5f7f8;
  --mobile-play-surface: #ffffff;
  --mobile-play-surface-raised: #f1f4f6;
  --mobile-play-border: #dce3e7;
  --mobile-play-text: #23323b;
  --mobile-play-muted: #71808a;
}
</style>
