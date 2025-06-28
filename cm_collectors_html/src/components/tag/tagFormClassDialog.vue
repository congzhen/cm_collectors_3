<template>
  <dialogForm ref="dialogFormRef" title="标签分类" width="400px" labelPosition="top" :modelValue="formData"
    :rules="formRules" @submit="submitHandle">
    <el-form-item label="标签分类名称" prop="name">
      <el-input v-model="formData.name" />
    </el-form-item>
  </dialogForm>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue'
import dialogForm from '../com/dialog/dialog.form.vue'
import { ElMessage, type FormRules } from 'element-plus'
import type { I_tagClass } from '@/dataType/tag.dataType';
import { tagServer } from '@/server/tag.server';
import { LoadingService } from '@/assets/loading'

const emits = defineEmits(['success'])

const dialogFormRef = ref<InstanceType<typeof dialogForm>>()
let mode = 'add';
const formData = ref<I_tagClass>({} as I_tagClass)
const formRules = reactive<FormRules>({
  name: [{ required: true, trigger: 'blur', message: '请输入标签分类名称' }],
})

const submitHandle = async () => {
  LoadingService.show();
  try {
    const apiCall = mode === 'add'
      ? tagServer.createTagClass(formData.value)
      : tagServer.updateTagClass(formData.value);
    const result = await apiCall;
    if (result.status) {
      emits('success');
      dialogFormRef.value?.close();
    } else {
      ElMessage.error(result.msg);
    }
  } catch (error) {
    ElMessage.error('提交失败，请稍后再试');
  } finally {
    LoadingService.hide();
  }
}

const open = (_mode: 'add' | 'edit', tagClass: I_tagClass) => {
  mode = _mode;
  formData.value = { ...tagClass };
  dialogFormRef.value?.open()
}

defineExpose({ open })
</script>
