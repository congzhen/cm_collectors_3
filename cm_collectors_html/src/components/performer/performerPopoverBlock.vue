<template>
  <el-popover placement="left" trigger="click" :width="isModernAppearance ? 460 : 380"
    :popper-class="isModernAppearance ? 'modern-performer-popover' : ''">
    <template #reference>
      <performerBlock :performer="props.performer" :modern="isModernAppearance"></performerBlock>
    </template>
    <performerDetails :performer="props.performer" :issuing-date="props.issuingDate" :modern="isModernAppearance">
    </performerDetails>
  </el-popover>
</template>
<script lang="ts" setup>
import performerDetails from './performerDetails.vue'
import performerBlock from './performerBlock.vue'
import type { I_performer } from '@/dataType/performer.dataType';
import type { PropType } from 'vue';
import { computed } from 'vue';
import { appStoreData } from '@/storeData/app.storeData';
const store = appStoreData()
const props = defineProps({
  performer: {
    type: Object as PropType<I_performer>,
    required: true,
  },
  issuingDate: {
    type: String,
    default: ''
  }
})
const isModernAppearance = computed(() =>
  (store.appConfig.appearanceStyle || store.appConfig.headerStyle || 'modern') === 'modern'
)
</script>
<style lang="scss" scoped>
:global(.modern-performer-popover.el-popper) {
  --modern-performer-popover-bg: #1f1f1f;
  --modern-performer-popover-border: rgba(255, 255, 255, 0.12);

  padding: 0 !important;
  overflow: visible;
  border: 1px solid var(--modern-performer-popover-border) !important;
  border-radius: 9px !important;
  background: var(--modern-performer-popover-bg) !important;
  box-shadow: 0 18px 45px rgba(0, 0, 0, 0.38) !important;
}

:global(.modern-performer-popover.el-popper .el-popper__arrow::before) {
  border-color: var(--modern-performer-popover-border) !important;
  background: var(--modern-performer-popover-bg) !important;
}

:global(.bright .modern-performer-popover.el-popper) {
  --modern-performer-popover-bg: #ffffff;
  --modern-performer-popover-border: #dfe4e9;
  box-shadow: 0 16px 42px rgba(40, 54, 66, 0.18) !important;
}
</style>
