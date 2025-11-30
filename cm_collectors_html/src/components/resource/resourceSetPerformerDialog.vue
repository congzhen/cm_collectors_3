<template>
  <dialogCommon ref="dialogCommonRef" title="设置资源演员" width="700px" top="10vh" @submit="submitHandle">
    <div class="resource-set-performer" v-loading="loading">
      <div class="resource-set-performer-container" v-if="resource">
        <h3>{{ resource.title }}</h3>
        <div class="performer-tool">
          <selectPerformer ref="selectPerformerRef" v-model="select_performers"
            :performerBasesIds="store.appStoreData.currentPerformerBasesIds"
            :careerType="E_performerCareerType.Performer" multiple @change="selectPerformerChangeHandle" />
          <el-button @click="createPerformerHandle">新增演员</el-button>
        </div>
        <div class="performer-container">
          <div class="performer-item" v-for="performer, key in performers" :key="key">
            <performerPopoverBlock :performer="performer" :issuing-date="resource.issuingDate">
            </performerPopoverBlock>
          </div>
        </div>
      </div>
    </div>
  </dialogCommon>
  <performerFormDrawer ref="performerFormDrawerRef" :performerBasesId="store.appStoreData.currentMainPerformerBasesId"
    @success="performerFormDrawerSuccess" />
</template>
<script lang="ts" setup>
import dialogCommon from '@/components/com/dialog/dialog-common.vue';
import type { I_resource } from '@/dataType/resource.dataType';
import { ref } from 'vue'
import { ElMessage } from 'element-plus';
import { debounce, debounceNow } from '@/assets/debounce';
import performerPopoverBlock from '@/components/performer/performerPopoverBlock.vue'
import type { I_performer } from '@/dataType/performer.dataType';
import selectPerformer from '@/components/com/form/selectPerformer.vue'
import { appStoreData } from '@/storeData/app.storeData';
import { E_performerCareerType } from '@/dataType/app.dataType';
import { performerServer } from '@/server/performer.server';
import { resourceServer } from '@/server/resource.server';
import performerFormDrawer from '@/components/performer/performerFormDrawer.vue';
const emits = defineEmits(['success'])

const store = {
  appStoreData: appStoreData(),
}

const dialogCommonRef = ref<InstanceType<typeof dialogCommon>>();
const selectPerformerRef = ref<InstanceType<typeof selectPerformer>>();
const performerFormDrawerRef = ref<InstanceType<typeof performerFormDrawer>>();

const loading = ref(false)
const resource = ref<I_resource>()
const performers = ref<I_performer[]>([]);
const select_performers = ref<string[]>([]);


const init = (_resource: I_resource) => {
  resource.value = _resource;
  performers.value = _resource.performers;
  select_performers.value = _resource.performers.map(p => p.id);
}

const selectPerformerChangeHandle = debounce(async () => {
  try {
    loading.value = true;
    const result = await performerServer.dataListByIds(select_performers.value);
    if (result.status) {
      performers.value = result.data;
    }
  } catch (error) {
    console.log(error);
  } finally {
    loading.value = false;
  }
})

const submitHandle = debounceNow(async () => {
  try {
    loading.value = true;
    if (!resource.value) return;
    const result = await resourceServer.updatePerformer(resource.value.id, select_performers.value);
    if (result.status) {
      ElMessage.success('修改演员成功');
      emits('success', result.data);
      close();
    } else {
      ElMessage.error(result.msg);
    }
  } catch (error) {
    ElMessage.error('提交失败，请稍后再试');
    console.log(error);
  } finally {
    loading.value = false;
  }
})

const createPerformerHandle = () => {
  performerFormDrawerRef.value?.open('add', null);
}
const performerFormDrawerSuccess = (mode: boolean, performer: I_performer) => {
  console.log(mode)
  performers.value.push(performer);
  select_performers.value.push(performer.id);
  selectPerformerRef.value?.resetOptionsData();
}

const open = (_resource: I_resource) => {
  init(_resource);
  dialogCommonRef.value?.open();
}
const close = () => {
  dialogCommonRef.value?.close();
}
defineExpose({
  open,
})
</script>
<style scoped lang="scss">
.resource-set-performer-container {
  height: 500px;
  display: flex;
  flex-direction: column;
  gap: 10px;

  h3 {
    flex-shrink: 0;
    padding: 5px;
  }

  .performer-container {
    flex: 1;
    overflow-y: auto;
    display: flex;
    flex-wrap: wrap;
    align-content: flex-start;
    gap: 15px;

    .performer-item {
      width: 120px;
      overflow: hidden;
    }
  }

  .performer-tool {
    flex-shrink: 0;
    display: flex;
    gap: 10px;
  }
}
</style>
