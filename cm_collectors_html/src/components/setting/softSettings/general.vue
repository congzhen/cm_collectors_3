<template>
  <div>
    <el-form label-width="auto" style="max-width: 600px">
      <el-form-item label="软件logo名称">
        <el-input />
      </el-form-item>
      <el-form-item label="服务器端口">
        <el-input />
      </el-form-item>
      <el-form-item label="自动创建视频m3u8文件">
        <el-switch />
      </el-form-item>
      <el-form-item label="语言">
        <el-select />
      </el-form-item>
      <el-form-item label="数据库路径">
        <el-input />
        <alert-msg color="warning">
          可将数据库迁移至软件根目录以外（例如nas的共享文件夹，可以让多个客户端访问该数据库），空值表示放置在软件根目录，选择新目录后需要从新启动软件生效。如果目录不存在，则会自动指定软件根目录。
        </alert-msg>
      </el-form-item>
    </el-form>
  </div>
</template>
<script lang="ts" setup>
import alertMsg from '@/components/com/feedback/alertMsg.vue';
</script>
