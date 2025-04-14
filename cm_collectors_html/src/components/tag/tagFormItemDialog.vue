<template>
  <dialogForm
    ref="dialogFormRef"
    title="标签"
    width="400px"
    labelPosition="top"
    :modelValue="formData"
    :rules="formRules"
  >
    <el-form-item label="标签名称" prop="name">
      <el-input v-model="formData.name" />
    </el-form-item>
    <el-form-item label="标签分类" prop="tag_class_id">
      <el-select v-model="formData.tag_class_id">
        <el-option label="韩国" value="4"></el-option>
        <el-option label="美国" value="3"></el-option>
        <el-option label="英国" value="2"></el-option>
      </el-select>
    </el-form-item>
  </dialogForm>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue'
import dialogForm from '../com/dialog/dialog.form.vue'
import type { FormRules } from 'element-plus'
const dialogFormRef = ref<InstanceType<typeof dialogForm>>()
const formData = reactive({
  id: '',
  name: '',
  tag_class_id: '',
})
const formRules = reactive<FormRules>({
  name: [{ required: true, trigger: 'blur', message: '请输入标签名称' }],
  tag_class_id: [{ required: true, trigger: 'blur', message: '请选择标签分类' }],
})
const open = (mode: 'add' | 'edit', id: string, name: string, tag_class_id: string) => {
  formData.id = id
  formData.name = name
  formData.tag_class_id = tag_class_id
  dialogFormRef.value?.open()
}

defineExpose({ open })
</script>
