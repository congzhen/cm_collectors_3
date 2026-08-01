<template>
  <aside class="modern-right-details" :class="{ 'modern-right-details--bright': isBrightTheme }">
    <el-scrollbar v-if="props.resource" class="modern-right-details-scrollbar">
      <div class="modern-right-details-shell">
        <header class="modern-right-details-header">
          <span>DETAILS</span>
          <strong>资源详情</strong>
        </header>

        <section class="modern-right-details-overview">
          <div class="modern-right-details-cover">
            <el-image :src="getResourceCoverPoster(props.resource)" fit="contain" />
          </div>

          <detailsBtn class="modern-right-details-actions" :resource="props.resource" show-labels single-row compact-labels
            @update-resouce-success="updateResourceSuccessHandle"
            @delete-resource-success="deleteResourceSuccessHandle">
          </detailsBtn>

          <h2>{{ props.resource.title }}</h2>

          <dl class="modern-right-details-meta">
            <div v-if="props.resource.issueNumber">
              <dt>版号</dt>
              <dd>{{ props.resource.issueNumber }}</dd>
            </div>
            <div v-if="props.resource.issuingDate">
              <dt>年份</dt>
              <dd>{{ props.resource.issuingDate }}</dd>
            </div>
            <div v-if="props.resource.country">
              <dt>国家</dt>
              <dd>{{ appLang.country(props.resource.country) }}</dd>
            </div>
            <div v-if="props.resource.definition">
              <dt>清晰度</dt>
              <dd>{{ appLang.definition(props.resource.definition) }}</dd>
            </div>
            <div class="modern-right-details-meta--wide">
              <dt>收录时间</dt>
              <dd>{{ props.resource.addTime }}</dd>
            </div>
            <div>
              <dt>评分</dt>
              <dd>{{ props.resource.score }}</dd>
            </div>
          </dl>

          <div class="modern-right-details-stars">
            <span>评星</span>
            <el-rate :model-value="props.resource.stars || 0" disabled />
          </div>
        </section>

        <detailsInfo class="modern-right-details-content" :resource="props.resource" :show-overview="false"
          modern :modern-performers="false" :sample-columns="4">
        </detailsInfo>
      </div>
    </el-scrollbar>
  </aside>
</template>

<script setup lang="ts">
import { computed, type PropType } from 'vue'
import type { I_resource } from '@/dataType/resource.dataType'
import { getResourceCoverPoster } from '@/common/photo'
import { AppLang } from '@/language/app.lang'
import { appStoreData } from '@/storeData/app.storeData'
import detailsBtn from './detailsBtn.vue'
import detailsInfo from './detailsInfo.vue'

const appLang = AppLang()
const store = appStoreData()
const props = defineProps({
  resource: {
    type: Object as PropType<I_resource> | undefined,
    default: undefined,
  },
})
const emits = defineEmits<{
  updateResouceSuccess: [resource: I_resource]
  deleteResourceSuccess: []
}>()

const isBrightTheme = computed(() => store.appConfig.theme === 'bright')
const updateResourceSuccessHandle = (resource: I_resource) => emits('updateResouceSuccess', resource)
const deleteResourceSuccessHandle = () => emits('deleteResourceSuccess')
</script>

<style scoped lang="scss">
.modern-right-details {
  --modern-details-bg: var(--home-panel-bg, #1f1f1f);
  --modern-details-soft-bg: var(--home-panel-soft-bg, #27292d);
  --modern-details-text: var(--home-text, #e4e7ed);
  --modern-details-text-muted: var(--home-text-muted, #a8abb2);
  --modern-details-border: var(--home-border, rgba(255, 255, 255, 0.1));
  --modern-details-accent: #25b5b3;

  width: clamp(320px, 22vw, 360px);
  height: 100%;
  padding-left: 6px;
  box-sizing: border-box;
  color: var(--modern-details-text);

  &--bright {
    --modern-details-bg: var(--home-panel-bg, #ffffff);
    --modern-details-soft-bg: var(--home-panel-soft-bg, #f7f8fa);
    --modern-details-text: var(--home-text, #303133);
    --modern-details-text-muted: var(--home-text-muted, #737d87);
    --modern-details-border: var(--home-border, #e2e6eb);
    --modern-details-accent: #159fa1;
  }
}

.modern-right-details-scrollbar {
  height: 100%;
  background: var(--modern-details-bg);
  border: 1px solid var(--modern-details-border);
  border-radius: 8px;
  box-sizing: border-box;
}

.modern-right-details-shell {
  min-width: 0;
  padding: 0 10px 16px;
  box-sizing: border-box;
}

.modern-right-details-header {
  height: 42px;
  display: flex;
  align-items: center;
  gap: 7px;
  color: var(--modern-details-text-muted);
  border-bottom: 1px solid var(--modern-details-border);
  font-size: 11px;
  letter-spacing: 0.06em;

  strong {
    color: var(--modern-details-text);
    font-size: 12px;
    font-weight: 600;
    letter-spacing: 0;
  }
}

.modern-right-details-overview {
  padding-top: 10px;

  h2 {
    margin: 12px 2px 9px;
    color: var(--modern-details-text);
    font-size: 18px;
    line-height: 1.35;
    overflow-wrap: anywhere;
  }
}

.modern-right-details-cover {
  width: 100%;
  overflow: hidden;
  border-radius: 7px;
  background: transparent;

  :deep(.el-image) {
    display: block;
    width: 100%;
  }

  :deep(.el-image__inner) {
    display: block;
    width: 100%;
    height: auto;
  }
}

.modern-right-details-actions {
  margin-top: 7px;

  :deep(.el-button-group) {
    display: flex !important;
    flex-wrap: nowrap !important;
    gap: 4px;
  }

  :deep(.el-button) {
    min-width: 0;
    height: 32px;
    margin: 0;
    padding: 5px 4px;
    color: var(--modern-details-text-muted);
    background: var(--modern-details-soft-bg);
    border-color: var(--modern-details-border);
    border-radius: 5px !important;
    font-size: 11px;

    &:first-child {
      color: #ffffff;
      background: var(--modern-details-accent);
      border-color: var(--modern-details-accent);
    }

    &:last-child {
      color: var(--el-color-danger);
    }
  }
}

.modern-right-details-meta {
  margin: 0;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 7px 12px;
  color: var(--modern-details-text-muted);
  font-size: 11px;

  div {
    min-width: 0;
    display: grid;
    grid-template-columns: auto minmax(0, 1fr);
    gap: 7px;
  }

  dt,
  dd {
    min-width: 0;
    margin: 0;
  }

  dt {
    white-space: nowrap;
  }

  dd {
    color: var(--modern-details-text);
    overflow-wrap: anywhere;
  }

  .modern-right-details-meta--wide {
    grid-column: 1 / -1;
  }
}

.modern-right-details-stars {
  min-height: 30px;
  margin-top: 10px;
  padding: 0 2px;
  display: flex;
  align-items: center;
  gap: 12px;
  color: var(--modern-details-text-muted);
  background: transparent;
  border: 0;
  font-size: 12px;

  :deep(.el-rate) {
    height: auto;
  }
}

.modern-right-details-content {
  margin-top: 10px;
  border-bottom: 1px solid var(--modern-details-border);
}

.modern-right-details-content :deep(.details-sample-images .images-container) {
  gap: 5px;
}

@media (max-width: 1100px) {
  .modern-right-details {
    width: 310px;
  }

  .modern-right-details-actions :deep(.el-button span) {
    font-size: 10px;
  }
}
</style>
