<template>
  <HeaderViewModern v-if="headerStyle === 'modern'" :mode="props.mode"
    @create-resouce-success="createResouceSuccessHandle" />
  <HeaderViewClassic v-else :mode="props.mode"
    @create-resouce-success="createResouceSuccessHandle" />
</template>

<script setup lang="ts">
import { computed, type PropType } from 'vue'
import { E_headerMode, type T_headerStyle } from '@/dataType/app.dataType'
import type { I_resource } from '@/dataType/resource.dataType'
import { appStoreData } from '@/storeData/app.storeData'
import HeaderViewClassic from './HeaderViewClassic.vue'
import HeaderViewModern from './HeaderViewModern.vue'

const props = defineProps({
  mode: {
    type: String as PropType<E_headerMode>,
    default: E_headerMode.Index,
  },
})

const emits = defineEmits<{
  createResouceSuccess: [resource: I_resource]
}>()

const store = appStoreData()
const headerStyle = computed<T_headerStyle>(() => store.appConfig.headerStyle || 'modern')

const createResouceSuccessHandle = (resource: I_resource) => {
  emits('createResouceSuccess', resource)
}
</script>
