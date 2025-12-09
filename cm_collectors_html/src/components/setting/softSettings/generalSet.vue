<template>
  <div v-loading="loading" class="setting-data">
    <el-form v-model="formData" label-width="260px">
      <el-form-item label="软件logo名称">
        <el-input v-model="formData.logoName" />
        <el-text class="warning-text" type="warning" size="small">重新加载后生效</el-text>
      </el-form-item>
      <el-form-item label="主题">
        <el-select v-model="formData.theme">
          <el-option label="暗黑" value="default" />
          <el-option label="明亮" value="bright" />
        </el-select>
        <el-text class="warning-text" type="warning" size="small"> 重新加载后生效</el-text>
      </el-form-item>
      <el-form-item label="管理需登录">
        <el-switch v-model="formData.isAdminLogin" />
      </el-form-item>
      <el-form-item label="管理密码">
        <el-input type="password" v-model="formData.adminPassword" show-password />
        <el-text class="warning-text" type="warning" size="small">
          留空表示不修改密码。如果对外开放访问，请务必设置强密码。忘记密码可在服务器配置文件config.yaml中查看或修改
        </el-text>
      </el-form-item>
      <!--
        <el-form-item label="自动创建视频m3u8文件">
          <el-switch v-model="formData.isAutoCreateM3u8" />
        </el-form-item>
        -->
      <el-form-item label="TV Box 地址">
        <div>
          <p>[域名]或[IP:端口号]/api/tvbox/home</p>
          <el-text class="warning-text" type="warning" size="small">
            例如: http://192.168.1.51:12345/api/tvbox/home
          </el-text>
        </div>
      </el-form-item>
      <el-form-item label="不允许服务器打开文件或文件夹">
        <el-switch v-model="formData.notAllowServerOpenFile" />
      </el-form-item>
      <el-form-item label="语言">
        <el-select v-model="formData.language">
          <el-option label="简体中文" value="zhCn" />
        </el-select>
      </el-form-item>
      <el-form-item label="web可播放视频格式">
        <selectPlayVideoFormats v-model="formData.playVideoFormats" multiple />
      </el-form-item>
      <el-form-item label="web可播放音频格式">
        <selectPlayAudioFormats v-model="formData.playAudioFormats" multiple />
      </el-form-item>
      <el-form-item label="文件管理挂载路径">
        <selectServerFileManagementRootPath v-model="formData.serverFileManagementRootPath" multiple />
        <el-text class="warning-text" type="warning" size="small">
          指定挂载磁盘，未在列表中显示的磁盘（例如网络磁盘 \\192.168.1.51\影视）可手动填写后按回车键添加。<br>
          如果指定的路径在服务器上不存在或无法访问，则该路径将不会在文件管理器中显示。
        </el-text>
      </el-form-item>
      <el-form-item label="视频限流器">
        <div>
          <div>
            <el-switch v-model="formData.videoRateLimit.enabled" />
          </div>
          <div>
            <el-text class="warning-text" type="warning" size="small">
              开启限流器后如果限制请求或桶容量过小，会影响云播或导致云播失败。清桶时间为1分钟，限流器修改后，需要等清桶后才能生效。
            </el-text>
          </div>
        </div>
      </el-form-item>
      <el-form-item label="限流器每秒请求限制">
        <el-input-number v-model="formData.videoRateLimit.requestsPerSecond" :min="1" :max="100" />
      </el-form-item>
      <el-form-item label="限流器桶容量">
        <el-input-number v-model="formData.videoRateLimit.burst" :min="1" :max="100" />
      </el-form-item>
      <el-form-item label="刮削器使用指定浏览器">
        <el-switch v-model="formData.scraper.useBrowserPath" />
      </el-form-item>
      <el-form-item label="刮削器指定浏览器路径">
        <el-input v-model="formData.scraper.browserPath">
          <template #append>
            <el-button icon="FolderOpened" @click="openServerFileManagement('scraper')" />
          </template>
        </el-input>
        <el-text class="warning-text" type="warning" size="small">
          <div>启用上方开关后，刮削器将使用此处指定的浏览器路径进行数据抓取。</div>
          <div> 支持的浏览器包括Chrome、Edge等基于Chromium内核的浏览器。</div>
          <div>常见安装位置：</div>
          <div>Windows Chrome: C:/Program Files/Google/Chrome/Application/chrome.exe</div>
          <div>Windows Edge: C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe</div>
          <div>macOS Chrome: /Applications/Google Chrome.app/Contents/MacOS/Google Chrome</div>
          <div>macOS Edge: /Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge</div>
          <div>Linux Chrome: /usr/bin/google-chrome 或 /usr/bin/chromium-browser</div>
          <div>Linux Edge: /usr/bin/microsoft-edge</div>
        </el-text>
      </el-form-item>
      <el-form-item label="windows托盘菜单">
        <div class="tary-menu-container">
          <div class="tary-menu" v-for="item, index in formData.taryMenu" :key="index">
            <el-input class="name" v-model="item.name" placeholder="请输入托盘名" />
            <el-input class="path" v-model="item.path" placeholder="请选择执行程序">
              <template #append>
                <el-button icon="FolderOpened" @click="openServerFileManagement('taryMenu', index)" />
              </template>
            </el-input>
            <el-button class="btn" type="danger" icon="Delete" @click="delTaryMenu(index)" />
          </div>
          <el-button @click="addTaryMenu">添加新托盘菜单</el-button>
          <div>
            <el-text class="warning-text" type="warning" size="small">
              修改托盘菜单后需要重启程序才能生效
            </el-text>
          </div>
        </div>
      </el-form-item>
    </el-form>
    <div class="save-button-container">
      <el-button type="primary" @click="saveHandle" icon="Edit">保存</el-button>
    </div>
  </div>
  <serverFileManagementDialog ref="serverFileManagementDialogRef" @selectedFiles="selectedFilesHandle">
  </serverFileManagementDialog>
</template>
<script lang="ts" setup>
import selectPlayVideoFormats from '@/components/com/form/selectPlayVideoFormats.vue';
import selectPlayAudioFormats from '@/components/com/form/selectPlayAudioFormats.vue';
import selectServerFileManagementRootPath from '@/components/com/form/selectServerFileManagementRootPath.vue';
import serverFileManagementDialog from '@/components/serverFileManagement/serverFileManagementDialog.vue';
import { ref, onMounted } from 'vue'
import type { I_appSystemConfig } from '@/dataType/app.dataType';
import { appDataServer } from '@/server/app.server';
import { ElMessage } from 'element-plus';
import { debounceNow } from '@/assets/debounce';
import type { I_sfm_FileEntry } from '@/components/serverFileManagement/com/dataType';
import dataset from '@/assets/dataset';

const serverFileManagementDialogRef = ref<InstanceType<typeof serverFileManagementDialog>>();

const formData = ref<I_appSystemConfig>({
  logoName: 'CM File Collectors',
  isAdminLogin: false,
  adminPassword: '',
  isAutoCreateM3u8: false,
  language: 'zhCn',
  notAllowServerOpenFile: false,
  theme: 'default',
  playVideoFormats: dataset.playVideoFormats,
  playAudioFormats: dataset.playAudioFormats,
  serverFileManagementRootPath: dataset.serverFileManagementRootPath,
  videoRateLimit: {
    enabled: false,
    requestsPerSecond: 5,
    burst: 10
  },
  scraper: {
    useBrowserPath: false,
    browserPath: ''
  },
  taryMenu: [],
})

const loading = ref(false);


const init = async () => {
  loading.value = false;
  await getAppConfig();
}

const getAppConfig = async () => {
  try {
    loading.value = true;
    const result = await appDataServer.getAppConfig()
    if (!result.status) {
      ElMessage.error(result.msg);
      return
    }
    formData.value = result.data;
    console.log(formData.value);
    return
  } catch (error) {
    console.log(error)
  } finally {
    loading.value = false;
  }
}
let openServerFileManagementOptType = '';
let openServerFileManagementOptIndex = 0;
const openServerFileManagement = (optType: string, index = 0) => {
  openServerFileManagementOptType = optType;
  openServerFileManagementOptIndex = index;
  serverFileManagementDialogRef.value?.open();
}
const selectedFilesHandle = (slc: I_sfm_FileEntry[]) => {
  if (slc.length == 0) {
    return;
  }
  switch (openServerFileManagementOptType) {
    case 'scraper':
      formData.value.scraper.browserPath = slc[slc.length - 1].path;
      break;
    case 'taryMenu':
      formData.value.taryMenu[openServerFileManagementOptIndex].path = slc[slc.length - 1].path;
      break;
  }
}
const addTaryMenu = () => {
  formData.value.taryMenu.push({
    name: '',
    path: ''
  })
}
const delTaryMenu = (index: number) => {
  formData.value.taryMenu.splice(index, 1);
}

const saveHandle = debounceNow(async () => {
  try {
    loading.value = true;
    const result = await appDataServer.setAppConfig(formData.value)
    if (!result.status) {
      ElMessage.error(result.msg);
      return
    }
    ElMessage.success('保存成功');
    return
  } catch (error) {
    console.log(error)
  } finally {
    loading.value = false;
  }
})

onMounted(async () => {
  init();
})
</script>
<style lang="scss" scoped>
.setting-data {
  max-width: 960px;
  height: 100%;
  display: flex;
  gap: 10px;
  flex-direction: column;

  .warning-text {
    line-height: 1.1rem;
    padding-top: 5px;
  }

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

  .tary-menu-container {
    width: 100%;

    .tary-menu {
      display: flex;
      gap: 10px;
      padding-bottom: 8px;

      .name {
        width: 150px;
        flex-shrink: 0;
      }

      .path {
        flex: 1;
      }

      .btn {
        flex-shrink: 0;
      }
    }
  }

}
</style>
