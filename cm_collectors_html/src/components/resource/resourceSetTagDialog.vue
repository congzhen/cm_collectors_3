<template>
  <dialogCommon ref="dialogCommonRef" title="设置资源标签" width="900px" top="10vh" @submit="submitHandle">
    <div class="resource-set-tag" v-loading="loading">
      <div class="resource-set-tag-container" v-if="resource">
        <h3>{{ resource.title }}</h3>
        <div class="tag-container">
          <div class="tag-class" v-for="tagClass, key in store.appStoreData.currentTagClass.filter(item => item.status)"
            :key="key">
            <el-alert :title="tagClass.name" type="warning" :closable="false" />
            <div class="tag-group">
              <el-checkbox-group v-model="tags[tagClass.id]">
                <el-checkbox v-for="tag, index in store.appStoreData.currentTagsByTagClassId(tagClass.id)" :key="index"
                  :label="tag.name" :value="tag.id" />
              </el-checkbox-group>
            </div>
          </div>
        </div>
      </div>
    </div>
  </dialogCommon>
</template>
<script lang="ts" setup>
import dialogCommon from '@/components/com/dialog/dialog-common.vue';
import type { I_resource } from '@/dataType/resource.dataType';
import { ref } from 'vue'
import { appStoreData } from '@/storeData/app.storeData';
import { ElMessage } from 'element-plus';
import { debounceNow } from '@/assets/debounce';
import { resourceServer } from '@/server/resource.server';
const store = {
  appStoreData: appStoreData(),
}

const emits = defineEmits(['success'])

const dialogCommonRef = ref<InstanceType<typeof dialogCommon>>();

const loading = ref(false)
const resource = ref<I_resource>()
const tags = ref<Record<string, string[]>>({});

const init = (_resource: I_resource) => {
  resource.value = _resource;
  tags.value = {};
  store.appStoreData.currentTagClass.forEach(item => {
    tags.value[item.id] = _resource.tags.filter(tag => tag.tagClass_id == item.id).map(tag => tag.id);
  })
}

const submitHandle = debounceNow(async () => {
  try {
    loading.value = true;
    if (!resource.value) return;
    const _tags: string[] = [];
    for (const tagClassId in tags.value) {
      _tags.push(...tags.value[tagClassId]);
    }
    const result = await resourceServer.updateTag(resource.value.id, _tags);
    if (result.status) {
      ElMessage.success('提交成功');
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
.resource-set-tag-container {
  height: 60vh;
  display: flex;
  flex-direction: column;

  h3 {
    flex-shrink: 0;
    padding: 10px;
  }

  .tag-container {
    flex: 1;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 15px;

    .tag-class {
      padding: 0 10px;

      .tag-group {
        padding: 0 10px;
      }
    }
  }
}
</style>
