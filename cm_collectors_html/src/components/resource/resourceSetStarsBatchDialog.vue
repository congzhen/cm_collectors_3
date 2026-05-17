<template>
  <dialogCommon ref="dialogCommonRef" title="批量设置评星" width="420px" top="20vh" @submit="submitHandle">
    <div class="resource-set-stars" v-loading="loading">
      <el-form label-position="top">
        <el-form-item label="评星">
          <selectStarSet v-model="stars" />
        </el-form-item>
      </el-form>
    </div>
  </dialogCommon>
</template>
<script lang="ts" setup>
import dialogCommon from '@/components/com/dialog/dialog-common.vue';
import selectStarSet from '@/components/com/form/selectStarSet.vue';
import type { I_resource } from '@/dataType/resource.dataType';
import { debounceNow } from '@/assets/debounce';
import { resourceServer } from '@/server/resource.server';
import { ElMessage } from 'element-plus';
import { ref } from 'vue';

const emits = defineEmits(['success'])

const dialogCommonRef = ref<InstanceType<typeof dialogCommon>>();
const loading = ref(false);
const stars = ref(0);
let resourceIds: string[] = [];

const init = (_resourceSlc: I_resource[]) => {
  resourceIds = _resourceSlc.map(item => item.id);
  stars.value = 0;
}

const submitHandle = debounceNow(async () => {
  try {
    loading.value = true;
    if (resourceIds.length === 0) {
      ElMessage.error('请先选择资源');
      return;
    }
    const result = await resourceServer.batchSetStars(resourceIds, stars.value);
    if (result.status) {
      ElMessage.success('评星设置成功');
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
<style scoped lang="scss">
.resource-set-stars {
  padding: 12px 0;
}
</style>
