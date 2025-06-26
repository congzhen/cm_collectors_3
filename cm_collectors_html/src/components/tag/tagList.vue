<template>
  <div class="tag-list" v-loading="loading">
    <div v-if="tagData">
      <div class="tag-list-block" v-for="tagClass, key in tagData.tagClass" :key="key">
        <div class="tag-list-header">
          <label class="title">{{ tagClass.name }}</label>
          <div class="tool">
            <el-button-group>
              <el-button icon="ArrowUpBold" size="small" />
              <el-button icon="ArrowDownBold" size="small" />
            </el-button-group>
            <div>
              <el-checkbox v-model="tagClass.leftShow" label="左边栏显示" size="small" />
            </div>
            <el-button-group>
              <el-button icon="Edit" size="small" />
              <el-button icon="Delete" size="small" @click="tagClassDeleteHandle" />
            </el-button-group>
          </div>
        </div>
        <div class="tag-list-body">
          <tagItem v-for="tag, key in getTag(tagClass.id)" :key="key" :name="tag.name" />
          <el-button icon="CirclePlus" @click="tagItemFormHandle('add', '', '', '')" />
        </div>
      </div>
    </div>
  </div>
  <tagFormClassDialog ref="tagFormClassDialogRef"></tagFormClassDialog>
  <tagFormItemDialog ref="tagFormItemDialogRef"></tagFormItemDialog>
</template>
<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import tagFormClassDialog from './tagFormClassDialog.vue';
import tagFormItemDialog from './tagFormItemDialog.vue';
import tagItem from './tagItem.vue';
import { messageBoxConfirm } from '@/components/com/feedback/messageBox'
import { tagServer } from '@/server/tag.server';
import { ElMessage } from 'element-plus';
import type { I_tagData } from '@/dataType/tag.dataType';

const props = defineProps({
  id: {
    type: String,
    required: true,
  },
})
const loading = ref(false);
const tagData = ref<I_tagData>();

const tagFormItemDialogRef = ref<InstanceType<typeof tagFormItemDialog>>();

const init = async () => {
  await getTagData();
}

const getTagData = async () => {
  loading.value = true;
  // 获取标签数据
  const result = await tagServer.tagDataByFilesBasesId(props.id);
  if (result && result.status) {
    tagData.value = result.data;
  } else {
    ElMessage.error(result.msg);
  }
  loading.value = false;
}

const getTag = (tagClassId: string) => {
  if (!tagData.value) return [];
  return tagData.value.tag.filter(tag => tag.tagClass_id === tagClassId);
}



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

onMounted(async () => {
  await init()
})

</script>
<style lang="scss" scoped>
.tag-list {
  color: #a8abb2;
  height: 100%;

  .tag-list-block {
    padding-bottom: 10px;

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
