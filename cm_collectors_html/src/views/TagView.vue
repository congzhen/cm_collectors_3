<template>
  <div class="tag-container">
    <el-scrollbar>
      <div class="tag-block-list">
        <el-collapse v-model="activeNames">
          <tagCollapseItem name="1"></tagCollapseItem>
          <tagCollapseItem name="2" :tag-type="E_tagType.Star"></tagCollapseItem>
          <tagCollapseItem name="3" :tag-type="E_tagType.Performer"></tagCollapseItem>
          <tagCollapseItem name="4"></tagCollapseItem>
          <tagCollapseItem name="5"></tagCollapseItem>
          <tagCollapseItem name="6"></tagCollapseItem>
          <tagCollapseItem name="7"></tagCollapseItem>
          <tagCollapseItem name="8"></tagCollapseItem>
          <tagCollapseItem name="9"></tagCollapseItem>
          <tagCollapseItem name="10"></tagCollapseItem>
          <tagCollapseItem name="11"></tagCollapseItem>
          <tagCollapseItem name="12"></tagCollapseItem>
        </el-collapse>
      </div>
    </el-scrollbar>
  </div>
</template>
<script setup lang="ts">
import { E_tagType } from '@/dataType/app.dataType'
import tagCollapseItem from '@/components/tag/tagCollapseItem.vue'
import { ref } from 'vue'
const activeNames = ref(['1', '2', '3', '4', '5', '6', '7', '8', '9', '10', '11', '12'])
</script>
<style lang="scss" scoped>
.tag-container {
  width: 24.4em;
  height: 100%;
  overflow: hidden;
  border-right: 0.1em solid #414243;
  margin-right: 6px;
  .el-collapse {
    border: 0;
    :deep(.el-collapse-item__header) {
      border: 0;
      height: 3em;
      line-height: 3em;
    }
    :deep(.el-collapse-item__wrap) {
      border: 0;
    }
    :deep(.el-collapse-item__content) {
      padding-bottom: 2px;
    }
  }

  .tag-block-list {
    padding-right: 6px;
  }
}
</style>
