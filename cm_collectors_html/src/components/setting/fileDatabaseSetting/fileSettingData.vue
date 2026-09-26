<template>
  <div class="setting-data" v-loading="loading">
    <el-form v-if="finish" label-width="auto">
      <SharedConfigBar ref="sharedBar" :files-bases-id="props.filesBasesId" module="display" :config="filesConfig"
        local-hint="标签与演员选择、路径、自定义头像及封面预设列表保持本库独立。" @config="applySharedConfig" @saved="emit('setSuccess', props.filesBasesId)" />

      <SettingSectionTitle>文件库信息</SettingSectionTitle>
      <el-form-item label="文件数据库名称">
        <el-input v-model="filesBasesInfo.name" />
      </el-form-item>
      <el-form-item label="(主)演员集">
        <el-select v-model="mainPerformerBasesId">
          <el-option v-for="item, index in store.performerBasesStoreData.listByIds(filesBasesRelatedPerformerBases)"
            :key="index" :label="item.name" :value="item.id"></el-option>
        </el-select>
      </el-form-item>
      <el-form-item label="关联演员集">
        <el-checkbox-group v-model="filesBasesRelatedPerformerBases">
          <el-checkbox v-for="item, key in store.performerBasesStoreData.performerBases" :key="key" :label="item.name"
            :value="item.id" :disabled="item.id == mainPerformerBasesId" />
        </el-checkbox-group>
      </el-form-item>
      <el-form-item label="状态">
        <el-switch v-model="filesBasesInfo.status" inline-prompt active-text="启用" inactive-text="禁用" />
      </el-form-item>

      <SharedDisplayFields :config="filesConfig" :readonly="!!sharedBar?.state?.following">
        <template #actors>
          <el-form-item label="优先显示演员">
            <selectPerformer ref="selectPerformerRef" v-model="filesConfig.performerPreferred" multiple
              :careerType="E_performerCareerType.Performer"
              :performer-bases-ids="[store.filesBasesStoreData.getMainPerformerBasesIdByFilesBasesId(filesBasesInfo.id)]" />
          </el-form-item>

          <el-form-item :label="filesConfig.performerPreferredEnabled ? '其余演员排序' : '演员排序'">
            <el-select v-model="filesConfig.performerSortMode">
              <el-option label="默认排序" value="default" />
              <el-option label="资源数量最多" value="resourceCountDesc" />
              <el-option label="播放热度最高" value="hotDesc" />
              <el-option label="近期偏好" value="recentDesc" />
            </el-select>
            <el-checkbox v-model="filesConfig.performerPreferredEnabled" label="启用自定义优先演员" />
            <div class="performer-sort-hint">取消勾选后全部按所选规则排序，上方选择仍会保留；所有模式均遵守“屏蔽无照片演员”。</div>
          </el-form-item>
          <el-form-item v-if="filesConfig.performerSortMode === 'recentDesc'" label="近期统计天数">
            <el-input-number v-model="filesConfig.performerRecentDays" :min="1" :max="365" :precision="0" />
            <div class="performer-sort-hint">按当前库最近 N 天（含今天）的播放次数排序；从升级后开始统计，无记录时按默认排序补齐。</div>
          </el-form-item>
        </template>
        <template #tags>
          <el-form-item label="封面上显示标签(自定义)">
            <selectTag ref="selectTagRef" v-model="filesConfig.coverDisplayTag" data-source="database"
              :filesBasesId="props.filesBasesId" multiple reorder />
          </el-form-item>
        </template>
        <template #sample>
          <el-form-item label="剧照相对文件夹">
            <el-input v-model="filesConfig.sampleFolder" />
          </el-form-item>
        </template>
        <template #avatar>
          <el-form-item label="自定义头像">
            <setCustomAvatar v-model="filesConfig.performer_photo" />
          </el-form-item>
        </template>
        <template #poster>
          <el-form-item label="封面海报">
            <coverPosterAdmin v-model:cover-poster-data-default-select="filesConfig.coverPosterDataDefaultSelect"
              v-model:cover-poster-data="filesConfig.coverPosterData" />
          </el-form-item>
        </template>
      </SharedDisplayFields>
    </el-form>

    <div class="save-button-container">
      <el-button-group>
        <el-button :disabled="sharedBar?.state?.following || sharedBar?.dialogOpen" @click="importHandle">导入</el-button>
        <el-button @click="exportHandle">导出</el-button>
      </el-button-group>
      <el-button type="primary" @click="saveHandle" :disabled="sharedBar?.dialogOpen" icon="Edit">保存本库设置</el-button>
    </div>
  </div>
</template>
<script lang="ts" setup>
import SettingSectionTitle from '@/components/setting/SettingSectionTitle.vue';
import { onMounted, ref } from 'vue';
import SharedConfigBar from '@/components/setting/SharedConfigBar.vue';
import SharedDisplayFields from '@/components/setting/sharedConfig/SharedDisplayFields.vue';
import { E_performerCareerType } from '@/dataType/app.dataType';
import selectPerformer from '@/components/com/form/selectPerformer.vue';
import selectTag from '@/components/com/form/selectTag.vue';
//import selectPlayAtlasMode from '@/components/com/form/selectPlayAtlasMode.vue';
import coverPosterAdmin from './coverPosterAdmin.vue';
//import routeConversionAdmin from './routeConversionAdmin.vue';
import setCustomAvatar from '@/components/com/form/setCustomAvatar.vue';
import { filesBasesServer } from '@/server/filesBases.server';
import { ElMessage } from 'element-plus';
import type { I_filesBases_base } from '@/dataType/filesBases.dataType';
import { createDefaultConfigApp, defualtConfigApp, type I_config_app } from '@/dataType/config.dataType';
import { filesBasesStoreData } from '@/storeData/filesBases.storeData';
import { performerBasesStoreData } from '@/storeData/performerBases.storeData';
import { debounceNow } from '@/assets/debounce';
import { filesBasesConfigExport, filesBasesConfigImport } from '@/common/filesBasesConfig';

const store = {
  filesBasesStoreData: filesBasesStoreData(),
  performerBasesStoreData: performerBasesStoreData(),
}
const props = defineProps({
  filesBasesId: {
    type: String,
    required: true,
  },
})
const emit = defineEmits(['setSuccess']);

const selectPerformerRef = ref<InstanceType<typeof selectPerformer>>()
const selectTagRef = ref<InstanceType<typeof selectTag>>()

const sharedBar = ref<InstanceType<typeof SharedConfigBar>>();
const applySharedConfig = (config: object) => { filesConfig.value = config as I_config_app; };
const finish = ref(false);
const loading = ref(false);
const filesBasesInfo = ref<I_filesBases_base>({} as I_filesBases_base);
const mainPerformerBasesId = ref('');
const filesBasesRelatedPerformerBases = ref<string[]>([]);
const filesConfig = ref<I_config_app>({} as I_config_app);

const init = async () => {
  await getFielsBasesInfo();
}

//获取FielsBases信息
const getFielsBasesInfo = async () => {
  // 开始加载时设置加载状态为true
  loading.value = true;
  // 调用后端API，根据ID获取信息
  const result = await filesBasesServer.infoById(props.filesBasesId);

  // 如果获取信息失败，显示错误消息并返回
  if (!result.status) {
    ElMessage.error(result.msg);
    return;
  }

  // 更新FielsBases信息
  filesBasesInfo.value = {
    id: result.data.id,
    name: result.data.name,
    sort: result.data.sort,
    addTime: result.data.addTime,
    status: result.data.status,
  }

  // 初始化关联信息数组
  filesBasesRelatedPerformerBases.value = [];
  // 遍历结果中的关联执信息
  result.data.filesRelatedPerformerBases.forEach(item => {
    if (item.main) {
      // 设置主演员集ID
      mainPerformerBasesId.value = item.performerBases_id;
    }
    // 将演员集ID添加到关联信息数组中
    filesBasesRelatedPerformerBases.value.push(item.performerBases_id);
  });

  // 解析配置数据
  if (result.data.filesBasesSetting.config_json_data != '') {
    const parsedConfig = JSON.parse(result.data.filesBasesSetting.config_json_data);
    const mergedConfig: I_config_app = createDefaultConfigApp();
    // 如果配置数据不存在，则使用默认配置值
    for (const key in defualtConfigApp) {
      if (parsedConfig.hasOwnProperty(key)) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        (mergedConfig as any)[key] = parsedConfig[key];
      }
    }
    filesConfig.value = mergedConfig;
  } else {
    filesConfig.value = createDefaultConfigApp();
  }
  finish.value = true
  // 加载完成后设置加载状态为false
  loading.value = false;
}

const saveHandle = debounceNow(async () => {
  if (sharedBar.value?.dialogOpen) return;
  if (!filesBasesRelatedPerformerBases.value.includes(mainPerformerBasesId.value)) {
    ElMessage.error('请设置主演员集');
    return;
  }
  const result = await filesBasesServer.setData(props.filesBasesId, filesBasesInfo.value, filesConfig.value, mainPerformerBasesId.value, filesBasesRelatedPerformerBases.value);
  if (!result.status) {
    ElMessage.error(result.msg);
    return;
  }
  ElMessage.success('保存成功');
  emit('setSuccess', props.filesBasesId);
})

const exportHandle = debounceNow(async () => {
  const newFilesConfig = JSON.parse(JSON.stringify(filesConfig.value)) as I_config_app;
  const performers = selectPerformerRef.value?.getOptionsData()
  if (performers) {
    const performerPreferred = newFilesConfig.performerPreferred
      .map(id => performers.find(item => item.id === id))
      .filter(item => item !== undefined)
      .map(item => item!.name);
    newFilesConfig.performerPreferred = performerPreferred
  }
  const tags = selectTagRef.value?.getOptionsData()
  if (tags) {
    const tagPreferred = newFilesConfig.coverDisplayTag
      .map(id => tags.find(item => item.id === id))
      .filter(item => item !== undefined)
      .map(item => item!.name);
    newFilesConfig.coverDisplayTag = tagPreferred
  }
  await filesBasesConfigExport(props.filesBasesId, newFilesConfig)
})
const importHandle = debounceNow(async () => {
  const data = await filesBasesConfigImport()
  if (data == null) return

  // 将名称转换回ID
  const performers = selectPerformerRef.value?.getOptionsData()
  if (performers && data.performerPreferred) {
    const performerPreferredIds = data.performerPreferred
      .map(name => performers.find(item => item.name === name))
      .filter(item => item !== undefined)
      .map(item => item!.id);
    data.performerPreferred = performerPreferredIds
  }

  const tags = selectTagRef.value?.getOptionsData()
  if (tags && data.coverDisplayTag) {
    const coverDisplayTagIds = data.coverDisplayTag
      .map(name => tags.find(item => item.name === name))
      .filter(item => item !== undefined)
      .map(item => item!.id);
    data.coverDisplayTag = coverDisplayTagIds
  }
  console.log(data);
  // 更新配置
  filesConfig.value = { ...createDefaultConfigApp(), ...data }
})

onMounted(() => {
  init()
})

</script>
<style lang="scss" scoped>
.setting-data {
  max-width: 960px;
  height: 100%;
  display: flex;
  gap: 10px;
  flex-direction: column;

  .el-form {
    flex: 1;
    padding: 0 20px;
    overflow-y: auto;
    overflow-x: hidden;

    .el-alert {
      margin-bottom: 10px;
    }

    .alert-msg {
      padding: 0 10px;
    }
  }

  .save-button-container {
    flex-shrink: 1;
    padding: 5px 15px;
    background-color: #262727;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
}
.performer-sort-hint {
  width: 100%;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.6;
  margin-top: 6px;
}
</style>
