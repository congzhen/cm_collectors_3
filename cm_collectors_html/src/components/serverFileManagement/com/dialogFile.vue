<template>
  <el-dialog v-model="dialogVisible" :width="props.width" :append-to-body="true" :show-close="false"
    :close-on-click-modal="false">
    <el-form ref="ruleFormTagRef" :model="formData" :rules="formRules" label-position="top" status-icon
      v-loading="loading">
      <el-form-item :label="sfmLang('fileName')" prop="name">
        <el-input v-model="formData.name" spellcheck="false" />
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
import { sfm_CreateFile } from './request';
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
})
const formRules = reactive<FormRules>({
  name: [{ required: true, trigger: 'blur', message: sfmLang('enterFileName') }],
})

const submitHandle = (formEl: FormInstance | undefined) => {
  try {
    if (!formEl) return
    formEl.validate(async (valid) => {
      if (valid) {
        await CreateFile()
      }
    })
  } finally {
  }
}

const CreateFile = async () => {
  loading_open();
  const result = await sfm_CreateFile(formData.name, formData.path);
  if (result.status) {
    close();
    message(sfmLang('createFileSuccess'), 'success');
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
  formData.name = ''
  formData.path = _path
  dialogVisible.value = true
}
const close = () => {
  dialogVisible.value = false
}

defineExpose({ open })
</script>
