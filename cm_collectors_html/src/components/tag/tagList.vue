<template>
  <div class="tag-list">
    <div class="tag-list-block">
      <div class="tag-list-header">
        <label class="title">角色</label>
        <div class="tool">
          <el-button-group>
            <el-button icon="ArrowUpBold" size="small" />
            <el-button icon="ArrowDownBold" size="small" />
          </el-button-group>
          <div>
            <el-checkbox label="左边栏显示" size="small" />
          </div>
          <el-button-group>
            <el-button icon="Edit" size="small" />
            <el-button icon="Delete" size="small" @click="tagClassDeleteHandle" />
          </el-button-group>
        </div>
      </div>
      <div class="tag-list-body">
        <tagItem />
        <tagItem />
        <tagItem />
        <tagItem />
        <tagItem />
        <el-button icon="CirclePlus" @click="tagItemFormHandle('add', '', '', '')" />
      </div>
    </div>
  </div>
  <tagFormClassDialog ref="tagFormClassDialogRef"></tagFormClassDialog>
  <tagFormItemDialog ref="tagFormItemDialogRef"></tagFormItemDialog>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import tagFormClassDialog from './tagFormClassDialog.vue';
import tagFormItemDialog from './tagFormItemDialog.vue';
import tagItem from './tagItem.vue';
import { messageBoxConfirm } from '@/components/com/feedback/messageBox'

const tagFormItemDialogRef = ref<InstanceType<typeof tagFormItemDialog>>();
const tagItemFormHandle = (mode: 'add' | 'edit', id: string, name: string, tagClassId: string) => {
  tagFormItemDialogRef.value?.open(mode, id, name, tagClassId)
}

const tagClassDeleteHandle = () => {
  messageBoxConfirm({
    text: '确定要删除吗？',

    successCallBack: () => {
      console.log('删除成功')
    },
    failCallBack: () => {
      console.log('取消删除')
    },
  })
}

</script>
<style lang="scss" scoped>
.tag-list {
  color: #a8abb2;

  .tag-list-block {
    .tag-list-header {
      height: 2em;
      padding: 0.5em 0;
      border: 1px solid #4c4d4f;
      border-top-left-radius: 3px;
      border-top-right-radius: 3px;
      background-color: #202121;
      display: flex;
      justify-content: space-between;

      .title {
        font-size: 1.2em;
        padding: 0 1em;
        font-weight: 500;
      }

      .tool {
        padding: 0 0.5em;
        display: flex;
        gap: 0.8em;
      }
    }

    .tag-list-body {
      border-left: 1px solid #4c4d4f;
      border-right: 1px solid #4c4d4f;
      border-bottom: 1px solid #4c4d4f;
      border-bottom-left-radius: 3px;
      border-bottom-right-radius: 3px;
      display: flex;
      flex-wrap: wrap;
      gap: 0.5em;
      padding: 1em;
    }
  }
}
</style>
