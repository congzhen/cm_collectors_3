<template>
  <drawerForm ref="drawerFormRef" :modelValue="formData" :rules="formRules" width="800px" :title="appLang.performer()"
    @submit="submitHandle">
    <el-alert v-if="sharedEditNotice" title="修改将同步到所有关联该演员的资源；演员资料单独保存。"
      type="info" :closable="false" style="margin-bottom: 16px" />
    <div class="performer-form-container">
      <div class="performer-form-left">
        <h3 class="performer-section-title">头像与状态</h3>
        <div class="performer-form-photo">
          <setImage ref="setImageRef" :src="getPerformerPhoto(formData)"></setImage>
        </div>
        <div class="performer-form-column performer-form-short-label">
          <el-form-item label="评星">
            <el-rate v-model="formData.stars" clearable />
          </el-form-item>
          <el-form-item label="息影">
            <el-checkbox v-model="formData.retreatStatus">{{ ' ' }}</el-checkbox>
          </el-form-item>
        </div>
      </div>
      <div class="performer-form-right">
        <div class="performer-form-column">
          <h3 class="performer-section-title">演员信息</h3>
          <el-form-item label="姓名" prop="name">
            <el-input v-model="formData.name" />
          </el-form-item>
          <el-form-item label="别名">
            <el-input v-model="formData.aliasName" type="textarea" :rows="2" />
          </el-form-item>
          <el-form-item label="简介">
            <el-input v-model="formData.introduction" type="textarea" :rows="4" />
          </el-form-item>
          <el-form-item label="国籍">
            <selectCountry v-model="formData.nationality" />
          </el-form-item>
          <el-form-item label="出生日期">
            <datePicker v-model="formData.birthday" />
          </el-form-item>
          <el-form-item label="职业" prop="career">
            <el-checkbox v-model="formData.careerPerformer" border>演员</el-checkbox>
            <el-checkbox v-model="formData.careerDirector" border>导演</el-checkbox>
          </el-form-item>
          <el-form-item v-if="store.appStoreData.currentConfigApp.plugInUnit_Cup"
            :label="store.appStoreData.currentCupText">
            <selectCup v-model="formData.cup" />
          </el-form-item>
          <el-form-item v-if="store.appStoreData.currentConfigApp.plugInUnit_Cup" label="三围">
            <inputCupBWH v-model:waist="formData.waist" v-model:bust="formData.bust" v-model:hip="formData.hip" />
          </el-form-item>
          <h3 class="performer-section-title performer-tag-title">演员标签</h3>
          <performerTagSelect v-model="formTagIds" :performer-bases-id="formPerformerBasesId_C" />

        </div>
      </div>
    </div>
    <template #footerBtn>
      <el-button @click="scraperPerformerInfoHandle">刮削演员信息</el-button>
    </template>
  </drawerForm>
  <scraperOnePerformerDialog ref="scraperOnePerformerDialogRef" @success="success" />
</template>
<script lang="ts" setup>
import { computed, reactive, ref } from 'vue'
import drawerForm from '../com/dialog/drawer-form.vue'
import selectCountry from '../com/form/selectCountry.vue'
import selectCup from '@/components/com/form/selectCup.vue'
import inputCupBWH from '@/components/com/form/inputCupBWH.vue'
import datePicker from '@/components/com/form/datePicker.vue'
import setImage from '@/components/com/form/setImage.vue'
import { ElMessage, type FormRules } from 'element-plus'
import type { I_performer } from '@/dataType/performer.dataType'
import { appStoreData } from '@/storeData/app.storeData'
import { LoadingService } from '@/assets/loading'
import { performerServer } from '@/server/performer.server'
import { getPerformerPhoto } from '@/common/photo';
import { AppLang } from '@/language/app.lang'
import scraperOnePerformerDialog from '../importResource/scraperOnePerformerDialog.vue'
import performerTagSelect from './performerTagSelect.vue'
const appLang = AppLang()

const store = {
  appStoreData: appStoreData(),
}
const props = defineProps({
  sharedEditNotice: { type: Boolean, default: false },
  performerBasesId: {
    type: String,
    default: '',
  },
})
const emits = defineEmits(['success'])

const drawerFormRef = ref<InstanceType<typeof drawerForm>>()
const setImageRef = ref<InstanceType<typeof setImage>>()
const scraperOnePerformerDialogRef = ref<InstanceType<typeof scraperOnePerformerDialog>>()

let mode: 'add' | 'edit' = 'add'
const initialFormData: I_performer = {
  id: '',
  performerBases_id: '',
  name: '',
  aliasName: '',
  keyWords: '',
  introduction: '',
  photo: '',
  stars: 0,
  retreatStatus: false,
  careerPerformer: true,
  careerDirector: false,
  nationality: '',
  cup: '',
  waist: '',
  hip: '',
  bust: '',
  birthday: '',
  status: true,
  resourceCount: 0,
  tags: [],
};

const formData = ref<I_performer>({ ...initialFormData })
const formTagIds = ref<string[]>([])
const formPerformerBasesId_C = computed(() => formData.value.performerBases_id || props.performerBasesId)

const formRules = reactive<FormRules>({
  name: [{ required: true, trigger: 'blur', message: '请输入姓名' }],
})



const submitHandle = async () => {
  if (formData.value.performerBases_id === '') {
    formData.value.performerBases_id = props.performerBasesId;
  }
  const photoBase64 = setImageRef.value?.getImageBase64() || '';
  LoadingService.show();
  try {
    const apiCall = mode === 'add'
      ? performerServer.create(formData.value, photoBase64, formTagIds.value)
      : performerServer.update(formData.value, photoBase64, formTagIds.value);
    const result = await apiCall;
    if (result.status) {
      success(result.data)
    } else {
      ElMessage.error(result.msg);
    }
  } catch (error) {
    ElMessage.error('提交失败，请稍后再试');
    console.log(error);
  } finally {
    LoadingService.hide();
  }
}

const scraperPerformerInfoHandle = () => {
  const performerBases_id = formData.value.performerBases_id === '' ? props.performerBasesId : formData.value.performerBases_id;
  scraperOnePerformerDialogRef.value?.open(formData.value.id, performerBases_id, formData.value.name);
}

const success = (performerInfo: I_performer) => {
  emits('success', mode === 'add' ? true : false, performerInfo);
  drawerFormRef.value?.close();
}

const open = (_mode: 'add' | 'edit', _performer: I_performer | null = null) => {
  setImageRef.value?.init();
  mode = _mode;
  if (_performer) {
    formData.value = { ..._performer, tags: _performer.tags ? [..._performer.tags] : [] };
    formTagIds.value = _performer.tags?.map(tag => tag.id) || [];
  } else {
    formData.value = { ...initialFormData };
    formTagIds.value = [];
  }
  drawerFormRef.value?.open()
}

// eslint-disable-next-line no-undef
defineExpose({ open, close: () => drawerFormRef.value?.close() })
</script>
<style lang="scss" scoped>
.performer-form-container {
  display: grid;
  grid-template-columns: 300px minmax(0, 1fr);
  align-items: start;
  gap: 16px;
  color: var(--el-text-color-primary);

  .performer-form-left,
  .performer-form-right {
    min-width: 0;
    padding: 12px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 8px;
    background: var(--el-bg-color);
  }
  .performer-section-title {
    margin: 0 0 12px;
    padding: 0 0 10px;
    border-bottom: 1px solid var(--el-border-color-lighter);
    font-size: 14px;
    line-height: 22px;
    font-weight: 600;
  }
  .performer-tag-title { margin-top: 4px; }

  // Keep the existing image viewport and crop dimensions.
  .performer-form-photo {
    width: 270px;
    height: 320px;
    padding: 1px;
    border: 1px solid var(--el-border-color);
    border-radius: 5px;
    background: var(--el-fill-color-light);
    transition: border-color .2s;
    &:hover { border-color: var(--el-color-primary); }
    :deep(.el-upload-dragger) { background: transparent; }
    :deep(.image-slot) { font-size: 13px; color: var(--el-text-color-secondary); }
  }
  .performer-form-short-label {
    padding-top: 16px;
    :deep(.el-form-item__label) { width: 54px !important; }
    :deep(.el-form-item:last-child) { margin-bottom: 0; }
  }
  :deep(.el-form-item) { margin-bottom: 16px; }
  :deep(.el-form-item__content) { min-width: 0; }
  .performer-form-right {
    :deep(.el-form-item__label) { width: 80px !important; }
    :deep(.el-date-editor) { width: 100%; }
    :deep(.el-checkbox.is-bordered) { margin-right: 12px; }
    :deep(.BWH) { width: 100%; }
    :deep(.BWH > div) { flex: 1; min-width: 0; }
    :deep(.performer-tag-empty) { padding: 4px 0 4px 80px; line-height: 1.7; }
  }
  .performer-form-column { display: flex; flex-direction: column; min-width: 0; }
}
</style>
