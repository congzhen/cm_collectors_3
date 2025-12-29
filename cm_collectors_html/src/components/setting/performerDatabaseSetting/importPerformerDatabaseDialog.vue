<template>
  <dialogForm ref="dialogFormRef" title="导入演员集" width="480px" :modelValue="formData" :rules="formRules"
    @submit="submitHandle">
    <el-form-item prop="performerDatabaseId" label="导入演员集">
      <selectDatabasePerformer v-model="formData.performerDatabaseId" />
    </el-form-item>
    <el-form-item prop="importFileName" label="导入文件">
      <el-input v-model="formData.importFileName" disabled>
        <template #append>
          <el-button icon="Upload" @click="importFileHandle"></el-button>
        </template>
      </el-input>
    </el-form-item>
    <el-form-item prop="performerDatabase">
      <el-checkbox v-model="formData.reconstructId" label="是否重建id" />
      <el-text class="warning-text" type="warning" size="small">
        导入数据，当遇到已存在的id时，将自动跳过该条数据，如果使用重建id，当检测到id冲突时，将自动生成一个新的唯一id来避免冲突
      </el-text>
    </el-form-item>
  </dialogForm>
</template>
<script lang="ts" setup>
import { loadJsonFile } from '@/assets/file';
import { LoadingService } from '@/assets/loading';
import dialogForm from '@/components/com/dialog/dialog-form.vue';
import selectDatabasePerformer from '@/components/com/form/selectDatabasePerformer.vue';
import { performerBasesServer } from '@/server/performerBases.server';
import { ElMessage, type FormRules } from 'element-plus';
import { reactive, ref } from 'vue';

const emits = defineEmits(['success'])
const dialogFormRef = ref<InstanceType<typeof dialogForm>>();
const formData = ref({
  performerDatabaseId: '',
  importFileName: '',
  importContent: '',
  reconstructId: false,
})
const formRules = reactive<FormRules>({
  performerDatabaseId: [{ required: true, trigger: 'blur', message: '请选择演员集' }],
  importFileName: [{ required: true, trigger: 'blur', message: '请选择导入文件' }],
})

const importFileHandle = async () => {
  const fileData = await loadJsonFile();
  if (fileData) {
    formData.value.importFileName = fileData.file.name;
    formData.value.importContent = fileData.content;
  }
}

const submitHandle = async () => {
  LoadingService.show();
  try {
    const result = await performerBasesServer.import(formData.value.performerDatabaseId, formData.value.importFileName, formData.value.importContent, formData.value.reconstructId);
    if (result.status) {
      dialogFormRef.value?.close();
      ElMessage.success(`共导入${result.data}条数据，请刷新后查看数据.`);
      emits('success');
    } else {
      ElMessage.error(result.msg);
    }
  } catch (error) {
    console.error(error);
    ElMessage.error('导入失败');
  } finally {
    LoadingService.hide();
  }
}

const open = () => {
  dialogFormRef.value?.open()
}

defineExpose({ open })
</script>
