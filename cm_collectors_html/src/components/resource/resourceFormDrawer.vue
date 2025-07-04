<template>
  <drawerForm ref="drawerFormRef" :modelValue="formData" :rules="formRules" width="1024px" title="资源">
    <div class="resource-form-container">
      <div class="resource-form-left">
        <div class="resource-form-class-title">
          <el-alert title="封面" type="info" :closable="false" />
        </div>
        <div class="resource-form-cover-poster">
          <div class="resource-form-cover-poster-text">
            <label>点击或拖拽上传照片</label>
          </div>
        </div>
        <div class="resource-form-mode">
          <el-radio-group v-model="formData.cover_poster_mode">
            <el-radio :value="1" border>Default 1</el-radio>
            <el-radio :value="2" border>Default 2</el-radio>
            <el-radio :value="3" border>Default 3</el-radio>
            <el-radio :value="4" border>Default 4</el-radio>
            <el-radio :value="0" border>
              Default 5Default 5Default 5Default 5Default 5Default 5Default 5
            </el-radio>
          </el-radio-group>
        </div>
        <div>
          <el-alert title="资源" type="info" :closable="false" />
          <ul class="resource-form-drama-series">
            <li>
              <label class="drama-series-index">1.</label>
              <el-input />
              <el-button-group class="drama-series-tool">
                <el-button icon="FolderOpened"></el-button>
                <el-button icon="Delete"></el-button>
              </el-button-group>
            </li>
            <li>
              <label class="drama-series-index">2.</label>
              <el-input />
              <el-button-group class="drama-series-tool">
                <el-button icon="FolderOpened"></el-button>
                <el-button icon="Delete"></el-button>
              </el-button-group>
            </li>
            <li>
              <label class="drama-series-index">3.</label>
              <el-input />
              <el-button-group class="drama-series-tool">
                <el-button icon="FolderOpened"></el-button>
                <el-button icon="Delete"></el-button>
              </el-button-group>
            </li>
          </ul>
          <div class="resource-form-browse">
            <el-button-group>
              <el-button icon="Monitor">浏览本地</el-button>
              <el-button icon="MostlyCloudy">浏览服务器</el-button>
            </el-button-group>
          </div>
          <div class="resource-drag-upload">
            <el-upload action="/" :on-change="dropFileOrFolderHere" :show-file-list="false" :auto-upload="false"
              :multiple="true" drag>
              <div class="el-upload__text">拖拽文件或文件夹至此</div>
            </el-upload>
          </div>
        </div>
      </div>
      <div class="resource-form-right">
        <div class="resource-form-class-title">
          <el-alert title="信息" type="success" :closable="false" />
        </div>
        <div class="resource-form-block">
          <el-form-item label="资源类型">
            <el-radio-group v-model="formData.mode" size="small">
              <el-radio-button :value="E_resourceDramaSeriesType.Movies">视频</el-radio-button>
              <el-radio-button :value="E_resourceDramaSeriesType.Comic">漫画</el-radio-button>
              <el-radio-button :value="E_resourceDramaSeriesType.Atlas">图集</el-radio-button>
              <el-radio-button :value="E_resourceDramaSeriesType.Files">文件</el-radio-button>
              <el-radio-button :value="E_resourceDramaSeriesType.VideoLink">
                视频链接
              </el-radio-button>
              <el-radio-button :value="E_resourceDramaSeriesType.NetDisk">网盘</el-radio-button>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="资源标题" prop="title">
            <el-input v-model="formData.title" />
          </el-form-item>
          <el-form-item label="版号 / 番号">
            <el-input v-model="formData.issue_number" />
          </el-form-item>
          <div class="resource-form-two">
            <div>
              <el-form-item label="发行日期">
                <datePicker v-model="formData.issuing_date" />
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
          <el-form-item label="导演">
            <el-input v-model="formData.issue_number" />
          </el-form-item>
          <el-form-item label="演员">
            <el-input v-model="formData.issue_number" />
          </el-form-item>
          <el-form-item label="摘要">
            <el-input v-model="formData.abstract" :rows="4" type="textarea" />
          </el-form-item>
        </div>
        <div class="resource-form-class-title">
          <el-alert title="标签" type="warning" :closable="false" />
        </div>
      </div>
    </div>
  </drawerForm>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue'
import drawerForm from '../com/dialog/drawer.form.vue'
import datePicker from '../com/form/datePicker.vue'
import selectCountry from '../com/form/selectCountry.vue'
import selectDefinition from '../com/form/selectDefinition.vue'
import selectStarSet from '../com/form/selectStarSet.vue'
import type { FormRules, UploadFile } from 'element-plus'
import { E_resourceDramaSeriesType } from '@/dataType/app.dataType'
import type { I_resource } from '@/dataType/resource.dataType'
import { el } from 'element-plus/es/locales.mjs'
const drawerFormRef = ref<InstanceType<typeof drawerForm>>()
const formData = reactive<I_resource>({
  id: '',
  files_bases_id: '',
  mode: E_resourceDramaSeriesType.Movies,
  title: '',
  issue_number: '',
  cover_poster: '',
  cover_poster_mode: 0,
  cover_poster_width: '',
  cover_poster_height: '',
  issuing_date: '',
  country: '',
  definition: '',
  stars: 0,
  abstract: '',
  tags: {},
  directors: [],
  performers: [],
  drama_series: [],
})
const formRules = reactive<FormRules>({
  title: [{ required: true, trigger: 'blur', message: '请输入标题' }],
})

const dropFileOrFolderHere = async (uploadFile: UploadFile) => {
  console.log(uploadFile.raw)
}

const open = () => {
  drawerFormRef.value?.open()
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
  }

  .resource-form-left {
    width: 380px;
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
      padding: 5px 0;

      li {
        display: flex;
        gap: 5px;
        padding-top: 5px;

        .drama-series-index {
          width: 30px;
          flex-shrink: 0;
          text-align: center;
          line-height: 32px;
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
