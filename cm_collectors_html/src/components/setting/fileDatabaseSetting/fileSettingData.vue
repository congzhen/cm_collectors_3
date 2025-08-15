<template>
  <div class="setting-data" v-loading="loading">
    <el-form v-if="finish" label-width="auto">

      <el-alert title="基础设置" type="success" :closable="false" />

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
      <el-form-item label="国家列表">
        <selectCountry v-model="filesConfig.country" multiple />
      </el-form-item>
      <el-form-item label="清晰度">
        <selectDefinition v-model="filesConfig.definition" multiple />
      </el-form-item>

      <el-alert title="左侧边栏" type="success" :closable="false" />

      <el-form-item label="左侧边栏显示项">
        <selectLeftDisplay v-model="filesConfig.leftDisplay" multiple />
      </el-form-item>
      <el-form-item label="左侧边栏显示模式">
        <selectLeftColumnMode v-model="filesConfig.leftColumnMode" />
      </el-form-item>
      <el-form-item label="左侧边栏宽度">
        <el-input-number v-model="filesConfig.leftColumnWidth" :min="100" />
      </el-form-item>
      <el-form-item label="标签显示模式">
        <selectTagMode v-model="filesConfig.tagMode" />
      </el-form-item>
      <el-form-item label="演员标签">
        <el-checkbox v-model="filesConfig.performerPhoto" label="显示演员照片" border />
        <el-checkbox v-model="filesConfig.shieldNoPerformerPhoto" label="屏蔽无照片演员" border />
      </el-form-item>
      <el-form-item label="演员标签显示数量">
        <el-input-number v-model="filesConfig.performerShowNum" />
      </el-form-item>
      <el-form-item label="优先显示演员">
        <selectPerformer v-model="filesConfig.performerPreferred" multiple :careerType="E_performerCareerType.Performer"
          :performer-bases-ids="[store.filesBasesStoreData.getMainPerformerBasesIdByFilesBasesId(filesBasesInfo.id)]" />
      </el-form-item>

      <el-alert title="显示设置" type="success" :closable="false" />

      <el-form-item label="分页显示数量">
        <el-input-number v-model="filesConfig.pageLimit" />
      </el-form-item>
      <el-form-item label="资源显示模式">
        <selectResourcesMode v-model="filesConfig.resourcesShowMode" />
      </el-form-item>
      <el-form-item v-if="filesConfig.resourcesShowMode == 'coverPosterBox'" label="封面海报盒子-信息宽度">
        <el-input-number v-model="filesConfig.coverPosterBoxInfoWidth" :min="20" :max="9999" />
      </el-form-item>
      <el-form-item v-if="filesConfig.resourcesShowMode == 'coverPosterWaterfall'" label="封面海报瀑布流-列数">
        <el-input-number v-model="filesConfig.coverPosterWaterfallColumn" :min="1" :max="20" />
      </el-form-item>
      <el-form-item label="详情剧集显示模式">
        <selectDetailsDramaSeriesMode v-model="filesConfig.detailsDramaSeriesMode" />
      </el-form-item>
      <el-form-item label="详情显示模式">
        <selectResourceDetailsShowMode v-model="filesConfig.resourceDetailsShowMode" />
      </el-form-item>
      <!--
      <el-form-item label="预览图">
        <el-checkbox v-model="filesConfig.showPreviewImage" label="显示预览图" border />
      </el-form-item>
      <el-form-item label="预览图文件夹，多个文件夹用,分割">
        <el-input v-model="filesConfig.previewImageFolder" />
      </el-form-item>
      -->
      <el-form-item label="封面上显示标签(属性)">
        <el-select v-model="filesConfig.coverDisplayTagAttribute" multiple>
          <el-option :label="appLang.attributeTags('definition')" value="definition" />
          <el-option :label="appLang.attributeTags('year')" value="issuingDate" />
          <el-option :label="appLang.attributeTags('country')" value="country" />
          <el-option :label="appLang.attributeTags('starRating')" value="stars" />
        </el-select>
      </el-form-item>
      <el-form-item label="封面上显示标签(自定义)">
        <selectTag v-model="filesConfig.coverDisplayTag" data-source="database" :filesBasesId="props.filesBasesId"
          multiple />
      </el-form-item>
      <el-form-item label="标签背景色">
        <div class="color-picker-block">
          <div v-for="_, index in filesConfig.coverDisplayTagRgbas" :key="index">
            <el-color-picker v-model="filesConfig.coverDisplayTagRgbas[index]" show-alpha />
          </div>
          <el-button-group class="color-picker-btn" size="small">
            <el-button icon="Plus" @click="filesConfig.coverDisplayTagRgbas.push(getRandomColor())" />
            <el-button icon="Minus" @click="filesConfig.coverDisplayTagRgbas.pop()" />
          </el-button-group>
        </div>
      </el-form-item>
      <el-form-item label="标签字体颜色">
        <div class="color-picker-block">
          <div v-for="_, index in filesConfig.coverDisplayTagColors" :key="index">
            <el-color-picker v-model="filesConfig.coverDisplayTagColors[index]" show-alpha />
          </div>
          <el-button-group class="color-picker-btn" size="small">
            <el-button icon="Plus" @click="filesConfig.coverDisplayTagColors.push(getRandomColor())" />
            <el-button icon="Minus" @click="filesConfig.coverDisplayTagColors.pop()" />
          </el-button-group>
        </div>
      </el-form-item>
      <!--
      <el-form-item label="随机海报">
        <el-checkbox v-model="filesConfig.randomPosterStatus" label="开启随机海报" border />
        <el-checkbox v-model="filesConfig.randomPosterAutoSize" label="随机海报自适应宽高" border />
      </el-form-item>
      <el-form-item label="随机海报宽度">
        <el-input-number v-model="filesConfig.randomPosterWidth" />
      </el-form-item>
      <el-form-item label="随机海报高度">
        <el-input-number v-model="filesConfig.randomPosterHeight" />
      </el-form-item>
      <el-form-item label="随机海报路径">
        <el-input v-model="filesConfig.randomPosterPath" placeholder="随机海报文件夹所在路径" />
      </el-form-item>
      -->
      <el-alert title="参数设置" type="success" :closable="false" />

      <el-form-item label="视频 - 打开方式">
        <el-select v-model="filesConfig.openResModeMovies">
          <el-option label="内置" :value="E_resourceOpenMode.Soft" />
          <el-option label="系统" :value="E_resourceOpenMode.System" />
        </el-select>
      </el-form-item>
      <el-form-item label="内置播放器" v-if="filesConfig.openResModeMovies === E_resourceOpenMode.Soft">
        <el-select v-model="filesConfig.openResModeMovies_SoftType">
          <el-option label="窗口模式" :value="E_resourceOpenMode_SoftType.Windows" />
          <el-option label="弹窗模式" :value="E_resourceOpenMode_SoftType.Dialog" />
        </el-select>
      </el-form-item>
      <el-form-item label="漫画 - 打开方式">
        <el-select v-model="filesConfig.openResModeComic">
          <el-option label="内置" :value="E_resourceOpenMode.Soft" />
          <el-option label="系统" :value="E_resourceOpenMode.System" />
        </el-select>
      </el-form-item>
      <el-form-item label="图集 - 打开方式">
        <el-select v-model="filesConfig.openResModeAtlas">
          <el-option label="内置" :value="E_resourceOpenMode.Soft" />
          <el-option label="系统" :value="E_resourceOpenMode.System" />
        </el-select>
      </el-form-item>
      <!--
      <el-form-item label="开启记录模块">
        <el-checkbox v-model="filesConfig.historyModule" label="历史记录" border />
        <el-checkbox v-model="filesConfig.hotModule" label="当前热度" border />
        <el-checkbox v-model="filesConfig.youLikeModule" label="猜你喜欢" border />
      </el-form-item>
      <el-form-item label="历史记录显示数量">
        <el-input-number v-model="filesConfig.historyNumber" />
      </el-form-item>
      <el-form-item label="当前热度显示数量">
        <el-input-number v-model="filesConfig.hotNumber" />
      </el-form-item>
      <el-form-item label="猜你喜欢显示数量">
        <el-input-number v-model="filesConfig.youLikeNumber" />
      </el-form-item>

      <el-form-item label="猜你喜欢取词量">
        <el-input-number v-model="filesConfig.youLikeWordNumber" />
      </el-form-item>
      <el-form-item label="猜你喜欢参与取词的标签分类">
        <selectTagClass v-model="filesConfig.youLikeTagClass" :filesBasesId="props.filesBasesId" multiple />
      </el-form-item>
      <el-form-item label="当前猜你喜欢词汇">
        <el-tag type="primary">Tag 1</el-tag>
        <el-tag type="primary">Tag 2</el-tag>
        <el-tag type="primary">Tag 3</el-tag>
      </el-form-item>
      -->

      <!--
      <el-alert title="图集设置" type="success" :closable="false" />
      <el-form-item label="图集显示模式">
        <selectPlayAtlasMode v-model="filesConfig.playAtlasMode" />
      </el-form-item>
      <el-form-item label="图集分批读取数量">
        <el-input-number v-model="filesConfig.playAtlasPageLimit" :min="10" :max="1000" />
      </el-form-item>
      <el-form-item label="图集缩略图">
        <el-checkbox v-model="filesConfig.playAtlasThumbnail" label="图集缩略图" border />
      </el-form-item>
      -->

      <el-alert title="演员&导演自定义" type="success" :closable="false" />
      <el-form-item label="演员显示文字">
        <el-input v-model="filesConfig.performer_Text" />
      </el-form-item>
      <el-form-item label="导演显示文字">
        <el-input v-model="filesConfig.director_Text" />
      </el-form-item>
      <el-form-item label="自定义头像">
        <setCustomAvatar v-model="filesConfig.performer_photo" />
      </el-form-item>

      <el-alert title="插件设置" type="success" :closable="false" />
      <el-form-item label="Cup插件">
        <el-checkbox v-model="filesConfig.plugInUnit_Cup" label="开启演员Cup插件" border />
        <alert-msg color="warning">
          该插件在演员资料中添加Cup选项，并在左边栏出现Cup标签选择。
        </alert-msg>
      </el-form-item>
      <el-form-item label="Cup显示文字">
        <el-input v-model="filesConfig.plugInUnit_Cup_Text" />
      </el-form-item>

      <el-alert title="封面海报设置" type="success" :closable="false" />
      <el-form-item label="封面海报">
        <coverPosterAdmin v-model:cover-poster-data-default-select="filesConfig.coverPosterDataDefaultSelect"
          v-model:cover-poster-data="filesConfig.coverPosterData" />
      </el-form-item>
      <el-form-item label="封面海报显示宽度">
        <el-checkbox v-model="filesConfig.coverPosterWidthStatus" label="开启封面海报宽度控制" border />
        <alert-msg color="warning">
          开启该功能，会限定每个资源封面海报的宽度。
        </alert-msg>
      </el-form-item>
      <el-form-item label="宽度基数">
        <el-input-number v-model="filesConfig.coverPosterWidthBase" />
      </el-form-item>

      <el-form-item label="封面海报显示高度">
        <el-checkbox v-model="filesConfig.coverPosterHeightStatus" label="开启封面海报高度控制" border />
        <alert-msg color="warning">
          开启该功能，会限定每个资源封面海报的高度。
        </alert-msg>
      </el-form-item>
      <el-form-item label="高度基数">
        <el-input-number v-model="filesConfig.coverPosterHeightBase" />
      </el-form-item>

      <!--
      <el-alert title="路径虚拟转换" type="success" :closable="false" />
      <el-form-item label="转换配置">
        <routeConversionAdmin v-model:route-conversion="filesConfig.routeConversion"></routeConversionAdmin>
      </el-form-item>
      -->


    </el-form>
    <!-- 保存按钮 -->
    <div class="save-button-container">
      <el-button type="primary" @click="saveHandle" icon="Edit">保存</el-button>
    </div>
  </div>
</template>
<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { E_performerCareerType, E_resourceOpenMode, E_resourceOpenMode_SoftType } from '@/dataType/app.dataType';
import selectCountry from '@/components/com/form/selectCountry.vue';
import selectDefinition from '@/components/com/form/selectDefinition.vue';
import selectLeftDisplay from '@/components/com/form/selectLeftDisplay.vue';
import selectLeftColumnMode from '@/components/com/form/selectLeftColumnMode.vue';
import selectTagMode from '@/components/com/form/selectTagMode.vue';
import selectPerformer from '@/components/com/form/selectPerformer.vue';
import selectResourcesMode from '@/components/com/form/selectResourcesMode.vue';
import selectDetailsDramaSeriesMode from '@/components/com/form/selectDetailsDramaSeriesMode.vue';
import selectResourceDetailsShowMode from '@/components/com/form/selectResourceDetailsShowMode.vue';
import selectTag from '@/components/com/form/selectTag.vue';
//import selectPlayAtlasMode from '@/components/com/form/selectPlayAtlasMode.vue';
import coverPosterAdmin from './coverPosterAdmin.vue';
//import routeConversionAdmin from './routeConversionAdmin.vue';
import setCustomAvatar from '@/components/com/form/setCustomAvatar.vue';
import alertMsg from '@/components/com/feedback/alertMsg.vue';
import { filesBasesServer } from '@/server/filesBases.server';
import { ElMessage } from 'element-plus';
import type { I_filesBases_base } from '@/dataType/filesBases.dataType';
import { defualtConfigApp, type I_config_app } from '@/dataType/config.dataType';
import { filesBasesStoreData } from '@/storeData/filesBases.storeData';
import { performerBasesStoreData } from '@/storeData/performerBases.storeData';
import { debounceNow } from '@/assets/debounce';
import { getRandomColor } from '@/assets/tool';
import { appLang } from '@/language/app.lang';
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
    const mergedConfig: I_config_app = { ...defualtConfigApp };
    // 如果配置数据不存在，则使用默认配置值
    for (const key in defualtConfigApp) {
      if (parsedConfig.hasOwnProperty(key)) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        (mergedConfig as any)[key] = parsedConfig[key];
      }
    }
    filesConfig.value = mergedConfig;
  } else {
    filesConfig.value = defualtConfigApp;
  }
  finish.value = true
  // 加载完成后设置加载状态为false
  loading.value = false;
}

const saveHandle = debounceNow(async () => {
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

onMounted(() => {
  init()
})

</script>
<style lang="scss" scoped>
.setting-data {
  width: 960px;
  height: 100%;
  display: flex;
  gap: 10px;
  flex-direction: column;

  .el-form {
    flex: 1;
    padding: 0 20px;
    overflow: auto;

    .el-alert {
      margin-bottom: 10px;
    }

    .alert-msg {
      padding: 0 10px;
    }
  }

  .color-picker-block {
    display: flex;
    gap: 6px;

    .color-picker-btn {
      display: flex;
      align-items: center;
    }
  }

  .save-button-container {
    flex-shrink: 1;
    padding: 5px 15px;
    background-color: #262727;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
