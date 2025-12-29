<template>
  <dialogForm ref="dialogFormRef" title="演员集" width="400px" labelPosition="top" :modelValue="formData"
    :rules="formRules" @submit="submitHandle">
    <el-form-item label="演员集名称" prop="name">
      <el-input v-model="formData.name" />
    </el-form-item>
  </dialogForm>
</template>
<script lang="ts" setup>
import { LoadingService } from '@/assets/loading';
import dialogForm from '@/components/com/dialog/dialog-form.vue'
import { performerBasesServer } from '@/server/performerBases.server';
import { ElMessage, type FormRules } from 'element-plus';
import { reactive, ref } from 'vue';
import { performerBasesStoreData } from '@/storeData/performerBases.storeData';
const store = {
  performerBasesStoreData: performerBasesStoreData(),
}
const emits = defineEmits(['success'])

const dialogFormRef = ref<InstanceType<typeof dialogForm>>()

const formData = ref({
  name: '',
})
const formRules = reactive<FormRules>({
  name: [{ required: true, trigger: 'blur', message: '请输入演员集名称' }],
})
const submitHandle = async () => {
  LoadingService.show();
  try {
    const result = await performerBasesServer.create(formData.value.name);
    if (result.status) {
      store.performerBasesStoreData.add(result.data);
      emits('success');
      dialogFormRef.value?.close();
    } else {
      ElMessage.error(result.msg);
    }
  } catch {
    ElMessage.error('提交失败，请稍后再试');
  } finally {
    LoadingService.hide();
  }
}
const open = () => {
  dialogFormRef.value?.open()
}

defineExpose({ open })
</script>
