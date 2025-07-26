<template>
  <el-dialog v-model="dialogVisible" :width="props.width" :append-to-body="true" :show-close="false"
    :close-on-click-modal="false">
    <el-form ref="ruleFormTagRef" :model="formData" :rules="formRules" label-position="top" status-icon
      v-loading="loading">
      <el-form-item :label="sfmLang('folderName')" prop="name">
        <el-input v-model="formData.name" spellcheck="false" />
      </el-form-item>
      <el-form-item>
        <div class="permissions-container">
          <div>{{ sfmLang('owner') }}</div>
          <el-checkbox-group v-model="formData.ownerPerm">
            <el-checkbox :label="sfmLang('read')" :value="4" />
            <el-checkbox :label="sfmLang('write')" :value="2" />
            <el-checkbox :label="sfmLang('execute')" :value="1" />
          </el-checkbox-group>

          <div>{{ sfmLang('group') }}</div>
          <el-checkbox-group v-model="formData.groupPerm">
            <el-checkbox :label="sfmLang('read')" :value="4" />
            <el-checkbox :label="sfmLang('write')" :value="2" />
            <el-checkbox :label="sfmLang('execute')" :value="1" />
          </el-checkbox-group>

          <div>{{ sfmLang('public') }}</div>
          <el-checkbox-group v-model="formData.publicPerm">
            <el-checkbox :label="sfmLang('read')" :value="4" />
            <el-checkbox :label="sfmLang('write')" :value="2" />
            <el-checkbox :label="sfmLang('execute')" :value="1" />
          </el-checkbox-group>
        </div>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogVisible = false" :disabled="loading"> {{ sfmLang('close') }} </el-button>
        <el-button type="primary" @click="submitHandle(ruleFormTagRef)" :disabled="loading"> {{ sfmLang('submit') }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>
<script lang="ts" setup>
import { type FormInstance, type FormRules } from 'element-plus';
import { reactive, ref, type PropType } from 'vue';
import { sfm_CreateFolder } from './request';
import { message, messageBoxAlert } from './fn';
import { sfm_languages } from './lang';
import type { E_LangType } from './dataType';
const sfmLang = (key: string) => (sfm_languages[props.lang] as Record<string, string>)[key];

const dialogVisible = ref(false)
const loading = ref(false)
const loading_open = () => {
  loading.value = true
}
const loading_close = () => {
  loading.value = false
}

const props = defineProps({
  width: {
    type: String,
    default: '400px',
  },
  lang: {
    type: String as PropType<E_LangType>,
    required: true,
  },
})

const emit = defineEmits(['success'])

const ruleFormTagRef = ref<FormInstance>()

const formData = reactive({
  name: '',
  path: '',
  ownerPerm: [] as number[],
  groupPerm: [] as number[],
  publicPerm: [] as number[],
})
const formRules = reactive<FormRules>({
  name: [{ required: true, trigger: 'blur', message: sfmLang('enterFolderName') }],
})
// 权限计算函数
const calculatePermissions = (): string => {
  const sum = (arr: number[]) => arr.reduce((a, b) => a + b, 0)

  return [
    sum(formData.ownerPerm),
    sum(formData.groupPerm),
    sum(formData.publicPerm)
  ].join('').padEnd(3, '0').padStart(4, '0')
}

const submitHandle = (formEl: FormInstance | undefined) => {
  try {
    if (!formEl) return
    formEl.validate(async (valid) => {
      if (valid) {
        await CreateFolder()
      }
    })
  } finally {
  }
}

const CreateFolder = async () => {
  const permissions = calculatePermissions()
  loading_open();
  const result = await sfm_CreateFolder(formData.name, formData.path, permissions);
  if (result.status) {
    close();
    message(sfmLang('createFolderSuccess'), 'success');
    emit('success');
  } else {
    messageBoxAlert({
      text: result.msg,
      type: 'error',
      ok: sfmLang('confirm'),
    })
  }
  loading_close();
}

const open = (_path: string) => {
  formData.name = '';
  formData.path = _path;
  formData.ownerPerm = [4, 2, 1];
  formData.groupPerm = [4, 1];
  formData.publicPerm = [4, 1];
  dialogVisible.value = true;
}
const close = () => {
  dialogVisible.value = false
}

defineExpose({ open })
</script>
<style lang="scss" scoped>
.permissions-container {
  display: flex;
  flex-direction: column;
  padding-left: 5px;

  .el-checkbox-group {
    padding-left: 60px;
  }
}
</style>
