<template>
  <div v-loading="loading">
    <el-form v-model="formData" label-width="260px" style="max-width: 760px">
      <el-form-item label="软件logo名称">
        <el-input v-model="formData.logoName" />
        <el-text class="warning-text" type="warning" size="small">重新加载后生效</el-text>
      </el-form-item>
      <el-form-item label="管理需登录">
        <el-switch v-model="formData.isAdminLogin" />
      </el-form-item>
      <el-form-item label="管理密码">
        <el-input type="password" v-model="formData.adminPassword" show-password />
        <el-text class="warning-text" type="warning" size="small">留空表示不修改密码</el-text>
      </el-form-item>
      <el-form-item label="自动创建视频m3u8文件">
        <el-switch v-model="formData.isAutoCreateM3u8" />
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
      <el-form-item>
        <el-button type="primary" @click="saveHandle">保存</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>
<script lang="ts" setup>
import selectPlayVideoFormats from '@/components/com/form/selectPlayVideoFormats.vue';
import selectPlayAudioFormats from '@/components/com/form/selectPlayAudioFormats.vue';
import { ref, onMounted } from 'vue'
import type { I_appSystemConfig } from '@/dataType/app.dataType';
import { appDataServer } from '@/server/app.server';
import { ElMessage } from 'element-plus';
import { debounceNow } from '@/assets/debounce';
const formData = ref<I_appSystemConfig>({
  logoName: 'CM File Collectors',
  isAdminLogin: false,
  adminPassword: '',
  isAutoCreateM3u8: false,
  language: 'zhCn',
  playVideoFormats: ['h264', 'vp8', 'vp9', 'av1', 'hevc'],
  playAudioFormats: ['aac', 'opus', 'mp3', 'vorbis', 'pcm_s16le', 'pcm_s24le'],
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
.warning-text {
  line-height: 1.1rem;
}
</style>
