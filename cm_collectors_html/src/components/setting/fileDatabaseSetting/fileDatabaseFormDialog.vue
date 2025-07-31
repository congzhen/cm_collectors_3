<template>
  <dialogForm ref="dialogFormRef" title="文件库" width="400px" labelPosition="top" :modelValue="formData"
    :rules="formRules" @submit="submitHandle">
    <el-form-item label="文件库名称" prop="name">
      <el-input v-model="formData.name" />
    </el-form-item>
    <el-form-item label="(主)演员集">
      <el-select v-model="formData.mainPerformerBasesId" @change="mainPerformerBasesChange">
        <el-option v-for="item, index in store.performerBasesStoreData.performerBases" :key="index" :label="item.name"
          :value="item.id"></el-option>
      </el-select>
    </el-form-item>
    <el-form-item label="关联演员集">
      <el-checkbox-group v-model="formData.relatedPerformerBasesIds">
        <el-checkbox v-for="item, key in store.performerBasesStoreData.performerBases" :key="key" :label="item.name"
          :value="item.id" :disabled="item.id == formData.mainPerformerBasesId" />
      </el-checkbox-group>
    </el-form-item>
  </dialogForm>
</template>
<script lang="ts" setup>
import { LoadingService } from '@/assets/loading';
import dialogForm from '@/components/com/dialog/dialog.form.vue'
import { filesBasesServer } from '@/server/filesBases.server';
import { ElMessage, type FormRules } from 'element-plus';
import { reactive, ref } from 'vue';
import { filesBasesStoreData } from '@/storeData/filesBases.storeData';
import { performerBasesStoreData } from '@/storeData/performerBases.storeData';
const store = {
  filesBasesStoreData: filesBasesStoreData(),
  performerBasesStoreData: performerBasesStoreData(),
}

const emits = defineEmits(['success'])

const dialogFormRef = ref<InstanceType<typeof dialogForm>>()

const formData = ref({
  name: '',
  mainPerformerBasesId: store.performerBasesStoreData.activeFirstPerformerBasesId,
  relatedPerformerBasesIds: [store.performerBasesStoreData.activeFirstPerformerBasesId],
})
const formRules = reactive<FormRules>({
  name: [{ required: true, trigger: 'blur', message: '请输入文件库名称' }],
  mainPerformerBasesId: [{ required: true, trigger: 'blur', message: '请选择(主)演员集' }],
})
const submitHandle = async () => {
  LoadingService.show();
  try {
    const result = await filesBasesServer.create(formData.value.name, formData.value.mainPerformerBasesId, formData.value.relatedPerformerBasesIds);
    if (result.status) {
      store.filesBasesStoreData.add(result.data);
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

const mainPerformerBasesChange = () => {
  if (!formData.value.relatedPerformerBasesIds.find(item => item === formData.value.mainPerformerBasesId)) {
    formData.value.relatedPerformerBasesIds.push(formData.value.mainPerformerBasesId)
  }
}

const open = () => {
  dialogFormRef.value?.open()
}

defineExpose({ open })
</script>
