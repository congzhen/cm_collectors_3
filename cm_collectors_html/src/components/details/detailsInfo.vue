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
        <div class="info-base-item">收录时间: {{ props.resource.addTime }}</div>
        <div class="info-base-rate">
          <el-rate v-model="props.resource.stars" disabled />
        </div>
      </div>
      <div class="info-block">
        <el-alert class="tagAlert" title="资源" type="info" :closable="false" />
        <resourceDramaSeriesList class="resource" :drama-series="props.resource.dramaSeries"
          :show-mode="store.appStoreData.currentFilesBasesAppConfig.detailsDramaSeriesMode"
          @play-resource-drama-series="playResourceDramaSeriesHandle">
        </resourceDramaSeriesList>
      </div>
      <div class="info-block">
        <el-alert class="tagAlert" title="演员" type="success" :closable="false" />
        <div class="performer-list">
          <div class="performer-item" v-for="performer, key in props.resource.performers" :key="key">
            <performerPopoverBlock :performer="performer" :issuing-date="props.resource.issuingDate">
            </performerPopoverBlock>
          </div>
        </div>
      </div>
      <div class="info-block">
        <el-alert class="tagAlert" title="标签" type="warning" :closable="false" />
        <div class="tag-list">
          <el-tag type="info" effect="plain" size="large" v-for="item, key in props.resource.tags" :key="key">
            {{ item.name }}
          </el-tag>
        </div>
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
import type { PropType } from 'vue';
import { appLang } from '@/language/app.lang'
import { appStoreData } from '@/storeData/app.storeData';
import { playResource } from '@/common/play';
import resourceDramaSeriesList from '@/components/resource/resourceDramaSeriesList.vue'
import performerPopoverBlock from '@/components/performer/performerPopoverBlock.vue'

const store = {
  appStoreData: appStoreData(),
}
const props = defineProps({
  resource: {
    type: Object as PropType<I_resource> | undefined,
    default: undefined
  },
})

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
