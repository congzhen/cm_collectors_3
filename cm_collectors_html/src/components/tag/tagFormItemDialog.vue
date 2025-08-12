<template>
  <dialogForm ref="dialogFormRef" title="标签" width="400px" labelPosition="top" :modelValue="formData" :rules="formRules"
    @submit="submitHandle">
    <el-form-item label="标签名称" prop="name">
      <el-input v-model="formData.name" />
    </el-form-item>
    <el-form-item label="标签分类" prop="tagClass_id">
      <el-select v-model="formData.tagClass_id">
        <el-option v-for="item, key in store.appStoreData.currentTagClass" :key="key" :label="item.name"
          :value="item.id"></el-option>
      </el-select>
    </el-form-item>
  </dialogForm>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue'
import dialogForm from '../com/dialog/dialog-form.vue'
import { ElMessage, type FormRules } from 'element-plus'
import type { I_tag } from '@/dataType/tag.dataType'
import { appStoreData } from '@/storeData/app.storeData'
import { LoadingService } from '@/assets/loading'
import { tagServer } from '@/server/tag.server'
const store = {
  appStoreData: appStoreData(),
}
const emits = defineEmits(['success'])
const dialogFormRef = ref<InstanceType<typeof dialogForm>>()
let mode = 'add';
const formData = ref<I_tag>({} as I_tag)
const formRules = reactive<FormRules>({
  name: [{ required: true, trigger: 'blur', message: '请输入标签名称' }],
  tagClass_id: [{ required: true, trigger: 'blur', message: '请选择标签分类' }],
})
const submitHandle = async () => {
  LoadingService.show();
  try {
    const apiCall = mode === 'add'
      ? tagServer.createTag(formData.value)
      : tagServer.updateTag(formData.value);
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
const open = (_mode: 'add' | 'edit', tag: I_tag) => {
  mode = _mode;
  formData.value = { ...tag };
  dialogFormRef.value?.open()
}

defineExpose({ open })
</script>
