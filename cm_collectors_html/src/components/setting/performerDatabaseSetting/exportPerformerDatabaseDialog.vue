<template>
  <dialogForm ref="dialogFormRef" title="导出演员集" width="400px" labelPosition="top" :modelValue="formData"
    :rules="formRules" @submit="submitHandle">
    <el-form-item prop="performerDatabase">
      <selectDatabasePerformer v-model="formData.performerDatabaseId" />
    </el-form-item>
  </dialogForm>
</template>
<script lang="ts" setup>
import { saveJsonFile } from '@/assets/file';
import { LoadingService } from '@/assets/loading';
import dialogForm from '@/components/com/dialog/dialog-form.vue';
import selectDatabasePerformer from '@/components/com/form/selectDatabasePerformer.vue';
import { performerBasesServer } from '@/server/performerBases.server';
import { ElMessage, type FormRules } from 'element-plus';
import { reactive, ref } from 'vue';
import { performerBasesStoreData } from '@/storeData/performerBases.storeData';
import { currentFormatDate } from '@/assets/timer';
const store = {
  performerBasesStoreData: performerBasesStoreData(),
}
const dialogFormRef = ref<InstanceType<typeof dialogForm>>();

const formData = ref({
  performerDatabaseId: '',
})
const formRules = reactive<FormRules>({
  performerDatabaseId: [{ required: true, trigger: 'blur', message: '请选择演员集' }],
})

const submitHandle = async () => {
  LoadingService.show();
  try {
    const result = await performerBasesServer.export(formData.value.performerDatabaseId);
    if (result.status) {
      dialogFormRef.value?.close();
      const fileName = store.performerBasesStoreData.getNameById(formData.value.performerDatabaseId);
      const timestamp = currentFormatDate('Ymd_His')
      saveJsonFile(result.data, `${fileName}-${timestamp}.json`);
    } else {
      ElMessage.error(result.msg);
    }
  } catch {
    ElMessage.error('导出失败');
  } finally {
    LoadingService.hide();
  }
}

const open = () => {
  dialogFormRef.value?.open()
}

defineExpose({ open })
</script>
