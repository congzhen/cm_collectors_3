<template>
  <dialogCommon ref="dialogCommonRef" title="设置批量资源标签" width="900px" top="10vh" @submit="submitHandle">
    <div class="resource-set-tag" v-loading="loading">
      <div class="resource-set-tag-container">
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
import { debounceNow } from '@/assets/debounce';
import { ElMessage } from 'element-plus';
import { resourceServer } from '@/server/resource.server';
const store = {
  appStoreData: appStoreData(),
}
const emits = defineEmits(['success'])

const dialogCommonRef = ref<InstanceType<typeof dialogCommon>>();

const loading = ref(false)
const tags = ref<Record<string, string[]>>({});
let resourceIds: string[] = [];


const init = (_resourceSlc: I_resource[]) => {
  tags.value = {};
  resourceIds = _resourceSlc.map(item => item.id);
}

const getTagIds = (): string[] => {
  //将tags中的id 转化为数组
  return Object.values(tags.value).flat();
}

const submitHandle = debounceNow(async () => {
  try {
    loading.value = true;
    if (resourceIds.length === 0) {
      ElMessage.error('请先选择资源');
      return;
    }
    const tagIds = getTagIds();
    if (tagIds.length === 0) {
      ElMessage.error('请选择标签');
      return;
    }
    const result = await resourceServer.batchAddTag(resourceIds, tagIds);
    if (result.status) {
      ElMessage.success('提交成功');
      emits('success');
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

const open = (_resourceSlc: I_resource[]) => {
  init(_resourceSlc);
  dialogCommonRef.value?.open();
}
const close = () => {
  dialogCommonRef.value?.close();
}
defineExpose({
  open,
})
</script>
