<template>
  <div class="tag-list" v-loading="loading">
    <div v-if="tagClass">
      <div :class="['tag-list-block', item.status ? '' : 'disable']" v-for="item, key in tagClass" :key="key">
        <div class="tag-list-header">
          <label class="title">{{ item.name }}</label>
          <div class="tool">
            <el-button-group>
              <el-button icon="ArrowUpBold" size="small" @click="moveTagClassUp(key)" :disabled="key === 0" />
              <el-button icon="ArrowDownBold" size="small" @click="moveTagClassDown(key)"
                :disabled="key === tagClass.length - 1" />
            </el-button-group>
            <div>
              <el-checkbox v-model="item.leftShow" label="左边栏显示" size="small" @change="changeTagClassLeftShow(item)" />
            </div>
            <el-button-group>
              <el-button icon="Edit" size="small" @click="tagClassEditHandle(item)" />
              <el-button v-if="item.status" icon="Delete" size="small" @click="tagClassDeleteHandle(item)" />
              <el-button v-else icon="RefreshLeft" size="small" @click="tagClassRestorationHandle(item)" />
            </el-button-group>
          </div>
        </div>
        <div class="tag-list-body">
          <draggable class="tag-list-draggable" :list="tagObjectList[item.id]" item-key="id" @end="onDragEnd">
            <template #item="{ element, index }">
              <tagItem :key="index" :name="element.name" :status="element.status" @edit="editTagHandle(element)"
                @delete="deleteTagHandle(element)" @restore="restoreTagHandle(element)" />
            </template>
          </draggable>
          <el-button icon="CirclePlus" @click="createTagHandle(item.id)" />
        </div>
      </div>

    </div>
  </div>
  <tagFormClassDialog ref="tagFormClassDialogRef" @success="successHandle"></tagFormClassDialog>
  <tagFormItemDialog ref="tagFormItemDialogRef" @success="successHandle"></tagFormItemDialog>
</template>
<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import tagFormClassDialog from './tagFormClassDialog.vue';
import tagFormItemDialog from './tagFormItemDialog.vue';
import tagItem from './tagItem.vue';
import { messageBoxConfirm } from '@/components/com/feedback/messageBox'
import { tagServer } from '@/server/tag.server';
import { ElMessage } from 'element-plus';
import type { I_tag, I_tagClass, I_tagSort } from '@/dataType/tag.dataType';
import draggable from 'vuedraggable';
import { debounce } from '@/assets/debounce';

interface TagMap {
  [key: string]: I_tag[];
}

const props = defineProps({
  id: {
    type: String,
    required: true,
  },
})

const emits = defineEmits(['updateTagDataCompleted'])

const loading = ref(false);
const tagClass = ref<I_tagClass[]>([]);
const tagObjectList = ref<TagMap>({});

const tagFormClassDialogRef = ref<InstanceType<typeof tagFormClassDialog>>();
const tagFormItemDialogRef = ref<InstanceType<typeof tagFormItemDialog>>();

// 初始化
const init = async (fn: () => void = () => { }) => {
  tagObjectList.value = {};
  await getTagData();
  fn();
}


// 获取标签数据
const getTagData = async () => {
  loading.value = true;
  // 获取标签数据
  const result = await tagServer.tagDataByFilesBasesId(props.id);
  if (result && result.status) {
    tagClass.value = result.data.tagClass;
    result.data.tag.forEach(tag => {
      if (!tagObjectList.value[tag.tagClass_id]) {
        tagObjectList.value[tag.tagClass_id] = [];
      }
      tagObjectList.value[tag.tagClass_id].push(tag);
    });
  } else {
    ElMessage.error(result.msg);
  }
  loading.value = false;
}

// 保存标签排序
const saveTagDataSort = debounce(async () => {
  const tagClassSort: I_tagSort[] = [];
  tagClass.value.forEach((_tagClass, _tagClassIndex) => {
    if (_tagClass.sort != _tagClassIndex) {
      tagClassSort.push({ id: _tagClass.id, sort: _tagClassIndex })
    }
  });
  const tagSort: I_tagSort[] = [];
  for (const tagClassId in tagObjectList.value) {
    tagObjectList.value[tagClassId].forEach((tag, tagIndex) => {
      if (tag.sort != tagIndex) {
        tagSort.push({ id: tag.id, sort: tagIndex })
      }
    });
  }
  const result = await tagServer.setTagDataSort(tagClassSort, tagSort);
  if (result && result.status) {
    // 更新tagClass
    tagClassSort.forEach(item => {
      const _tc = tagClass.value.find(tagClassItem => tagClassItem.id == item.id)
      if (_tc) {
        _tc.sort = item.sort
      }
    })
    // 更新tag
    tagSort.forEach(item => {
      for (const tagClassId in tagObjectList.value) {
        const _tag = tagObjectList.value[tagClassId].find(tagItem => tagItem.id == item.id)
        if (_tag) {
          _tag.sort = item.sort
        }
      }
    })
    emits('updateTagDataCompleted')
  } else {
    ElMessage.error(result.msg);
  }
})
// 保存tagClass
const saveTagClass = debounce(async (tagClass: I_tagClass) => {
  const result = await tagServer.updateTagClass(tagClass);
  if (result && result.status) {
    emits('updateTagDataCompleted')
  } else {
    ElMessage.error(result.msg);
  }
})

// 标签分类移动
const moveTagClassUp = (index: number) => {
  if (index <= 0 || index >= tagClass.value.length) return;
  const list = tagClass.value;
  [list[index - 1], list[index]] = [list[index], list[index - 1]];
  saveTagDataSort();
};

// 标签分类移动
const moveTagClassDown = (index: number) => {
  if (index < 0 || index >= tagClass.value.length - 1) return;
  const list = tagClass.value;
  [list[index + 1], list[index]] = [list[index], list[index + 1]];
  saveTagDataSort();
};

// 拖拽排序结束
const onDragEnd = () => {
  saveTagDataSort();
};

// 标签分类显示隐藏
const changeTagClassLeftShow = (tagClass: I_tagClass) => {
  saveTagClass(tagClass);
}

const tagClassEditHandle = (tagClass: I_tagClass) => {
  tagFormClassDialogRef.value?.open('edit', tagClass)
}

const successHandle = () => {
  const scrollContainer = document.querySelector('.tagListDrawer .el-drawer__body');
  const scrollTop = scrollContainer?.scrollTop;
  emits('updateTagDataCompleted');
  init(() => {
    if (scrollTop) {
      scrollContainer.scrollTop = scrollTop;
    }
  });
}

const tagClassDeleteHandle = (tagClass: I_tagClass) => {
  messageBoxConfirm({
    text: '确定要删除吗？',
    successCallBack: async () => {
      tagClass.status = false;
      const result = await tagServer.updateTagClass(tagClass);
      if (result && result.status) {
        emits('updateTagDataCompleted')
      } else {
        ElMessage.error(result.msg);
      }
    },
    failCallBack: () => {
      //console.log('取消删除')
    },
  })
}
const tagClassRestorationHandle = (tagClass: I_tagClass) => {
  messageBoxConfirm({
    text: '确定要恢复吗？',
    successCallBack: async () => {
      tagClass.status = true;
      const result = await tagServer.updateTagClass(tagClass);
      if (result && result.status) {
        emits('updateTagDataCompleted')
      } else {
        ElMessage.error(result.msg);
      }
    },
    failCallBack: () => {
      //console.log('取消恢复')
    },
  })
}
const createTagHandle = (tagClassId: string) => {
  const tag: I_tag = {
    id: '',
    name: '',
    tagClass_id: tagClassId,
    keyWords: '',
    hot: 0,
    sort: 0,
    status: true,
  }
  tagFormItemDialogRef.value?.open('add', tag)
}
const editTagHandle = (tag: I_tag) => {
  tagFormItemDialogRef.value?.open('edit', tag)
}
const deleteTagHandle = (tag: I_tag) => {
  messageBoxConfirm({
    text: '确定要删除吗？',
    successCallBack: async () => {
      tag.status = false;
      const result = await tagServer.updateTag(tag);
      if (result && result.status) {
        emits('updateTagDataCompleted')
      } else {
        ElMessage.error(result.msg);
      }
    },
    failCallBack: () => {
      //console.log('取消删除')
    },
  })
}
const restoreTagHandle = async (tag: I_tag) => {
  messageBoxConfirm({
    text: '确定要恢复吗？',
    successCallBack: async () => {
      tag.status = true;
      const result = await tagServer.updateTag(tag);
      if (result && result.status) {
        emits('updateTagDataCompleted')
      } else {
        ElMessage.error(result.msg);
      }
    },
    failCallBack: () => {
      //console.log('取消删除')
    },
  })
}


onMounted(async () => {
  await init()
})

defineExpose({ init })
</script>
<style lang="scss" scoped>
.tag-list {
  color: #a8abb2;
  height: 100%;

  .disable {
    .tag-list-body {
      background-color: rgba(255, 255, 255, 0.3);
    }
  }

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

      .tag-list-draggable {
        display: flex;
        flex-wrap: wrap;
        gap: 0.5em;
      }
    }
  }
}
</style>
