<template>
  <div class="server-file-management">
    <div class="address-bar">
      <div class="address-bar-button-group">
        <el-button-group>
          <el-button :icon="Back" @click="clAddressBarButtonHandle('Back')" :disabled="currentIndex === 0" />
          <el-button :icon="Right" @click="clAddressBarButtonHandle('Right')"
            :disabled="currentIndex === history.length - 1" />
          <el-button :icon="Top" @click="clAddressBarButtonHandle('Top')" :disabled="pathSlc.length === 0" />
          <el-button :icon="Refresh" @click="clAddressBarButtonHandle('Refresh')" />
        </el-button-group>
      </div>
      <div class="address-bar-breadcrumb">
        <el-breadcrumb :separator-icon="ArrowRight">
          <el-breadcrumb-item @click="clAddressBarHandle(-1)">
            <el-link>
              <el-icon size="16">
                <HomeFilled />
              </el-icon>
            </el-link>
          </el-breadcrumb-item>
          <el-breadcrumb-item v-for="name, key in pathSlc" :key="key" @click="clAddressBarHandle(key)">
            <el-link> {{ name }} </el-link>
          </el-breadcrumb-item>
        </el-breadcrumb>
      </div>
    </div>
    <div class="tool-bar">
      <el-button-group>
        <el-button v-if="props.toolBar.includes(E_sfm_ToolBar.CreateFile)" @click="dialogTool(E_sfm_ToolBar.CreateFile)"
          :disabled="isRootPath_C">
          {{ sfmLang('createFile') }}
        </el-button>
        <el-button v-if="props.toolBar.includes(E_sfm_ToolBar.CreateFolder)"
          @click="dialogTool(E_sfm_ToolBar.CreateFolder)" :disabled="isRootPath_C">
          {{ sfmLang('createFolder') }}
        </el-button>
        <el-button v-if="props.toolBar.includes(E_sfm_ToolBar.UploadFile)" @click="dialogTool(E_sfm_ToolBar.UploadFile)"
          :disabled="isRootPath_C">
          {{ sfmLang('upload') }}
        </el-button>
        <el-button v-if="props.toolBar.includes(E_sfm_ToolBar.Download)" @click="dialogTool(E_sfm_ToolBar.Download)"
          :disabled="isRootPath_C">
          {{ sfmLang('download') }}
        </el-button>
        <el-button v-if="props.toolBar.includes(E_sfm_ToolBar.Search)" @click="searchHandle" :disabled="isRootPath_C">
          {{ sfmLang('search') }}
        </el-button>
        <el-button v-if="props.toolBar.includes(E_sfm_ToolBar.Copy)" @click="dialogTool(E_sfm_ToolBar.Copy)"
          :disabled="isRootPath_C">
          {{ sfmLang('copy') }}
        </el-button>
        <el-button v-if="props.toolBar.includes(E_sfm_ToolBar.Move)" @click="dialogTool(E_sfm_ToolBar.Move)"
          :disabled="isRootPath_C">
          {{ sfmLang('move') }}
        </el-button>
        <el-button v-if="filesAction.files.length > 0" @click="dialogTool(E_sfm_ToolBar.Paste)"
          :disabled="isRootPath_C">
          {{ sfmLang('paste') }}
        </el-button>
        <el-button v-if="props.toolBar.includes(E_sfm_ToolBar.Delete)" @click="dialogTool(E_sfm_ToolBar.Compress)"
          :disabled="isRootPath_C">
          {{ sfmLang('compress') }}
        </el-button>
        <el-button v-if="props.toolBar.includes(E_sfm_ToolBar.Permissions)"
          @click="dialogTool(E_sfm_ToolBar.Permissions)" :disabled="isRootPath_C">
          {{ sfmLang('permissions') }}
        </el-button>
        <el-button v-if="props.toolBar.includes(E_sfm_ToolBar.Delete)" @click="dialogTool(E_sfm_ToolBar.Delete)"
          :disabled="isRootPath_C">
          {{ sfmLang('delete') }}
        </el-button>
      </el-button-group>
    </div>
    <div class="file-container">
      <el-table ref="fileTableRef" :data="fileEntryDataList_C" v-loading="loading" height="100%">
        <el-table-column type="selection" width="55" />
        <el-table-column prop="name" :label="sfmLang('name')" min-width="260" show-overflow-tooltip
          v-if="props.column.includes(E_sfm_Column.Name)" sortable :sort-method="sortByName">
          <template #default="scope">
            <div class="file-name-block" @click="clFileHandle(scope.row)">
              <el-icon v-if="scope.row.is_dir">
                <Folder />
              </el-icon>
              <el-icon v-else>
                <Document />
              </el-icon>
              <label class="file-name">{{ scope.row.name }}</label>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="permissions" :label="sfmLang('permissions')" width="100"
          v-if="props.column.includes(E_sfm_Column.Permissions)" />
        <el-table-column prop="size" :label="sfmLang('size')" width="120"
          v-if="props.column.includes(E_sfm_Column.Size)">
          <template #default="scope">
            {{ sizeFormat(scope.row.size) }}
          </template>
        </el-table-column>
        <el-table-column prop="modified_at" :label="sfmLang('modifiedAt')" width="180" sortable
          :sort-method="sortByModifiedAt" v-if="props.column.includes(E_sfm_Column.ModifiedAt)">
          <template #default="scope">
            {{ dateFormat(scope.row.modified_at, 'Y-m-d H:i:s') }}
          </template>
        </el-table-column>
        <el-table-column :label="sfmLang('actions')" fixed="right" :width="operateWidth_C"
          v-if="props.column.includes(E_sfm_Column.Operate)">
          <template #default="scope">
            <div class="file-operate">
              <el-button-group>
                <el-button size="small" v-if="props.fileOperate.includes(E_sfm_FileOperate.Select)"
                  @click="selectFileHandle(scope.row)" :disabled="isRootPath_C">
                  {{ sfmLang('select') }}
                </el-button>
                <el-button size="small" v-if="props.fileOperate.includes(E_sfm_FileOperate.Open)"
                  @click="emit('openFile', scope.row)" :disabled="isRootPath_C || !canOpenFile(scope.row)">
                  {{ sfmLang('open') }}
                </el-button>
                <el-button size="small" v-if="props.fileOperate.includes(E_sfm_FileOperate.Download)"
                  @click="dialogTool(E_sfm_ToolBar.Download, [(scope.row as I_sfm_FileEntry)])"
                  :disabled="isRootPath_C">
                  {{ sfmLang('download') }}
                </el-button>
                <el-button size="small" v-if="props.fileOperate.includes(E_sfm_FileOperate.Extract)"
                  @click="unCompressFile(scope.row)" :disabled="isRootPath_C">
                  {{ sfmLang('extract') }}
                </el-button>
                <el-button size="small" v-if="props.fileOperate.includes(E_sfm_FileOperate.Permissions)"
                  @click="dialogTool(E_sfm_ToolBar.Permissions, [(scope.row as I_sfm_FileEntry)])"
                  :disabled="isRootPath_C">
                  {{ sfmLang('permissions') }}
                </el-button>
                <el-button size="small" v-if="props.fileOperate.includes(E_sfm_FileOperate.Rename)"
                  @click="renameFile(scope.row)" :disabled="isRootPath_C">
                  {{ sfmLang('rename') }}
                </el-button>
                <el-button size="small" v-if="props.fileOperate.includes(E_sfm_FileOperate.Delete)"
                  @click="dialogTool(E_sfm_ToolBar.Delete, [(scope.row as I_sfm_FileEntry)])" :disabled="isRootPath_C">
                  {{ sfmLang('delete') }}
                </el-button>
              </el-button-group>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>
    <dialogFile ref="dialogFileRef" :lang="props.lang" @success="refresh" />
    <dialogFolder ref="dialogFolderRef" :lang="props.lang" @success="refresh" />
    <dialogUpload ref="dialogUploadRef" :lang="props.lang" :timeout="props.uploadTimeout" @success="refresh" />
    <dialogPermissions ref="dialogPermissionsRef" :lang="props.lang" @success="refresh" />
    <dialogRename ref="dialogRenameRef" :lang="props.lang" @success="refresh" />
    <drawerSearch ref="drawerSearchRef" :lang="props.lang" @location="location"></drawerSearch>
  </div>
</template>
<script setup lang="ts">
import { computed, ref, onMounted, type PropType } from 'vue'
import { E_sfm_ToolBar, E_sfm_FileOperate, type I_sfm_FileEntry, E_sfm_FileType, type I_sfm_FilesAction, imageExtensions, textExtensions, codeExtensions, E_sfm_Column, E_LangType } from './com/dataType'
import { dateFormat, generateRandomNumber, getFileExtension, type IMessageBox, message, messageBoxAlert, messageBoxConfirm, messageBoxPrompt, pathToArray, sizeFormat } from './com/fn'
import { Back, Right, Top, Refresh, ArrowRight, HomeFilled, Folder, Document } from '@element-plus/icons-vue'
import { sfm_GetPathDir, sfm_DownloadFile, sfm_PasteCopy, sfm_PasteMove, type IResponse, sfm_DeleteFile, sfm_CompressFile, sfm_UnCompressFile } from './com/request'
import dialogFile from './com/dialog.file.vue'
import dialogFolder from './com/dialog.folder.vue'
import dialogUpload from './com/dialog.upload.vue'
import dialogPermissions from './com/dialog.permissions.vue'
import dialogRename from './com/dialog.rename.vue'
import drawerSearch from './com/drawer.search.vue'

import { sfm_languages } from './com/lang'
import { ElTable } from 'element-plus'
const sfmLang = (key: string) => (sfm_languages[props.lang] as Record<string, string>)[key];

const props = defineProps({
  lang: {
    type: String as PropType<E_LangType>,
    default: 'zhCn',
  },
  uploadTimeout: { // 上传超时时间
    type: Number,
    default: 300000 // 默认300秒
  },
  show: { // 显示的文件类型
    type: Array<E_sfm_FileType>,
    default: () => {
      return [E_sfm_FileType.Directory, E_sfm_FileType.File]
    }
  },
  column: {
    // 栏目
    type: Array<E_sfm_Column>,
    default: () => {
      return [
        E_sfm_Column.Name,
        E_sfm_Column.Permissions,
        E_sfm_Column.Size,
        E_sfm_Column.ModifiedAt,
        E_sfm_Column.Operate,
      ]
    },
  },
  toolBar: {
    // 工具栏按钮
    // 创建文件、创建文件夹、上传文件、下载、复制、移动、压缩、权限、删除
    type: Array<E_sfm_ToolBar>,
    default: () => {
      return [
        E_sfm_ToolBar.CreateFile,
        E_sfm_ToolBar.CreateFolder,
        E_sfm_ToolBar.UploadFile,
        E_sfm_ToolBar.Download,
        E_sfm_ToolBar.Search,
        E_sfm_ToolBar.Copy,
        E_sfm_ToolBar.Move,
        E_sfm_ToolBar.Compress,
        E_sfm_ToolBar.Permissions,
        E_sfm_ToolBar.Delete,
      ]
    },
  },
  fileOperate: { // 文件操作按钮
    // 下载、权限、解压、重命名、删除
    type: Array<E_sfm_FileOperate>,
    default: () => {
      return [
        E_sfm_FileOperate.Download,
        E_sfm_FileOperate.Permissions,
        E_sfm_FileOperate.Extract,
        E_sfm_FileOperate.Rename,
        E_sfm_FileOperate.Delete,
      ]
    },
  },
  deleteValidation: { // 是否需要删除验证
    type: Boolean,
    default: true,
  }
})

const emit = defineEmits(['openFile', 'selectFile'])

const dialogFileRef = ref<InstanceType<typeof dialogFile>>()
const dialogFolderRef = ref<InstanceType<typeof dialogFolder>>()
const dialogUploadRef = ref<InstanceType<typeof dialogUpload>>()
const dialogPermissionsRef = ref<InstanceType<typeof dialogPermissions>>()
const dialogRenameRef = ref<InstanceType<typeof dialogRename>>()
const drawerSearchRef = ref<InstanceType<typeof drawerSearch>>()

const loading = ref(false)
const loading_open = () => {
  loading.value = true
}
const loading_close = () => {
  loading.value = false
}

// 表格引用
const fileTableRef = ref<typeof ElTable>()
// 历史记录
const history = ref<string[][]>([[]])
// 当前历史指针
const currentIndex = ref(0)
//当前地址
const pathSlc = ref<string[]>([])

// 操作栏宽度
const operateWidth_C = computed(() => {
  return props.fileOperate.length * 50 + 20
})

// 是否根目录
const isRootPath_C = computed(() => {
  return pathSlc.value.length === 0
})

const fileEntryDataList_C = computed(() => {
  return fileEntryDataList.value.filter(item => {
    return props.show.includes(item.type)
  })
})

// 文件列表
const fileEntryDataList = ref<I_sfm_FileEntry[]>([])

// 复制与移动
const filesAction = ref<I_sfm_FilesAction>({
  type: E_sfm_ToolBar.Copy,
  files: [],
})

// 更新历史记录
const updateHistory = (newPath: string[]) => {
  // 如果在历史中间位置进行新操作，需要截断后续历史
  if (currentIndex.value < history.value.length - 1) {
    history.value = history.value.slice(0, currentIndex.value + 1)
  }
  history.value.push(newPath)
  currentIndex.value++
}

// 点击地址栏按钮
const clAddressBarButtonHandle = (btnType: 'Back' | 'Right' | 'Top' | 'Refresh') => {
  switch (btnType) {
    case 'Back':
      if (currentIndex.value > 0) {
        currentIndex.value--
        pathSlc.value = history.value[currentIndex.value]
        getPathDir()
      }
      break;
    case 'Right':
      if (currentIndex.value < history.value.length - 1) {
        currentIndex.value++
        pathSlc.value = history.value[currentIndex.value]
        getPathDir()
      }
      break;
    case 'Top':
      clAddressBarHandle(pathSlc.value.length - 2)
      break;
    case 'Refresh':
      refresh();
      break;
  }
}
// 刷新
const refresh = () => {
  getPathDir(false);
}

//定位
const location = (row: I_sfm_FileEntry) => {
  const pathArr = pathToArray(row.path)
  const newPath = pathArr.slice(0, pathArr.length - 1)
  pathSlc.value = newPath
  updateHistory(newPath)
  getPathDir();
}

// 点击地址栏
const clAddressBarHandle = (index: number) => {
  const newPath = pathSlc.value.slice(0, index + 1)
  pathSlc.value = newPath
  updateHistory(newPath)
  getPathDir();
}

// 点击文件
const clFileHandle = async (row: I_sfm_FileEntry) => {
  if (row.is_dir) {
    const newPath = [...pathSlc.value, row.name]
    pathSlc.value = newPath
    const status = await getPathDir();
    if (!status) {
      pathSlc.value = pathSlc.value.slice(0, pathSlc.value.length - 1)
      return;
    }
    updateHistory(newPath)
  }
}

// 获取路径
const getPath = (): string => {
  if (pathSlc.value.length === 0) {
    return "";
  }
  return pathSlc.value[0] == '/' ? '/' : '' + pathSlc.value.join("/");
}
// 获取文件路径
const getFilePath = (file: I_sfm_FileEntry): string => {
  return file.path;
}

// 获取路径下的文件列表
const getPathDir = async (scroll = true) => {
  loading_open();
  const result = await sfm_GetPathDir(getPath());
  if (!result.status) {
    messageBoxAlert({
      text: result.msg,
      type: 'error',
    })
    loading_close();
    return false;
  }
  /*
  result.data.map(file => {
    file.path = `${getPath()}/${file.name}`;
  });
  */
  fileEntryDataList.value = result.data;
  if (scroll) {
    fileTableRef.value?.setScrollTop(0);
  }
  loading_close();
  return true;
}



//获取选中的文件
const getSelectedFiles = (): I_sfm_FileEntry[] => {
  const selectedRows = fileTableRef.value?.getSelectionRows() || [];
  return selectedRows;
}

//获取选中文件的完整路径数组
const getSelectedFilesCompletePath = (): string[] => {
  const selectedRows = getSelectedFiles();
  return selectedRows.map(file => file.path);
}

// 判断文件是否可以打开
const canOpenFile = (file: I_sfm_FileEntry): boolean => {
  // 文件夹不可打开
  if (file.is_dir) return false;

  const ext = getFileExtension(file.name);

  return [
    ...imageExtensions,
    ...textExtensions,
    // ...documentExtensions,
    ...codeExtensions,
  ].includes(ext);
}

const selectFileHandle = (obj: I_sfm_FileEntry) => {
  emit('selectFile', obj)
}

const searchHandle = () => {
  drawerSearchRef.value?.open(getPath());
}


const dialogTool = async (type: E_sfm_ToolBar, dataList: I_sfm_FileEntry[] = []) => {
  const nowPath = getPath();
  if (nowPath == '') {
    messageBoxAlert({
      text: sfmLang('pathOperationError'),
      type: 'error',
      ok: sfmLang('confirm'),
    })
    return;
  }
  switch (type) {
    case E_sfm_ToolBar.CreateFile:
      dialogFileRef.value?.open(nowPath);
      break;
    case E_sfm_ToolBar.CreateFolder:
      dialogFolderRef.value?.open(nowPath);
      break;
    case E_sfm_ToolBar.UploadFile:
      dialogUploadRef.value?.open(nowPath);
      break;
    case E_sfm_ToolBar.Download:
      {
        let selectedRows;
        if (dataList.length === 0) {
          selectedRows = getSelectedFiles();
        } else {
          selectedRows = dataList;
        }
        const downloadQueue: Promise<void>[] = [];
        if (selectedRows.length === 0) {
          noSelectFilesMessageAlert();
          return;
        }
        for (const file of selectedRows) {
          if (!file.is_dir) {
            const path = getFilePath(file);
            downloadQueue.push(new Promise<void>((resolve) => {
              sfm_DownloadFile(path);
              resolve();
            }));
            // 添加 100ms 延迟避免浏览器阻塞
            await new Promise(resolve => setTimeout(resolve, 100));
          }
        }
        await Promise.all(downloadQueue);
      }
      break;
    case E_sfm_ToolBar.Copy:
      {
        const files = getSelectedFilesCompletePath();
        if (files.length === 0) {
          noSelectFilesMessageAlert();
          return;
        }
        filesAction.value = {
          type: E_sfm_ToolBar.Copy,
          files: files,
        }
      }
      break;
    case E_sfm_ToolBar.Move:
      {
        const files = getSelectedFilesCompletePath();
        if (files.length === 0) {
          noSelectFilesMessageAlert();
          return;
        }
        filesAction.value = {
          type: E_sfm_ToolBar.Move,
          files: files,
        }
      }
      break;
    case E_sfm_ToolBar.Paste:
      {
        let result: IResponse<boolean>;
        loading_open();
        if (filesAction.value.type === E_sfm_ToolBar.Copy) {
          result = await sfm_PasteCopy(getPath(), filesAction.value.files);
        } else {
          result = await sfm_PasteMove(getPath(), filesAction.value.files);
        }
        if (result.status) {
          filesAction.value.files = [];
          refresh();
          message(sfmLang('pasteSuccess'), 'success');
        } else {
          messageBoxAlert({
            text: result.msg,
            type: 'error',
            ok: sfmLang('confirm'),
          })
        }
        loading_close();
      }
      break;
    case E_sfm_ToolBar.Compress:
      {
        const nowPath = getPath();
        const files = getSelectedFilesCompletePath();
        if (files.length === 0) {
          noSelectFilesMessageAlert();
          return;
        }
        loading_open();
        const result = await sfm_CompressFile(nowPath, files);
        if (result.status) {
          filesAction.value.files = [];
          refresh();
          message(sfmLang('compressSuccess'), 'success');
        } else {
          messageBoxAlert({
            text: result.msg,
            type: 'error',
            ok: sfmLang('confirm'),
          })
        }
        loading_close();
      }
      break;
    case E_sfm_ToolBar.Permissions:
      {
        let files: string[];
        if (dataList.length === 0) {
          files = getSelectedFilesCompletePath();
        } else {
          files = dataList.map(file => file.path);
        }
        if (files.length === 0) {
          noSelectFilesMessageAlert();
          return;
        }
        dialogPermissionsRef.value?.open(files);
      }
      break;
    case E_sfm_ToolBar.Delete:
      {
        deleteFile(dataList);
      }
  }
}

// 删除
const deleteFile = (dataList: I_sfm_FileEntry[] = []) => {
  const nowPath = getPath();
  let files: string[];
  if (dataList.length === 0) {
    files = getSelectedFilesCompletePath();
  } else {
    files = dataList.map(file => file.path);
  }
  if (files.length === 0) {
    noSelectFilesMessageAlert();
    return;
  }

  const boxMsg: IMessageBox = {
    title: sfmLang('warning'),
    text: sfmLang('confirmDelete'),
    type: 'warning',
    ok: sfmLang('confirm'),
    cancel: sfmLang('cancel'),
    successCallBack: async () => {
      deleteFile_exec(nowPath, files)
    }
  }
  if (props.deleteValidation) {
    const n = generateRandomNumber(4);
    const reg = new RegExp(`^${n}$`);
    boxMsg.text = sfmLang('deleteCode') + ` : ${n}`;
    messageBoxPrompt(boxMsg, reg, sfmLang('inputError'));
  } else {
    messageBoxConfirm(boxMsg)
  }
}
const deleteFile_exec = async (nowPath: string, files: string[]) => {
  loading_open();
  const result = await sfm_DeleteFile(nowPath, files);
  if (result.status) {
    filesAction.value.files = [];
    refresh();
    message(sfmLang('deleteSuccess'), 'success');
  } else {
    messageBoxAlert({
      text: result.msg,
      type: 'error',
      ok: sfmLang('confirm'),
    })
  }
  loading_close();
}

// 重命名
const renameFile = (fileEntry: I_sfm_FileEntry) => {
  dialogRenameRef.value?.open(fileEntry.name, getFilePath(fileEntry));
}

// 解压
const unCompressFile = async (fileEntry: I_sfm_FileEntry) => {
  loading_open();
  const result = await sfm_UnCompressFile(getFilePath(fileEntry));
  if (result.status) {
    refresh();
    message(sfmLang('uncompressSuccess'), 'success');
  } else {
    messageBoxAlert({
      text: result.msg,
      type: 'error',
      ok: sfmLang('confirm'),
    })
  }
  loading_close();
}



//排序方法
const sortByName = (a: I_sfm_FileEntry, b: I_sfm_FileEntry) => {
  // 文件夹优先，然后按名称排序
  if (a.is_dir && !b.is_dir) return -1;
  if (!a.is_dir && b.is_dir) return 1;
  return a.name.localeCompare(b.name);
};
//排序方法
const sortByModifiedAt = (a: I_sfm_FileEntry, b: I_sfm_FileEntry) => {
  // 按修改时间排序
  return new Date(a.modified_at).getTime() - new Date(b.modified_at).getTime();
};

// 没有选择文件
const noSelectFilesMessageAlert = () => {
  message(sfmLang('selectFiles'), 'error')
}

onMounted(() => {
  getPathDir();
})


defineExpose({ getSelectedFiles })
</script>
<style lang="scss" scoped>
.server-file-management {
  width: calc(100% - 20px);
  height: calc(100% - 10px);
  font-size: 12px;
  padding: 5px 10px;
  display: flex;
  flex-direction: column;
  overflow: hidden;

  .address-bar {
    display: flex;
    flex-shrink: 0;
    gap: 10px;
    padding-bottom: 5px;

    .address-bar-button-group {
      flex-shrink: 0;
    }

    .address-bar-breadcrumb {
      width: 100%;
      border: 1px solid #dcdfe6;
      border-radius: 5px;
      padding: 6px 10px;
    }
  }

  .tool-bar {
    display: flex;
    flex-shrink: 0;
    flex-wrap: wrap;
    padding-bottom: 5px;
  }

  .file-container {
    flex: 1;
    border-top: 1px solid var(--el-border-color-lighter);
    overflow: hidden;

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

    .file-operate {
      width: 100%;
      display: flex;

      .el-button-group {
        flex-shrink: 0;
      }
    }
  }

  @at-root {
    :root {
      --vt-c-white: #ffffff;
      --vt-c-white-soft: #333333;
      --vt-c-black: #181818;
      --vt-c-black-soft: #222222;
      --vt-c-divider-light-1: rgba(255, 255, 255, 0.16);
      --vt-c-divider-light-2: rgba(255, 255, 255, 0.08);
    }
  }

  // 暗黑主题专属样式
  &.dark {
    --color-background: var(--vt-c-black);
    --color-background-soft: var(--vt-c-black-soft);
    --color-border: var(--vt-c-divider-light-2);
    --color-text: var(--vt-c-white);
  }
}
</style>
