<template>
  <div v-loading="loading" class="setting-data">
    <el-form v-model="formData" label-position="left" label-width="220px" class="general-form">
      <div class="settings-grid">
        <section class="settings-card" aria-label="外观与语言">
          <div class="card-heading">
            <el-icon class="card-icon"><Monitor /></el-icon>
            <div>
              <h3>外观与语言</h3>
              <p>设置软件名称、主题与页面风格。</p>
            </div>
          </div>
          <div class="settings-fields">
            <el-form-item label="软件logo名称">
              <el-input class="name-control" v-model="formData.logoName" />
            </el-form-item>
            <el-form-item label="主题">
              <el-select class="compact-control" v-model="formData.theme" @change="themeChangeHandle">
                <el-option label="暗黑" value="default" />
                <el-option label="明亮" value="bright" />
              </el-select>
            </el-form-item>
            <el-form-item label="主页样式">
              <el-select class="compact-control" v-model="formData.homeMode">
                <el-option v-for="item in homeModeOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
            </el-form-item>
            <el-form-item label="界面外观">
              <el-select class="compact-control" v-model="formData.appearanceStyle">
                <el-option label="经典外观" value="classic" />
                <el-option label="现代外观" value="modern" />
              </el-select>
              <el-text class="appearance-help" type="info" size="small">
                统一控制顶部栏、详情弹窗和筛选面板的显示样式。
              </el-text>
            </el-form-item>
            <el-form-item label="语言">
              <el-select class="compact-control" v-model="formData.language">
                <el-option label="简体中文" value="zhCn" />
              </el-select>
            </el-form-item>
          </div>
        </section>
        <section class="settings-card" aria-label="设备与播放">
          <div class="card-heading">
            <el-icon class="card-icon"><VideoPlay /></el-icon>
            <div>
              <h3>设备与播放</h3>
              <p>调整设备适配和浏览器支持的媒体格式。</p>
            </div>
          </div>
          <div class="settings-fields">
            <el-form-item class="switch-row" label="关闭移动端页面布局">
              <el-switch v-model="formData.closeMobileDisplay" />
              <el-text type="info" size="small">只影响页面布局，不限制大屏设备使用移动端播放器。</el-text>
            </el-form-item>
            <el-form-item label="大屏移动设备播放器">
              <el-select class="compact-control" v-model="formData.largeMobilePlayer">
                <el-option label="桌面播放器" value="desktop" />
                <el-option label="移动端播放器" value="mobile" />
              </el-select>
              <el-text class="field-help" type="info" size="small">保留桌面页面布局，下次打开播放器生效。</el-text>
            </el-form-item>
            <el-form-item label="大屏判定尺寸（高级）">
              <div>
                <label class="dimension-field">短边 <el-input-number v-model="formData.largeMobileShortSide" :min="320" :max="4096" :precision="0" /></label>
                <label class="dimension-field">长边 <el-input-number v-model="formData.largeMobileLongSide" :min="320" :max="4096" :precision="0" /></label>
                <div><el-text type="info" size="small">单位为浏览器视口 CSS px，不是屏幕物理分辨率。移动设备任意一边达到阈值即使用大屏播放策略；不改变网站页面布局。</el-text></div>
              </div>
            </el-form-item>
            <el-form-item label="web可播放视频格式">
              <selectPlayVideoFormats v-model="formData.playVideoFormats" multiple />
            </el-form-item>
            <el-form-item label="web可播放音频格式">
              <selectPlayAudioFormats v-model="formData.playAudioFormats" multiple />
            </el-form-item>
          </div>
        </section>
        <section class="settings-card" aria-label="访问与安全">
          <div class="card-heading">
            <el-icon class="card-icon"><Lock /></el-icon>
            <div>
              <h3>访问与安全</h3>
              <p>管理登录验证、服务器权限和文件访问。</p>
            </div>
          </div>
          <div class="settings-fields">
            <el-form-item class="switch-row" label="管理需登录">
              <el-switch v-model="formData.isAdminLogin" />
            </el-form-item>
            <el-form-item label="管理密码">
              <el-input class="password-control" type="password" v-model="formData.adminPassword" show-password />
              <el-text class="field-help" type="info" size="small">
                留空表示不修改密码。忘记密码可在服务器配置文件 config.yaml 中查看或修改。
              </el-text>
              <el-text class="warning-text" type="warning" size="small">
                如果对外开放访问，请务必设置强密码。
              </el-text>
            </el-form-item>
            <el-form-item class="switch-row" label="不允许服务器打开文件或文件夹">
              <el-switch v-model="formData.notAllowServerOpenFile" />
            </el-form-item>
            <el-form-item class="switch-row" label="允许应用程序关闭服务器">
              <el-switch v-model="formData.allowAppCloseServer" />
            </el-form-item>
            <el-form-item label="文件管理挂载路径">
              <selectServerFileManagementRootPath v-model="formData.serverFileManagementRootPath" multiple />
              <el-text class="field-help" type="info" size="small">
                指定挂载磁盘，未在列表中显示的磁盘（例如网络磁盘 \\192.168.1.51\影视）可手动填写后按回车键添加。<br>
                如果指定的路径在服务器上不存在或无法访问，则该路径将不会在文件管理器中显示。
              </el-text>
            </el-form-item>
          </div>
        </section>
        <section class="settings-card" aria-label="TV Box 与云播">
          <div class="card-heading">
            <el-icon class="card-icon"><Connection /></el-icon>
            <div>
              <h3>TV Box 与云播</h3>
              <p>配置电视端访问和云播播放偏好。</p>
            </div>
          </div>
          <div class="settings-fields">
            <el-form-item label="TV Box 地址">
              <div>
                <p>[域名]或[IP:端口号]/api/tvbox/home</p>
                <el-text class="field-help" type="info" size="small">
                  例如: http://192.168.1.51:12345/api/tvbox/home
                </el-text>
              </div>
            </el-form-item>
            <el-form-item class="switch-row" label="TV Box 开启">
              <el-switch v-model="formData.tvBoxEnabled" />
            </el-form-item>
            <el-form-item class="switch-row" label="关闭云播功能">
              <el-switch v-model="formData.closePlayCloud" />
              <p class="row-help">云播插件下载地址：<a @click="playCloudPluginDownload">{{ playCloudPluginDownloadUrl }}</a></p>
            </el-form-item>
            <el-form-item class="switch-row" label="关闭云播询问弹窗">
              <el-switch v-model="formData.closePlayCloudDialog" />
            </el-form-item>
            <el-form-item label="默认云播播放方式">
              <el-radio-group v-model="formData.playCloudMode">
                <el-radio-button value="m3u8">m3u8</el-radio-button>
                <el-radio-button label="mp4">mp4</el-radio-button>
              </el-radio-group>
            </el-form-item>
          </div>
        </section>
        <section class="settings-card" aria-label="视频请求限流">
          <div class="card-heading">
            <el-icon class="card-icon"><Odometer /></el-icon>
            <div>
              <h3>视频请求限流</h3>
              <p>控制视频请求频率与桶容量。</p>
            </div>
          </div>
          <div class="settings-fields">
            <el-form-item class="switch-row" label="视频限流器">
              <el-switch v-model="formData.videoRateLimit.enabled" />
              <div class="row-help">
                <el-text class="warning-text" type="warning" size="small">请求限制或桶容量过小，可能影响云播或导致云播失败。</el-text>
                <el-text class="field-help" type="info" size="small">清桶时间为 1 分钟，修改后需等清桶后生效。</el-text>
              </div>
            </el-form-item>
            <el-form-item label="限流器每秒请求限制">
              <el-input-number v-model="formData.videoRateLimit.requestsPerSecond" :min="1" :max="100" />
            </el-form-item>
            <el-form-item label="限流器桶容量">
              <el-input-number v-model="formData.videoRateLimit.burst" :min="1" :max="100" />
            </el-form-item>
          </div>
        </section>
        <section class="settings-card" aria-label="刮削浏览器">
          <div class="card-heading">
            <el-icon class="card-icon"><Search /></el-icon>
            <div>
              <h3>刮削浏览器</h3>
              <p>指定数据抓取使用的浏览器。</p>
            </div>
          </div>
          <div class="settings-fields">
            <el-form-item class="switch-row" label="刮削器使用指定浏览器">
              <el-switch v-model="formData.scraper.useBrowserPath" />
            </el-form-item>
            <el-form-item label="刮削器指定浏览器路径">
              <el-input v-model="formData.scraper.browserPath">
                <template #append>
                  <el-button icon="FolderOpened" @click="openServerFileManagement('scraper')" />
                </template>
              </el-input>
              <el-text class="field-help" type="info" size="small">
                启用上方开关后，使用指定浏览器抓取数据。支持 Chrome、Edge 等 Chromium 内核浏览器。
              </el-text>
              <details class="browser-path-help">
                <summary>查看常见浏览器安装路径</summary>
                <dl class="browser-path-list">
                  <dt>Windows · Chrome</dt>
                  <dd><code>C:/Program Files/Google/Chrome/Application/chrome.exe</code></dd>
                  <dt>Windows · Edge</dt>
                  <dd><code>C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe</code></dd>
                  <dt>macOS · Chrome</dt>
                  <dd><code>/Applications/Google Chrome.app/Contents/MacOS/Google Chrome</code></dd>
                  <dt>macOS · Edge</dt>
                  <dd><code>/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge</code></dd>
                  <dt>Linux · Chrome</dt>
                  <dd><code>/usr/bin/google-chrome</code> 或 <code>/usr/bin/chromium-browser</code></dd>
                  <dt>Linux · Edge</dt>
                  <dd><code>/usr/bin/microsoft-edge</code></dd>
                </dl>
              </details>
            </el-form-item>
          </div>
        </section>
        <section class="settings-card" aria-label="Windows 与托盘">
          <div class="card-heading">
            <el-icon class="card-icon"><Grid /></el-icon>
            <div>
              <h3>Windows 与托盘</h3>
              <p>设置启动行为和托盘快捷菜单。</p>
            </div>
          </div>
          <div class="settings-fields">
            <el-form-item class="switch-row" label="Windows版启动，不打开应用程序">
              <el-switch v-model="formData.windowsStartNotRunApp" />
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
                  <el-text class="field-help" type="info" size="small">
                    修改托盘菜单后需要重启程序才能生效
                  </el-text>
                </div>
              </div>
            </el-form-item>
          </div>
        </section>
      </div>
    </el-form>
    <div class="save-button-container">
      <el-text type="info" size="small">
        当前配置需要重新加载后生效
      </el-text>
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
import { playCloudPluginDownload, playCloudPluginDownloadUrl } from '@/components/play/playCloud'
import { homeModeOptions } from '@/common/homeMode'
import { appStoreData } from '@/storeData/app.storeData'
import { normalizePlayerConfig } from '@/common/playerPreference'
import { setCloseMobileDisplay } from '@/assets/mobile'
import { applyTheme } from '@/common/theme'

const serverFileManagementDialogRef = ref<InstanceType<typeof serverFileManagementDialog>>();
const store = appStoreData()
const themeChangeHandle = (theme: string) => {
  store.appConfig.theme = theme
  applyTheme(theme)
}

const formData = ref<I_appSystemConfig>({
  logoName: 'CM File Collectors',
  isAdminLogin: false,
  adminPassword: '',
  language: 'zhCn',
  notAllowServerOpenFile: false,
  allowAppCloseServer: false,
  theme: 'default',
  homeMode: 'classic',
  appearanceStyle: 'modern',
  detailsDialogStyle: 'classic',
  headerStyle: 'modern',
  closeMobileDisplay: false,
  ...normalizePlayerConfig({}),
  closePlayCloud: true,
  closePlayCloudDialog: true,
  playCloudMode: 'm3u8',
  playVideoFormats: dataset.playVideoFormats,
  playAudioFormats: dataset.playAudioFormats,
  serverFileManagementRootPath: dataset.serverFileManagementRootPath,
  windowsStartNotRunApp: false,
  tvBoxEnabled: true,
  videoRateLimit: {
    enabled: false,
    requestsPerSecond: 5,
    burst: 10
  },
  scraper: {
    useBrowserPath: false,
    browserPath: ''
  },
  autoBackup: {
    enabled: false,
    backupPath: './auto_backup',
    intervalHours: 0,
    resourceChangeThreshold: 0,
    maxBackups: 5,
  },
  performerAvatarLibrary: {
    cachePath: './runtime/cache/gfriends/',
    customBaseUrl: '',
    defaultStrategy: 'recommended',
    clearCacheOnStartup: false,
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
    if (!formData.value.homeMode) {
      formData.value.homeMode = 'classic'
    }
    if (!formData.value.appearanceStyle) {
      formData.value.appearanceStyle = formData.value.headerStyle || formData.value.detailsDialogStyle || 'modern'
    }
    if (!formData.value.detailsDialogStyle) {
      formData.value.detailsDialogStyle = 'classic'
    }
    if (!formData.value.headerStyle) {
      formData.value.headerStyle = 'modern'
    }
    Object.assign(formData.value, normalizePlayerConfig(formData.value))
    formData.value.closeMobileDisplay = formData.value.closeMobileDisplay === true
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
    const short = formData.value.largeMobileShortSide;
    const long = formData.value.largeMobileLongSide;
    if (!short || !long || short > long) {
      ElMessage.error('请填写有效尺寸，短边不能大于长边');
      return;
    }
    const result = await appDataServer.setAppConfig(formData.value)
    if (!result.status) {
      ElMessage.error(result.msg);
      return
    }
    ElMessage.success('保存成功');
    store.appConfig = {
      ...store.appConfig,
      logoName: formData.value.logoName,
      isAdminLogin: formData.value.isAdminLogin,
      theme: formData.value.theme,
      homeMode: formData.value.homeMode,
      appearanceStyle: formData.value.appearanceStyle,
      detailsDialogStyle: formData.value.detailsDialogStyle,
      headerStyle: formData.value.headerStyle,
      closeMobileDisplay: formData.value.closeMobileDisplay,
      ...normalizePlayerConfig(formData.value),
      closePlayCloud: formData.value.closePlayCloud,
      closePlayCloudDialog: formData.value.closePlayCloudDialog,
      playCloudMode: formData.value.playCloudMode,
    }
    // 保存后立即同步移动端显示开关，避免必须刷新页面才生效。
    applyTheme(formData.value.theme)
    setCloseMobileDisplay(formData.value.closeMobileDisplay)
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
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  color: var(--el-text-color-primary);
}
.general-form {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 0 8px 12px;
  scrollbar-gutter: stable;
}
.settings-grid {
  max-width: 1320px;
  margin: 0 auto;
}
.settings-card {
  display: grid;
  grid-template-columns: 190px minmax(0, 1fr);
  gap: 16px;
  padding: 14px 0;
  border-bottom: 1px solid var(--el-border-color-lighter);
  &:first-child { padding-top: 4px; }
  &:last-child { border-bottom: 0; }
}
.card-heading {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding-top: 8px;
  h3 { font-size: 15px; font-weight: 600; line-height: 24px; }
  p { margin-top: 6px; color: var(--el-text-color-secondary); font-size: 12px; line-height: 1.7; }
}
.card-icon { flex-shrink: 0; margin-top: 3px; font-size: 18px; color: var(--el-text-color-secondary); }
.settings-fields {
  min-width: 0;
  padding: 0 16px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  background: var(--el-bg-color);
  > :deep(.el-form-item) {
    margin: 0;
    padding: 8px 0;
    border-bottom: 1px solid var(--el-border-color-extra-light);
    &:last-child { border-bottom: 0; }
    > .el-form-item__label { height: auto; line-height: 24px; padding: 4px 20px 4px 0; color: var(--el-text-color-regular); }
    > .el-form-item__content { min-width: 0; gap: 6px 10px; align-content: center; overflow-wrap: anywhere; }
  }
  :deep(.el-text), .row-help { line-height: 1.65; font-size: 12px; }
  .appearance-help, .field-help, .warning-text { display: block; width: 100%; margin-top: 2px; }
  .row-help { color: var(--el-text-color-secondary); }
  .compact-control { width: 240px; max-width: 100%; }
  .name-control { width: 360px; max-width: 100%; }
  .password-control { width: 560px; max-width: 100%; }
  .dimension-field { display: inline-flex; align-items: center; gap: 8px; margin: 0 12px 8px 0; }
  > :deep(.switch-row > .el-form-item__content) {
    flex-direction: row;
    flex-wrap: nowrap;
    justify-content: flex-start;
    > .el-switch { flex: 0 0 auto; }
    > .el-text, > .row-help { flex: 1; min-width: 0; margin: 0; text-align: left; }
  }
}
// Native disclosure keeps the examples keyboard-accessible without changing form state.
.browser-path-help {
  width: 100%;
  min-width: 0;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  summary {
    width: fit-content;
    max-width: 100%;
    padding: 4px 0;
    line-height: 1.6;
    cursor: pointer;
    color: var(--el-color-primary);
    &:focus-visible { outline: 2px solid var(--el-color-primary); outline-offset: 3px; border-radius: 2px; }
  }
}
.browser-path-list {
  margin: 8px 0 0;
  padding: 12px 14px;
  border-radius: 6px;
  background: var(--el-fill-color-light);
  line-height: 1.7;
  dt { margin-top: 10px; font-weight: 500; color: var(--el-text-color-regular); }
  dt:first-child { margin-top: 0; }
  dd { margin: 2px 0 0; overflow-wrap: anywhere; }
  code { font-size: 12px; }
}
.save-button-container {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 16px;
  padding: 8px 16px;
  border-top: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);
  .el-button { min-width: 96px; }
}
.tary-menu-container {
  width: 100%;
  .tary-menu { display: flex; flex-wrap: wrap; gap: 8px; padding-bottom: 10px; }
  .name { width: 120px; }
  .path { flex: 1; min-width: 150px; }
  .btn { flex-shrink: 0; }
}
@media (max-width: 1200px) {
  .settings-card { grid-template-columns: minmax(0, 1fr); gap: 12px; padding: 12px 0; }
  .card-heading { padding-top: 0; p { margin-top: 2px; } }
}
@media (max-width: 700px) {
  .general-form { padding: 0 0 16px; }
  .settings-fields {
    padding: 0 14px;
    .compact-control, .name-control, .password-control { width: 100%; }
    > :deep(.el-form-item) {
      display: block;
      > .el-form-item__label { width: auto !important; display: block; padding: 0 0 8px; }
    }
    > :deep(.switch-row) {
      display: flex;
      flex-wrap: wrap;
      > .el-form-item__label { width: 100% !important; }
      > .el-form-item__content { flex-basis: 100%; }
    }
  }
  .save-button-container { padding: 10px 12px; gap: 10px; }
}
</style>
