<template>
  <div class="setting-data" v-loading="loading">
    <el-form v-if="filesBasesInfo && filesConfig" label-width="auto">

      <el-alert title="基础设置" type="success" />

      <el-form-item label="文件数据库名称">
        <el-input v-model="filesBasesInfo.name" />
      </el-form-item>
      <el-form-item label="(主)演员集">
        <el-select />
      </el-form-item>
      <el-form-item label="关联演员集">
        <el-checkbox label="日本AV女优演员集" />
        <el-checkbox label="中国AV女优演员集" />
        <el-checkbox label="主播演员集" />
        <el-checkbox label="写真演员集" />
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

      <el-alert title="左侧边栏" type="success" />

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
        <selectPerformer v-model="filesConfig.performerPreferred" multiple
          :performer-bases-ids="[store.filesBasesStoreData.getMainPerformerBasesIdByFilesBasesId(filesBasesInfo.id)]" />
      </el-form-item>

      <el-alert title="显示设置" type="success" />

      <el-form-item label="分页显示数量">
        <el-input-number />
      </el-form-item>
      <el-form-item label="排序方式">
        <el-select />
      </el-form-item>
      <el-form-item label="显示模式">
        <el-select />
      </el-form-item>
      <el-form-item label="详情剧集显示模式">
        <el-select />
      </el-form-item>
      <el-form-item label="详情显示模式">
        <el-select />
      </el-form-item>
      <el-form-item label="预览图">
        <el-checkbox label="显示预览图" border />
      </el-form-item>
      <el-form-item label="预览图文件夹，多个文件夹用,分割">
        <el-input />
      </el-form-item>
      <el-form-item label="封面上显示标签">
        <el-input />
      </el-form-item>
      <el-form-item label="标签背景色">
        <el-color-picker show-alpha />
        <el-button-group size="small">
          <el-button icon="Plus" />
          <el-button icon="Minus" />
        </el-button-group>
      </el-form-item>
      <el-form-item label="标签字体颜色">
        <el-color-picker show-alpha />
        <el-button-group size="small">
          <el-button icon="Plus" />
          <el-button icon="Minus" />
        </el-button-group>
      </el-form-item>
      <el-form-item label="随机海报">
        <el-checkbox label="开启随机海报" border />
      </el-form-item>
      <el-form-item label="随机海报宽度">
        <el-input-number />
      </el-form-item>
      <el-form-item label="随机海报高度">
        <el-input-number />
      </el-form-item>
      <el-form-item label="随机海报路径">
        <el-upload>
          <el-button type="primary">Click to upload</el-button>
        </el-upload>
      </el-form-item>

      <el-alert title="参数设置" type="success" />

      <el-form-item label="开启记录模块">
        <el-checkbox label="历史记录" border />
        <el-checkbox label="当前热度" border />
        <el-checkbox label="猜你喜欢" border />
      </el-form-item>
      <el-form-item label="历史记录显示数量">
        <el-input-number />
      </el-form-item>
      <el-form-item label="当前热度显示数量">
        <el-input-number />
      </el-form-item>
      <el-form-item label="猜你喜欢显示数量">
        <el-input-number />
      </el-form-item>
      <el-form-item label="猜你喜欢取词量">
        <el-input-number />
      </el-form-item>
      <el-form-item label="猜你喜欢参与取词的标签分类">
        <el-select />
      </el-form-item>
      <el-form-item label="当前猜你喜欢词汇">
        <el-tag type="primary">Tag 1</el-tag>
        <el-tag type="primary">Tag 2</el-tag>
        <el-tag type="primary">Tag 3</el-tag>
      </el-form-item>

      <el-alert title="图集设置" type="success" />

      <el-form-item label="图集显示模式">
        <el-select />
      </el-form-item>
      <el-form-item label="图集分批读取数量">
        <el-input-number />
      </el-form-item>

      <el-alert title="演员&导演自定义" type="success" />
      <el-form-item label="演员显示文字">
        <el-input />
      </el-form-item>
      <el-form-item label="导演显示文字">
        <el-input />
      </el-form-item>
      <el-form-item label="演员默认头像">
        <el-input />
      </el-form-item>

      <el-alert title="插件设置" type="success" />
      <el-form-item label="演员Cup插件">
        <el-checkbox label="开启演员Cup插件" border />
        <alert-msg color="warning">
          该插件在演员资料中添加Cup选项，并在左边栏出现Cup标签选择。
        </alert-msg>
      </el-form-item>
      <el-form-item label="Cup显示文字">
        <el-input />
      </el-form-item>

      <el-alert title="封面海报设置" type="success" />
      <el-form-item label="封面海报显示宽度">
        <el-checkbox label="开启封面海报宽度控制" border />
        <alert-msg color="warning">
          开启该功能，会限定每个资源封面海报的宽度。
        </alert-msg>
      </el-form-item>
      <el-form-item label="宽度基数">
        <el-input-number />
      </el-form-item>

      <el-form-item label="封面海报显示高度">
        <el-checkbox label="开启封面海报高度控制" border />
        <alert-msg color="warning">
          开启该功能，会限定每个资源封面海报的高度。
        </alert-msg>
      </el-form-item>
      <el-form-item label="高度基数">
        <el-input-number />
      </el-form-item>

      <el-alert title="路径虚拟转换" type="success" />
      <alert-msg color="warning">
        视频文件夹整体移动位置时，可以使用虚拟路径转换功能，例如：from D:\video to E:\myVideo，*!to SoftwareDrive:\myVideo
        则转换至软件当前所在的盘符地址!*，如果需要真实转换数据库中的地址，请使用数据库资源路径替换器。
      </alert-msg>
      <el-button>添加虚拟路径</el-button>

    </el-form>
  </div>
</template>
<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import selectCountry from '@/components/com/form/selectCountry.vue';
import selectDefinition from '@/components/com/form/selectDefinition.vue';
import selectLeftDisplay from '@/components/com/form/selectLeftDisplay.vue';
import selectLeftColumnMode from '@/components/com/form/selectLeftColumnMode.vue';
import selectTagMode from '@/components/com/form/selectTagMode.vue';
import selectPerformer from '@/components/com/form/selectPerformer.vue';
import alertMsg from '@/components/com/feedback/alertMsg.vue';
import { filesBasesServer } from '@/server/filesBases.server';
import { ElMessage } from 'element-plus';
import type { I_filesBases_base } from '@/dataType/filesBases.dataType';
import { defualtConfigApp, type I_config_app } from '@/dataType/config.dataType';
import { filesBasesStoreData } from '@/storeData/filesBases.storeData';
const store = {
  filesBasesStoreData: filesBasesStoreData(),
}
const props = defineProps({
  filesBasesId: {
    type: String,
    required: true,
  },
})

const loading = ref(false);
const filesBasesInfo = ref<I_filesBases_base>();
const filesConfig = ref<I_config_app>();
const init = async () => {
  await getFielsBasesInfo();
}

const getFielsBasesInfo = async () => {
  loading.value = true;
  const result = await filesBasesServer.infoById(props.filesBasesId);
  if (!result.status) {
    ElMessage.error(result.msg);
    return;
  }
  filesBasesInfo.value = {
    id: result.data.id,
    name: result.data.name,
    sort: result.data.sort,
    addTime: result.data.addTime,
    status: result.data.status,
  }
  if (result.data.filesBasesSetting.config_json_data != '') {
    filesConfig.value = JSON.parse(result.data.filesBasesSetting.config_json_data);
  } else {
    filesConfig.value = defualtConfigApp;
  }
  loading.value = false;
  console.log(result);
}

onMounted(() => {
  init()
})

</script>
<style lang="scss" scoped>
.setting-data {
  width: 100%;
  height: 100%;
  overflow: auto;

  .el-form {
    width: 900px;
    padding: 0 20px;

    .el-alert {
      margin-bottom: 10px;
    }

    .alert-msg {
      padding: 0 10px;
    }
  }
}
</style>
