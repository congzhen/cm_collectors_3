<template>
  <div class="drawer-search">
    <el-drawer v-model="drawerVisible" :size="props.width">
      <template #header="{ close, titleId, titleClass }">
        <h5 :id="titleId" :class="titleClass">{{ path }}</h5>
      </template>
      <div class="search-body">
        <div class="search-input">
          <el-input v-model="searchQuery" style="width: 60%" :placeholder="sfmLang('searchNoQuery')">
            <template #append>
              <el-button :icon="Search" @click="searchHandle" />
            </template>
          </el-input>
        </div>
        <div class="search-result">
          <el-table ref="fileTableRef" :data="fileEntryDataList" v-loading="loading" height="100%">
            <el-table-column prop="name" :label="sfmLang('name')" min-width="260">
              <template #default="scope">
                <div class="file-name-block">
                  <el-icon v-if="scope.row.is_dir">
                    <Folder />
                  </el-icon>
                  <el-icon v-else>
                    <Document />
                  </el-icon>
                  <label class="file-name">{{ scope.row.name }}</label>
                </div>
                <div class="file-path">
                  {{ scope.row.path }}
                </div>
              </template>
            </el-table-column>
            <el-table-column :label="sfmLang('actions')" fixed="right" width="80px">
              <template #default="scope">
                <el-button-group>
                  <el-button size="small" @click="locationHandle(scope.row)">{{ sfmLang('location') }}</el-button>
                </el-button-group>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>
    </el-drawer>
  </div>
</template>
<script lang="ts" setup>
import { ref, type PropType } from 'vue'
import { Search, Folder, Document } from '@element-plus/icons-vue'
import { sfm_languages } from './lang';
import { sfm_SearchFiles } from './request';
import { message, messageBoxAlert } from './fn';
import { E_LangType, type I_sfm_FileEntry } from './dataType';
const sfmLang = (key: string) => (sfm_languages[props.lang] as Record<string, string>)[key];
const props = defineProps({
  width: {
    type: String,
    default: '680px',
  },
  lang: {
    type: String as PropType<E_LangType>,
    required: true,
  },
})
const emit = defineEmits(['location'])
const drawerVisible = ref(false);
const searchQuery = ref('');
const path = ref('');
const loading = ref(false)

// 文件列表
const fileEntryDataList = ref<I_sfm_FileEntry[]>([])

const loading_open = () => {
  loading.value = true
}
const loading_close = () => {
  loading.value = false
}

const searchHandle = async () => {
  const searchContent = searchQuery.value.trim();
  if (searchContent === '') {
    message(sfmLang('searchNoQuery'), 'error');
    return
  }
  loading_open();
  const result = await sfm_SearchFiles(path.value, searchContent);
  if (result.status) {
    fileEntryDataList.value = result.data;
  } else {
    messageBoxAlert({
      text: result.msg,
      type: 'error',
    })
  }
  loading_close();
}

const locationHandle = (obj: I_sfm_FileEntry) => {
  emit('location', obj)
}

const open = (_path: string) => {
  path.value = _path
  drawerVisible.value = true
}
const close = () => {
  drawerVisible.value = false
}

defineExpose({ open, close })
</script>
<style lang="scss" scoped>
.drawer-search {
  :deep(.el-drawer__header) {
    padding-top: 10px !important;
    margin-bottom: 10px !important;
  }

  :deep(.el-drawer__body) {
    padding: 10px 20px;
  }

  .search-body {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;

    .search-input {
      flex-shrink: 0;
      padding: 5px 0;
    }

    .search-result {
      flex-grow: 1;

      .file-name-block {
        display: flex;
        align-items: center;
        cursor: pointer;

        &:hover {
          color: var(--el-color-primary);
        }

        .el-icon {
          margin-right: 5px;
        }

        .file-name {
          cursor: pointer;

          &:hover {
            color: var(--el-color-primary);
          }
        }
      }

      .file-path {
        font-size: 10px;
        line-height: 12px;
      }
    }
  }
}
</style>
