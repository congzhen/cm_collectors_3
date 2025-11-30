<template>
  <div class="details-container" v-if="props.resource">
    <el-scrollbar>
      <div class="title">
        {{ props.resource.title }}
      </div>
      <div class="info-base">
        <div class="info-base-item" v-if="props.resource.issueNumber != ''">
          版号、番号、刊号: {{ props.resource.issueNumber }}
        </div>
        <div class="info-base-item">
          <el-breadcrumb separator="|">
            <el-breadcrumb-item v-if="props.resource.issuingDate && props.resource.issuingDate != ''">
              年份: {{ props.resource.issuingDate }}
            </el-breadcrumb-item>
            <el-breadcrumb-item v-if="props.resource.country != ''">
              国家: {{ appLang.country(props.resource.country) }}
            </el-breadcrumb-item>
            <el-breadcrumb-item v-if="props.resource.definition != ''">
              清晰度: {{ appLang.definition(props.resource.definition) }}
            </el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="info-base-item-flex">
          <div>收录时间: {{ props.resource.addTime }}</div>
          <div>评分: {{ props.resource.score }}</div>
        </div>
        <div class="info-base-rate">
          <el-rate v-model="localStars" disabled />
        </div>
      </div>
      <div class="info-block">
        <el-alert class="tagAlert" title="资源" type="info" :closable="false" />
        <resourceDramaSeriesList class="resource" :drama-series="props.resource.dramaSeries"
          :show-mode="store.appStoreData.currentFilesBasesAppConfig.detailsDramaSeriesMode"
          @play-resource-drama-series="playResourceDramaSeriesHandle">
        </resourceDramaSeriesList>
      </div>
      <div class="info-block" v-if="props.resource.directors.length > 0">
        <el-alert class="tagAlert" :title="appLang.director()" type="success" :closable="false" />
        <div class="performer-list">
          <div class="performer-item" v-for="performer, key in props.resource.directors" :key="key">
            <performerPopoverBlock :performer="performer" :issuing-date="props.resource.issuingDate">
            </performerPopoverBlock>
          </div>
        </div>
      </div>
      <div class="info-block">
        <el-alert class="tagAlert" :title="appLang.performer()" type="success" :closable="false" />
        <div class="performer-list">
          <div class="performer-item" v-for="performer, key in props.resource.performers" :key="key">
            <performerPopoverBlock :performer="performer" :issuing-date="props.resource.issuingDate">
            </performerPopoverBlock>
          </div>
        </div>
      </div>
      <div class="info-block" v-if="store.appStoreData.currentFilesBasesAppConfig.sampleStatus">
        <el-alert class="tagAlert" title="剧照" type="primary" :closable="false" />
        <div class="sample-list">
          <detailsSampleImages :resource="props.resource"></detailsSampleImages>
        </div>
      </div>

      <div class="info-block">
        <el-alert class="tagAlert" title="标签" type="warning" :closable="false" />
        <detailsTags :resource="props.resource"></detailsTags>
      </div>
      <div class="info-block">
        <el-alert class="tagAlert" title="摘要" type="info" :closable="false" />
        <div class="abstract">
          {{ props.resource.abstract }}
        </div>
      </div>
    </el-scrollbar>
  </div>
</template>
<script lang="ts" setup>
import type { I_resource, I_resourceDramaSeries } from '@/dataType/resource.dataType';
import { ref, watch, type PropType } from 'vue';
import { appStoreData } from '@/storeData/app.storeData';
import { playResource } from '@/common/play';
import resourceDramaSeriesList from '@/components/resource/resourceDramaSeriesList.vue'
import performerPopoverBlock from '@/components/performer/performerPopoverBlock.vue'
import detailsSampleImages from './detailsSampleImages.vue';
import detailsTags from './detailsTags.vue';
import { AppLang } from '@/language/app.lang'
const appLang = AppLang()

const store = {
  appStoreData: appStoreData(),
}
const props = defineProps({
  resource: {
    type: Object as PropType<I_resource> | undefined,
    default: undefined
  },
})
// 本地响应式变量，用于替代直接修改 props.resource.stars
const localStars = ref(props.resource?.stars || 0);

// 当 props.resource 变化时更新本地变量
watch(
  () => props.resource?.stars,
  (newVal) => {
    if (newVal !== undefined) {
      localStars.value = newVal;
    }
  },
  { immediate: true }
);

const playResourceDramaSeriesHandle = (ds: I_resourceDramaSeries) => {
  if (!props.resource) return
  playResource(props.resource, ds.id)
}

</script>
<style lang="scss" scoped>
.details-container {
  width: 100%;
  height: 100%;
  overflow: hidden;
}

.title {
  font-size: 1.1em;
  font-weight: 500;
  line-height: 1.2em;
  color: #ffffff;
}

.info-base {
  font-size: 12px;
  line-height: 1.5em;
  padding: 5px 0;


  :deep(.el-breadcrumb) {
    .el-breadcrumb__inner {
      font-size: 12px;
      color: #a8abb2;
    }
  }

  .info-base-item-flex {
    display: flex;
    gap: 10px;
  }

  .info-base-rate {
    .el-rate {
      height: 16px;
    }
  }
}

.info-block {
  padding: 5px 0;

  .el-alert {
    padding: 4px 8px;
  }

  .performer-list {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    padding-top: 4px;

    .performer-item {
      width: 32%;
      overflow: hidden;
    }
  }

  .sample-list {
    padding-top: 4px;
  }

  .tag-list {
    padding: 5px;
    display: flex;
    flex-wrap: wrap;
    gap: 5px;
  }

  .abstract {
    text-indent: 2em;
    padding: 10px;
  }
}
</style>
