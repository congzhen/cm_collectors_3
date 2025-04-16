<template>
  <drawerForm
    ref="drawerFormRef"
    :modelValue="formData"
    :rules="formRules"
    width="800px"
    title="演员"
  >
    <div class="performer-form-container">
      <div class="performer-form-left">
        <div class="performer-form-photo">
          <div class="performer-form-photo-text">
            <label>点击或拖拽上传照片</label>
          </div>
        </div>
        <div class="performer-form-column performer-form-short-label" style="padding-top: 1em">
          <el-form-item label="评分">
            <el-rate v-model="formData.stars" clearable />
          </el-form-item>
          <el-form-item label="息影">
            <el-checkbox v-model="formData.retreat_status">{{ ' ' }}</el-checkbox>
          </el-form-item>
        </div>
      </div>
      <div class="performer-form-right">
        <div class="performer-form-column">
          <el-form-item label="姓名" prop="name">
            <el-input v-model="formData.name" />
          </el-form-item>
          <el-form-item label="别名">
            <el-input v-model="formData.alias_name" type="textarea" :rows="2" />
          </el-form-item>
          <el-form-item label="简介">
            <el-input v-model="formData.introduction" type="textarea" :rows="4" />
          </el-form-item>
          <el-form-item label="国籍">
            <selectNationality v-model="formData.nationality" />
          </el-form-item>
          <el-form-item label="职业" prop="career">
            <el-checkbox-group v-model="formData.career">
              <el-checkbox :value="E_performerCareer.Performer" border>演员</el-checkbox>
              <el-checkbox :value="E_performerCareer.Director" border>导演</el-checkbox>
            </el-checkbox-group>
          </el-form-item>
          <el-form-item label="Cup">
            <selectCup v-model="formData.cup" />
          </el-form-item>
          <el-form-item label="三围">
            <inputCupBWH v-model="formData.cup_bwh" />
          </el-form-item>
          <el-form-item label="出生日期">
            <datePicker v-model="formData.birthday" />
          </el-form-item>
          <el-form-item label="所属数据集" prop="performer_bases_id">
            <selectDatabasePerformer v-model="formData.performer_bases_id" />
          </el-form-item>
        </div>
      </div>
    </div>
  </drawerForm>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue'
import drawerForm from '../com/dialog/drawer.form.vue'
import selectNationality from '@/components/com/form/selectNationality.vue'
import selectCup from '@/components/com/form/selectCup.vue'
import inputCupBWH from '@/components/com/form/inputCupBWH.vue'
import datePicker from '@/components/com/form/datePicker.vue'
import selectDatabasePerformer from '@/components/com/form/selectDatabasePerformer.vue'
import { E_performerCareer, type I_cupBWH } from '@/dataType/app.dataType'
import type { FormRules } from 'element-plus'
const drawerFormRef = ref<InstanceType<typeof drawerForm>>()
const formData = reactive({
  name: '',
  alias_name: '',
  introduction: '',
  stars: 0,
  retreat_status: false,
  career: [E_performerCareer.Performer],
  nationality: '',
  cup: '',
  cup_bwh: {
    waist: '',
    hip: '',
    bust: '',
  } as unknown as I_cupBWH,
  birthday: '',
  performer_bases_id: '',
})
const formRules = reactive<FormRules>({
  name: [{ required: true, trigger: 'blur', message: '请输入姓名' }],
  career: [{ type: 'array', required: true, trigger: 'change', message: '至少选择一项职业' }],
  performer_bases_id: [{ required: true, trigger: 'change', message: '请选择所属数据集' }],
})

const open = () => {
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
