<template>
  <drawerForm ref="drawerFormRef" :modelValue="formData" :rules="formRules" width="1024px" title="资源"
    @submit="submitHandle">
    <div class="resource-form-container">
      <div class="resource-form-left">
        <div class="resource-form-class-title">
          <el-alert title="封面" type="info" :closable="false" />
        </div>
        <div ref="resourceFormCoverPosterContainerRef" class="resource-form-cover-poster"
          :style="{ width: '100%', height: coverPosterHeight_C }">
          <setImage ref="setImageRef" :src="getResourceCoverPoster(formData)"
            :cropperWidth="store.appStoreData.currentConfigApp.coverPosterData[formData.coverPosterMode].width"
            :cropperHeight="store.appStoreData.currentConfigApp.coverPosterData[formData.coverPosterMode].height">
          </setImage>
        </div>
        <div class="resource-form-mode">
          <el-radio-group v-model="formData.coverPosterMode">
            <el-radio v-for="item, index in store.appStoreData.currentConfigApp.coverPosterData" :key="index"
              :value="index" border>
              {{ item.name }}
            </el-radio>
          </el-radio-group>
        </div>
        <div>
          <el-alert title="资源" type="info" :closable="false" />
          <ul class="resource-form-drama-series">
            <li v-for="item, index in dramaSeries" :key="index">
              <label class="drama-series-index">{{ (index + 1) }}.</label>
              <el-input v-model="item.src" size="small" />
              <el-button-group class="drama-series-tool" size="small">
                <el-button icon="Delete" @click="dramaSeriesDeleteHandle(index)"></el-button>
              </el-button-group>
            </li>
          </ul>
          <div class="resource-form-browse">
            <el-button icon="MostlyCloudy" type="primary" plain @click="selectServerFilesHandle">选择资源</el-button>
          </div>
        </div>
      </div>
      <div class="resource-form-right">
        <div class="resource-form-class-title">
          <el-alert title="信息" type="success" :closable="false" />
        </div>
        <div class="resource-form-block">
          <el-form-item label="资源类型">
            <el-radio-group v-model="formData.mode">
              <el-radio-button :value="E_resourceDramaSeriesType.Movies">视频</el-radio-button>
              <el-radio-button :value="E_resourceDramaSeriesType.Comic">漫画</el-radio-button>
              <el-radio-button :value="E_resourceDramaSeriesType.Atlas">图集</el-radio-button>
              <el-radio-button :value="E_resourceDramaSeriesType.Files">文件</el-radio-button>
              <!--
              <el-radio-button :value="E_resourceDramaSeriesType.VideoLink">
                视频链接
              </el-radio-button>
              <el-radio-button :value="E_resourceDramaSeriesType.NetDisk">
                网盘
              </el-radio-button>
              -->
            </el-radio-group>
          </el-form-item>
          <el-form-item label="资源标题" prop="title">
            <el-input v-model="formData.title" />
          </el-form-item>
          <el-form-item label="版号 / 番号">
            <el-input v-model="formData.issueNumber" />
          </el-form-item>
          <div class="resource-form-two">
            <div>
              <el-form-item label="发行日期">
                <datePicker v-model="formData.issuingDate" />
              </el-form-item>
            </div>
            <div>
              <el-form-item label="国家">
                <selectCountry v-model="formData.country" />
              </el-form-item>
            </div>
            <div>
              <el-form-item label="清晰度">
                <selectDefinition v-model="formData.definition" />
              </el-form-item>
            </div>
            <div>
              <el-form-item label="评分">
                <selectStarSet v-model="formData.stars" />
              </el-form-item>
            </div>
          </div>
          <el-form-item :label="appLang.director()">
            <selectPerformer v-model="directors" :performerBasesIds="store.appStoreData.currentPerformerBasesIds"
              :careerType="E_performerCareerType.Director" multiple />
          </el-form-item>
          <el-form-item :label="appLang.performer()">
            <selectPerformer v-model="performers" :performerBasesIds="store.appStoreData.currentPerformerBasesIds"
              :careerType="E_performerCareerType.Performer" multiple />
          </el-form-item>
          <el-form-item label="摘要">
            <el-input v-model="formData.abstract" :rows="4" type="textarea" />
          </el-form-item>
        </div>
        <div class="resource-form-class-title">
          <el-alert title="标签" type="warning" :closable="false" />
          <el-form-item v-for="tagClass, index in store.appStoreData.currentTagClass" :key="index"
            :label="tagClass.name">
            <selectTag v-model="tags[tagClass.id]" :tagClassId="tagClass.id" dataSource="store" multiple />
          </el-form-item>
        </div>
      </div>
    </div>
  </drawerForm>
  <serverFileManagementDialog ref="serverFileManagementDialogRef" @selectedFiles="selectedFilesHandle">
  </serverFileManagementDialog>
</template>
<script lang="ts" setup>
import { reactive, ref, computed } from 'vue'
import { E_performerCareerType } from '@/dataType/app.dataType';
import drawerForm from '../com/dialog/drawer.form.vue'
import setImage from '../com/form/setImage.vue';
import datePicker from '../com/form/datePicker.vue'
import selectCountry from '../com/form/selectCountry.vue'
import selectDefinition from '../com/form/selectDefinition.vue'
import selectStarSet from '../com/form/selectStarSet.vue'
import selectTag from '../com/form/selectTag.vue'
import selectPerformer from '../com/form/selectPerformer.vue'
import { ElMessage, type FormRules } from 'element-plus'
import { E_resourceDramaSeriesType } from '@/dataType/app.dataType'
import type { I_resource, I_resource_base, I_resourceDramaSeries_base } from '@/dataType/resource.dataType'
import serverFileManagementDialog from '../serverFileManagement/serverFileManagementDialog.vue';
import { appStoreData } from '@/storeData/app.storeData'
import type { I_sfm_FileEntry } from '@/components/serverFileManagement/com/dataType';
import { LoadingService } from '@/assets/loading';
import { resourceServer } from '@/server/resource.server';
import { getResourceCoverPoster } from '@/common/photo';
import { appLang } from '@/language/app.lang';
const store = {
  appStoreData: appStoreData(),
}

const emits = defineEmits(['success'])

const drawerFormRef = ref<InstanceType<typeof drawerForm>>();
const setImageRef = ref<InstanceType<typeof setImage>>();
const serverFileManagementDialogRef = ref<InstanceType<typeof serverFileManagementDialog>>();
const resourceFormCoverPosterContainerRef = ref<HTMLElement | null>(null);

let mode: 'add' | 'edit' = 'add'

const defaultFormData: I_resource_base = {
  id: '',
  filesBases_id: '',
  mode: E_resourceDramaSeriesType.Movies,
  title: '',
  issueNumber: '',
  coverPoster: '',
  coverPosterMode: 0,
  coverPosterWidth: 0,
  coverPosterHeight: 0,
  issuingDate: '',
  country: '',
  definition: '',
  stars: 0,
  abstract: '',
  status: true,
}

const formData = ref<I_resource_base>({ ...defaultFormData })
const performers = ref<string[]>([]);
const directors = ref<string[]>([]);
const tags = ref<Record<string, string[]>>({});
const dramaSeries = ref<I_resourceDramaSeries_base[]>([]);

const formRules = reactive<FormRules>({
  title: [{ required: true, trigger: 'blur', message: '请输入标题' }],
})

const coverPosterHeight_C = computed(() => {
  const index = formData.value.coverPosterMode;
  const coverPosterData = store.appStoreData.currentConfigApp.coverPosterData[index];
  const ratio = coverPosterData.height / coverPosterData.width;
  const containerWidth = drawerFormRef.value?.$el.offsetWidth || 300;
  return `${containerWidth * ratio}px`;
})

const init = (_mode: 'add' | 'edit', res: I_resource | null = null) => {
  mode = _mode;
  setImageRef.value?.init();
  store.appStoreData.currentTagClass.forEach(item => {
    if (mode == 'edit' && res) {
      tags.value[item.id] = res.tags.filter(tag => tag.tagClass_id == item.id).map(tag => tag.id);
    } else {
      tags.value[item.id] = [];
    }
  })
  if (mode == 'edit' && res) {
    const formKeys = Object.keys(defaultFormData);
    formKeys.forEach(key => {
      if (res.hasOwnProperty(key) && key in formData.value) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        (formData.value as any)[key] = (res as any)[key];
      }
    })
    performers.value = res.performers.map(item => item.id);
    directors.value = res.directors.map(item => item.id);
    dramaSeries.value = res.dramaSeries.map(item => ({ id: item.id, src: item.src }));
  } else {
    formData.value = { ...defaultFormData }
    performers.value = [];
    directors.value = [];
    dramaSeries.value = [];
  }
}

const selectServerFilesHandle = () => {
  serverFileManagementDialogRef.value?.open();
}
const selectedFilesHandle = (slc: I_sfm_FileEntry[]) => {
  slc.forEach(item => {
    dramaSeries.value.push({
      id: "",
      src: item.path
    })
  })
}
const dramaSeriesDeleteHandle = (index: number) => {
  dramaSeries.value.splice(index, 1)
}

const submitHandle = async () => {
  if (formData.value.filesBases_id === '') {
    formData.value.filesBases_id = store.appStoreData.currentFilesBases.id;
  }
  const photoBase64 = setImageRef.value?.getImageBase64() || '';
  if (setImageRef.value) {
    const { width, height } = setImageRef.value.getImageSize();
    formData.value.coverPosterWidth = width;
    formData.value.coverPosterHeight = height;
  }
  const _tags: string[] = [];
  for (const tagClassId in tags.value) {
    _tags.push(...tags.value[tagClassId]);
  }
  LoadingService.show();
  try {
    const apiCall = mode === 'add'
      ? resourceServer.create(formData.value, photoBase64, performers.value, directors.value, _tags, dramaSeries.value)
      : resourceServer.update(formData.value, photoBase64, performers.value, directors.value, _tags, dramaSeries.value);
    const result = await apiCall;
    if (result.status) {
      emits('success', result.data);
      close();
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

const open = (_mode: 'add' | 'edit', res: I_resource | null = null) => {
  init(_mode, res);
  drawerFormRef.value?.open();
}
const close = () => {
  drawerFormRef.value?.close();
}

// eslint-disable-next-line no-undef
defineExpose({ open })
</script>
<style lang="scss" scoped>
.resource-form-container {
  display: flex;
  gap: 1em;

  .el-alert {
    padding: 4px 10px;
    margin-bottom: 8px;
  }

  .resource-form-left {
    width: 300px;
    flex-shrink: 0;

    .resource-form-cover-poster {
      margin-top: 5px;
      width: calc(100% - 4px);
      height: auto;
      padding: 1px;
      border-radius: 5px;
      border: 1px solid #4c4d4f;
      flex-shrink: 0;

      .resource-form-cover-poster-text {
        width: 100%;
        height: 100%;
        min-height: 200px;
        display: flex;
        justify-content: center;
        align-items: center;
        font-size: 1.3em;
      }
    }

    .resource-form-mode {
      width: 100%;
      overflow: hidden;
      margin-bottom: 10px;

      .el-radio-group {
        width: 100%;
        display: flex;
        flex-wrap: wrap;
        padding: 5px 0;
        gap: 5px;

        .el-radio {
          min-width: calc(50% - 3px);
          margin: 0;
          white-space: nowrap;
          overflow: hidden;

          :deep(.el-radio__label) {
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
            font-size: 1em;
          }
        }
      }
    }

    .resource-form-drama-series {
      list-style-type: none;

      li {
        display: flex;
        gap: 5px;
        padding-bottom: 5px;

        .drama-series-index {
          width: 24px;
          flex-shrink: 0;
          text-align: center;
          line-height: 24px;
          font-size: 12px;
          font-weight: 500;
          font-style: italic;
        }

        .drama-series-tool {
          flex-shrink: 0;

          .el-button {
            padding: 8px;
          }
        }
      }
    }

    .resource-form-browse {
      display: flex;
      justify-content: center;
      padding: 5px 0;
    }
  }

  .resource-form-right {
    flex-grow: 1;

    .resource-form-two {
      display: flex;
      flex-wrap: wrap;

      >div {
        width: 50%;
        flex-shrink: 0;
        overflow: hidden;
      }
    }
  }
}
</style>
