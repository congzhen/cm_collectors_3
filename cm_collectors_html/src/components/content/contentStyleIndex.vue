<template>
  <contentStyle1 v-if="props.resourcesShowMode == 'coverPoster'" :resource="props.resources"
    @click="selectResourcesHandle">
  </contentStyle1>
  <contentStyle2 v-else-if="props.resourcesShowMode == 'coverPosterBox'" :resource="props.resources"
    @click="selectResourcesHandle"></contentStyle2>
  <contentStyle3 v-else-if="props.resourcesShowMode == 'coverPosterBoxWideSeparate'" :resource="props.resources"
    @click="selectResourcesHandle"></contentStyle3>
  <contentStyleSimple v-else-if="props.resourcesShowMode == 'coverPosterSimple'" :resource="props.resources"
    @click="selectResourcesHandle"></contentStyleSimple>
  <contentStyleSimpleExpand v-else-if="props.resourcesShowMode == 'coverPosterSimpleExpand'" :resource="props.resources"
    @click="selectResourcesHandle"></contentStyleSimpleExpand>
  <div v-else>
    Resources Show Mode Error
  </div>
</template>
<script lang="ts" setup>
import contentStyle1 from '@/components/content/contentStyle1.vue';
import contentStyle2 from '@/components/content/contentStyle2.vue';
import contentStyle3 from '@/components/content/contentStyle3.vue';
import contentStyleSimple from './contentStyleSimple.vue';
import contentStyleSimpleExpand from './contentStyleSimpleExpand.vue';
import type { T_resourcesShowMode } from '@/dataType/app.dataType';
import type { I_resource } from '@/dataType/resource.dataType';
import type { PropType } from 'vue';

const props = defineProps({
  resources: {
    type: Object as PropType<I_resource>,
    required: true,
  },
  resourcesShowMode: {
    type: String as PropType<T_resourcesShowMode>,
    required: true,
  },
})
const emits = defineEmits(['selectResources']);
const selectResourcesHandle = () => {
  emits('selectResources', props.resources)
}
</script>
