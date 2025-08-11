<template>
  <drawerForm ref="drawerFormRef" :modelValue="formData" :rules="formRules" width="800px" :title="appLang.performer()"
    @submit="submitHandle">
    <div class="performer-form-container">
      <div class="performer-form-left">
        <div class="performer-form-photo">
          <setImage ref="setImageRef" :src="getPerformerPhoto(formData)"></setImage>
        </div>
        <div class="performer-form-column performer-form-short-label" style="padding-top: 1em">
          <el-form-item label="评分">
            <el-rate v-model="formData.stars" clearable />
          </el-form-item>
          <el-form-item label="息影">
            <el-checkbox v-model="formData.retreatStatus">{{ ' ' }}</el-checkbox>
          </el-form-item>
        </div>
      </div>
      <div class="performer-form-right">
        <div class="performer-form-column">
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

        </div>
      </div>
    </div>
  </drawerForm>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue'
import drawerForm from '../com/dialog/drawer.form.vue'
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
import { appLang } from '@/language/app.lang';
const store = {
  appStoreData: appStoreData(),
}
const props = defineProps({
  performerBasesId: {
    type: String,
    default: '',
  },
})
const emits = defineEmits(['success'])

const drawerFormRef = ref<InstanceType<typeof drawerForm>>()
const setImageRef = ref<InstanceType<typeof setImage>>()

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
};

const formData = ref<I_performer>({ ...initialFormData })

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
      ? performerServer.create(formData.value, photoBase64)
      : performerServer.update(formData.value, photoBase64);
    const result = await apiCall;
    if (result.status) {
      drawerFormRef.value?.close();
      emits('success', mode === 'add' ? true : false);
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

const open = (_mode: 'add' | 'edit', _performer: I_performer | null = null) => {
  setImageRef.value?.init();
  mode = _mode;
  if (_performer) {
    formData.value = _performer;
  } else {
    formData.value = { ...initialFormData };
  }
  drawerFormRef.value?.open()
}

// eslint-disable-next-line no-undef
defineExpose({ open })
</script>
<style lang="scss" scoped>
.performer-form-container {
  display: flex;

  .performer-form-left {
    flex-shrink: 0;
  }

  .performer-form-photo {
    width: 270px;
    height: 320px;
    padding: 1px;
    border-radius: 5px;
    border: 1px solid #4c4d4f;
    flex-shrink: 0;

    .performer-form-photo-text {
      width: 100%;
      height: 100%;
      display: flex;
      justify-content: center;
      align-items: center;
      font-size: 1.5em;
    }
  }

  .performer-form-right {
    flex-grow: 1;
  }

  .performer-form-short-label {
    :deep(.el-form-item__label) {
      width: 70px !important;
    }
  }

  .performer-form-column {
    display: flex;
    flex-direction: column;
    flex-grow: 1;
  }
}
</style>
