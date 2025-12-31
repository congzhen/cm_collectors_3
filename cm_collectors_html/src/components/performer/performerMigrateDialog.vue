<template>
  <dialogForm ref="dialogFormRef" :title="title_C" width="400px" :modelValue="formData" :rules="formRules"
    @submit="submitHandle">
    <el-form-item label="迁移至" prop="performerBases_id">
      <selectDatabasePerformer v-model="formData.performerBases_id" />
    </el-form-item>
  </dialogForm>
</template>
<script lang="ts" setup>
import type { I_performer } from '@/dataType/performer.dataType';
import dialogForm from '../com/dialog/dialog-form.vue';
import { computed, reactive, ref } from 'vue';
import selectDatabasePerformer from '../com/form/selectDatabasePerformer.vue';
import { ElMessage, type FormRules } from 'element-plus';
import { performerServer } from '@/server/performer.server';
import { LoadingService } from '@/assets/loading';

const emits = defineEmits(['success']);

const dialogFormRef = ref<InstanceType<typeof dialogForm>>();

const performer = ref<I_performer>();
const formData = ref({
  performerBases_id: '',
});
const formRules = reactive<FormRules>({
  performerBases_id: [{ required: true, trigger: 'blur', message: '请选择演员数据集' }],
})
const title_C = computed(() => {
  return `迁移演员 - ${performer.value?.name}`;
});

const init = (_perforemr: I_performer) => {
  performer.value = _perforemr;
}

const submitHandle = async () => {
  if (!performer.value) return;
  if (formData.value.performerBases_id === performer.value.performerBases_id) {
    ElMessage.error('请选择不同的数据集进行迁移');
    return;
  }
  try {
    LoadingService.show();
    const result = await performerServer.migrate(performer.value.id, formData.value.performerBases_id);
    if (result.status) {
      emits('success');
      close();
    } else {
      ElMessage.error(result.msg);
    }
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
  } catch (error) {
    ElMessage.error('迁移失败');
  } finally {
    LoadingService.hide();
  }
}

const open = (_perforemr: I_performer) => {
  init(_perforemr);
  dialogFormRef.value?.open();
}
const close = () => {
  dialogFormRef.value?.close();
}

defineExpose({
  open,
});

</script>
