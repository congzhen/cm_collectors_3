<template>
  <div class="details-container" :class="{ 'details-container--modern': props.modern }" v-if="props.resource">
    <el-scrollbar>
      <template v-if="props.showOverview">
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
      </template>
      <div class="info-block">
        <div v-if="props.modern" class="modern-section-title">
          <span>资源（{{ props.resource.dramaSeries.length }}）</span>
        </div>
        <el-alert v-else class="tagAlert" title="资源" type="info" :closable="false" />
        <resourceDramaSeriesList class="resource" :drama-series="props.resource.dramaSeries"
          :show-mode="store.appStoreData.currentFilesBasesAppConfig.detailsDramaSeriesMode"
          :show-video-info="props.modern" :modern="props.modern"
          @play-resource-drama-series="playResourceDramaSeriesHandle">
        </resourceDramaSeriesList>
      </div>
      <div class="info-block" v-if="props.resource.directors.length > 0">
        <div v-if="props.modern" class="modern-section-title">
          <span>{{ appLang.director() }}</span>
        </div>
        <el-alert v-else class="tagAlert" :title="appLang.director()" type="success" :closable="false" />
        <div class="performer-list">
          <div class="performer-item" v-for="performer, key in props.resource.directors" :key="key">
            <performerPopoverBlock :performer="performer" :issuing-date="props.resource.issuingDate">
            </performerPopoverBlock>
          </div>
        </div>
      </div>
      <div class="info-block">
        <div v-if="props.modern" class="modern-section-title">
          <span>{{ appLang.performer() }}</span>
        </div>
        <el-alert v-else class="tagAlert" :title="appLang.performer()" type="success" :closable="false" />
        <div class="performer-list">
          <div class="performer-item" v-for="performer, key in props.resource.performers" :key="key">
            <performerPopoverBlock :performer="performer" :issuing-date="props.resource.issuingDate">
            </performerPopoverBlock>
          </div>
        </div>
      </div>
      <div class="info-block" v-if="store.appStoreData.currentFilesBasesAppConfig.sampleStatus">
        <div v-if="props.modern" class="modern-section-title">
          <span>剧照</span>
        </div>
        <el-alert v-else class="tagAlert" title="剧照" type="primary" :closable="false" />
        <div class="sample-list">
          <detailsSampleImages :resource="props.resource" :columns="props.modern ? 5 : 3"></detailsSampleImages>
        </div>
      </div>

      <div class="info-block">
        <div v-if="props.modern" class="modern-section-title">
          <span>标签（{{ props.resource.tags.length }}）</span>
        </div>
        <el-alert v-else class="tagAlert" title="标签" type="warning" :closable="false" />
        <detailsTags :resource="props.resource"></detailsTags>
      </div>
      <div class="info-block">
        <div v-if="props.modern" class="modern-section-title">
          <span>摘要</span>
        </div>
        <el-alert v-else class="tagAlert" title="摘要" type="info" :closable="false" />
        <div class="abstract" v-html="abstract_C"></div>
      </div>
    </el-scrollbar>
  </div>
</template>
<script lang="ts" setup>
import type { I_resource, I_resourceDramaSeries } from '@/dataType/resource.dataType';
import { ref, watch, computed, type PropType } from 'vue';
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
  showOverview: {
    type: Boolean,
    default: true,
  },
  modern: {
    type: Boolean,
    default: false,
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

const abstract_C = computed(() => {
  if (!props.resource) return ''
  //将props.resource.abstract中的换行符号转换为html的换行符号
  return props.resource.abstract.replace(/\n/g, '<br>')
})

const playResourceDramaSeriesHandle = (ds: I_resourceDramaSeries) => {
  if (!props.resource) return
  playResource(props.resource, ds.id)
}

</script>
<style lang="scss" scoped>
.details-container {
  width: 100%;
  min-width: 0;
  height: 100%;
  overflow: hidden;

  :deep(.el-scrollbar__view) {
    min-width: 0;
  }
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
  min-width: 0;
  padding: 5px 0;

  .el-alert {
    width: 100%;
    box-sizing: border-box;
    padding: 4px 8px;
  }

  .resource {
    width: 100%;
    min-width: 0;
    max-width: 100%;
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

.details-container--modern {
  color: var(--modern-details-text);

  .info-block {
    padding: 0;
    border-top: 1px solid var(--modern-details-border);
  }

  .modern-section-title {
    min-height: 42px;
    padding: 0 4px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    color: var(--modern-details-text);
    font-size: 13px;
    font-weight: 650;
  }

  .resource {
    padding-bottom: 7px;
  }

  .performer-list {
    padding: 0 4px 12px;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(82px, 110px));
    align-items: start;
    gap: 8px;

    .performer-item {
      width: auto;
      max-width: none;
    }
  }

  :deep(.performer-block) {
    height: 100%;
    padding: 5px;
    border: 1px solid var(--modern-details-border);
    border-radius: 7px;
    color: var(--modern-details-text-muted);
    background: var(--modern-details-soft-bg);
  }

  :deep(.performer-block-name) {
    padding: 7px 2px 3px;
    line-height: 16px;
  }

  :deep(.resourceDramaSeries-list-name li) {
    width: calc(100% - 16px);
    margin: 0 8px 6px;
    padding: 8px 10px;
    box-sizing: border-box;
    border: 1px solid var(--modern-details-border);
    border-radius: 5px;
    color: var(--modern-details-text-muted);
    background: var(--modern-details-soft-bg);
    font-style: normal;
  }

  :deep(.details-sample-images),
  :deep(.details-tag-list) {
    padding: 0 4px 10px;
    box-sizing: border-box;
  }

  .abstract {
    padding: 0 4px 12px;
    color: var(--modern-details-text-muted);
    line-height: 1.7;
    text-indent: 0;
  }
}
</style>
