<template>
  <el-dialog v-model="dialogVisible" class="modern-details-dialog" modal-class="modern-details-overlay"
    width="920px" top="4vh" :show-close="false" :append-to-body="true" :close-on-click-modal="false">
    <div v-if="props.resource" class="modern-details-shell">
      <button class="modern-details-close" type="button" aria-label="关闭" @click="close">
        <el-icon><Close /></el-icon>
      </button>

      <section class="modern-details-overview">
        <div class="modern-details-cover">
          <div class="modern-details-cover-backdrop" aria-hidden="true">
            <el-image :src="getResourceCoverPoster(props.resource)" fit="cover" />
          </div>
          <el-image class="modern-details-cover-image" :src="getResourceCoverPoster(props.resource)" fit="contain" />
        </div>

        <div class="modern-details-summary">
          <h2>{{ props.resource.title }}</h2>

          <detailsBtn class="modern-details-actions" :resource="props.resource" show-labels
            @paly="close" @update-resouce-success="updateResourceSuccessHandle"
            @delete-resource-success="deleteResourceSuccessHandle">
          </detailsBtn>

          <dl class="modern-details-meta">
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
            <div>
              <dt>收录时间</dt>
              <dd>{{ props.resource.addTime }}</dd>
            </div>
            <div>
              <dt>评分</dt>
              <dd>{{ props.resource.score }}</dd>
            </div>
            <div>
              <dt>评星</dt>
              <dd class="modern-details-rating">
                <el-rate :model-value="props.resource.stars || 0" disabled />
              </dd>
            </div>
          </dl>
        </div>
      </section>

      <section class="modern-details-content">
        <detailsInfo :resource="props.resource" :show-overview="false" modern></detailsInfo>
      </section>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, type PropType } from 'vue'
import type { I_resource } from '@/dataType/resource.dataType'
import { getResourceCoverPoster } from '@/common/photo'
import { AppLang } from '@/language/app.lang'
import detailsBtn from './detailsBtn.vue'
import detailsInfo from './detailsInfo.vue'

const appLang = AppLang()
const dialogVisible = ref(false)
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

const updateResourceSuccessHandle = (resource: I_resource) => {
  emits('updateResouceSuccess', resource)
}
const deleteResourceSuccessHandle = () => {
  emits('deleteResourceSuccess')
}
const open = () => {
  dialogVisible.value = true
}
const close = () => {
  dialogVisible.value = false
}

defineExpose({ open, close })
</script>

<style lang="scss">
.modern-details-overlay {
  background: rgba(2, 9, 15, 0.6);
  backdrop-filter: blur(3px);
}

.modern-details-dialog {
  --modern-details-bg: #0d1b28;
  --modern-details-soft-bg: rgba(255, 255, 255, 0.035);
  --modern-details-text: #edf5fb;
  --modern-details-text-muted: #9caebe;
  --modern-details-border: rgba(151, 176, 195, 0.14);

  max-width: calc(100vw - 32px);
  padding: 0;
  overflow: hidden;
  border: 1px solid var(--modern-details-border);
  border-radius: 10px;
  background: var(--modern-details-bg);
  box-shadow: 0 24px 70px rgba(0, 0, 0, 0.45);

  .el-dialog__header {
    display: none;
  }

  .el-dialog__body {
    padding: 0;
    color: var(--modern-details-text);
  }
}

.bright .modern-details-overlay {
  background: rgba(44, 57, 69, 0.24);
}

.bright .modern-details-dialog {
  --modern-details-bg: #ffffff;
  --modern-details-soft-bg: #f8fafc;
  --modern-details-text: #263442;
  --modern-details-text-muted: #667585;
  --modern-details-border: #e4e9ef;

  box-shadow: 0 18px 55px rgba(34, 49, 63, 0.2);
}

.modern-details-shell {
  position: relative;
  height: min(760px, 88vh);
  padding: 24px;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
}

.modern-details-close {
  position: absolute;
  z-index: 5;
  top: 12px;
  right: 12px;
  width: 30px;
  height: 30px;
  display: grid;
  place-items: center;
  border: 0;
  border-radius: 6px;
  color: var(--modern-details-text-muted);
  background: transparent;
  cursor: pointer;
  transition: color 0.16s ease, background-color 0.16s ease;

  &:hover {
    color: var(--modern-details-text);
    background: var(--modern-details-soft-bg);
  }
}

.modern-details-overview {
  min-height: 286px;
  display: grid;
  grid-template-columns: minmax(340px, 45%) minmax(0, 1fr);
  gap: 24px;
}

.modern-details-cover {
  position: relative;
  isolation: isolate;
  min-width: 0;
  height: 286px;
  overflow: hidden;
  border: 1px solid var(--modern-details-border);
  border-radius: 8px;
  background: #07111a;

  .modern-details-cover-image {
    position: relative;
    z-index: 1;
    width: 100%;
    height: 100%;
  }
}

.modern-details-cover-backdrop {
  position: absolute;
  z-index: 0;
  inset: -18px;
  opacity: 0.48;
  filter: blur(16px) saturate(0.82);
  transform: scale(1.08);

  .el-image {
    width: 100%;
    height: 100%;
  }

  &::after {
    position: absolute;
    inset: 0;
    background: rgba(3, 11, 18, 0.2);
    content: "";
  }
}

.bright .modern-details-cover {
  background: #f1f4f7;

  .modern-details-cover-backdrop {
    opacity: 0.34;

    &::after {
      background: rgba(255, 255, 255, 0.14);
    }
  }
}

.modern-details-summary {
  min-width: 0;
  padding: 4px 34px 0 0;

  h2 {
    margin: 0 0 16px;
    overflow: hidden;
    color: var(--modern-details-text);
    font-size: 22px;
    font-weight: 700;
    line-height: 1.3;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.modern-details-actions {
  margin-bottom: 16px;

  .el-button {
    --el-button-bg-color: var(--modern-details-soft-bg);
    --el-button-border-color: var(--modern-details-border);
    --el-button-text-color: var(--modern-details-text-muted);
    --el-button-hover-bg-color: var(--el-color-primary);
    --el-button-hover-border-color: var(--el-color-primary);
    --el-button-hover-text-color: #ffffff;
    height: 32px;
    padding: 0 12px;
  }

  .el-button:first-child {
    --el-button-bg-color: #22aeb3;
    --el-button-border-color: #22aeb3;
    --el-button-text-color: #ffffff;
  }

  .el-button:last-child {
    --el-button-hover-bg-color: var(--el-color-danger);
    --el-button-hover-border-color: var(--el-color-danger);
  }
}

.modern-details-meta {
  margin: 0;
  display: grid;
  gap: 8px;
  color: var(--modern-details-text-muted);
  font-size: 13px;

  > div {
    min-width: 0;
    display: grid;
    grid-template-columns: 76px minmax(0, 1fr);
    align-items: center;
  }

  dt,
  dd {
    min-width: 0;
    margin: 0;
  }

  dd {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.modern-details-rating {
  display: flex;
  align-items: center;
  gap: 8px;

  .el-rate {
    height: 18px;
  }
}

.modern-details-content {
  flex: 1;
  min-height: 0;
  margin-top: 18px;
  overflow: hidden;
  border-top: 1px solid var(--modern-details-border);
}

@media (max-width: 720px) {
  .modern-details-dialog {
    width: calc(100vw - 20px) !important;
    margin-top: 2vh;
  }

  .modern-details-shell {
    height: 94vh;
    padding: 16px;
  }

  .modern-details-overview {
    min-height: auto;
    grid-template-columns: 1fr;
    gap: 14px;
  }

  .modern-details-cover {
    height: 210px;
  }

  .modern-details-summary {
    padding-right: 0;

    h2 {
      padding-right: 34px;
    }
  }

  .modern-details-actions {
    .el-button-group {
      flex-wrap: wrap;
    }
  }

  .modern-details-meta {
    grid-template-columns: repeat(2, minmax(0, 1fr));

    > div {
      grid-template-columns: 60px minmax(0, 1fr);
    }
  }
}
</style>
