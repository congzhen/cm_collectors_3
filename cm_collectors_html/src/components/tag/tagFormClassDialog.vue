<template>
  <dialogForm
    ref="dialogFormRef"
    title="标签分类"
    width="400px"
    labelPosition="top"
    :modelValue="formData"
    :rules="formRules"
  >
    <el-form-item label="标签分类名称" prop="name">
      <el-input v-model="formData.name" />
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
})
const formRules = reactive<FormRules>({
  name: [{ required: true, trigger: 'blur', message: '请输入标签分类名称' }],
})
const open = (mode: 'add' | 'edit', id: string, name: string) => {
  formData.id = id
  formData.name = name
  dialogFormRef.value?.open()
}

defineExpose({ open })
</script>
