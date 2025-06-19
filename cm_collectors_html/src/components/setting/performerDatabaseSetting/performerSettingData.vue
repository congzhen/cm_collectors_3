<template>
  <div class="setting-data">
    <el-form label-width="auto">
      <el-form-item label="演员集名称">
        <div class="form-row">
          <el-input class="form-input" />
          <el-switch inline-prompt active-text="启用" inactive-text="禁用" />
        </div>

      </el-form-item>
      <el-form-item label="关联文件数据库">
        <el-checkbox label="中国数据库" />
        <el-checkbox label="主播数据库" />
        <el-checkbox label="写真" />
        <el-checkbox label="动漫" />
      </el-form-item>
    </el-form>
    <div class="setting-data-performer">
      <performerDataList></performerDataList>
    </div>
  </div>
</template>
<script setup lang="ts">
import performerDataList from '@/components/performer/performerDataList.vue';
</script>
<style lang="scss" scoped>
.setting-data {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;

  .el-form {
    width: 100%;
    padding: 0 20px;
  }

  .setting-data-performer {
    flex: 1;
    overflow: auto;
  }
}
</style>
