<template>
  <el-dialog v-model="dialogVisible" :width="props.width" :append-to-body="true" title="上传文件"
    :close-on-click-modal="false">
    <div class="upload-container">
      <div class="upload-files-block">
        <div class="upload-btn">
          <el-button-group>
            <el-button @click="triggerFileInput">{{ sfmLang('uploadFile') }}</el-button>
            <el-button @click="selectFolder">{{ sfmLang('uploadFolder') }}</el-button>
          </el-button-group>
        </div>
        <div class="upload-files-block">
          <div class="upload-drag" @dragover="handleDragOver" @dragleave="handleDragLeave" @drop="handleDrop"
            :class="{ 'drag-over': isDragOver }" @click="triggerFileInput">
            <el-icon class="upload-icon">
              <UploadFilled />
            </el-icon>
            <div class="upload-text">
              {{ sfmLang('clickToUpload') }}<br>
              ({{ sfmLang('supportMultiUpload') }})
            </div>
          </div>
        </div>
      </div>
      <div class="upload-files-list-container">
        <ul class="upload-files-list">
          <el-scrollbar height="100%">
            <li v-for="(item, index) in filesWithPath" :key="index" class="file-item">
              <div class="file-path">{{ item.path }}</div>
              <div class="file-name">{{ item.file.name }}
                <el-icon class="delete-btn" @click="deleteFile(index)">
                  <Delete />
                </el-icon>
              </div>
              <el-progress :percentage="item.progress" :format="() => `${item.progress}%`"
                :show-text="item.progress > 0" :stroke-width="8" :color="progressColor(item.progress)" />
              <el-alert v-if="item.message != ''" :title="item.message" type="error" :closable="false" />
            </li>
          </el-scrollbar>
        </ul>
      </div>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogVisible = false"> {{ sfmLang('close') }} </el-button>
        <el-button type="primary" @click="startUpload"> {{ sfmLang('upload') }} </el-button>
      </div>
    </template>
  </el-dialog>
  <input type="file" style="display:none" ref="singleFileInput" @change="handleFileSelect">
</template>
<script lang="ts" setup>
import { UploadFilled, Delete } from '@element-plus/icons-vue'
import { ref, type PropType } from 'vue';
import { message } from './fn';
import { apiList } from './request';
import { sfm_languages } from './lang';
import type { E_LangType } from './dataType';
const sfmLang = (key: string) => (sfm_languages[props.lang] as Record<string, string>)[key];


// 定义带路径的文件类型
interface FileWithPath {
  id: number;       // ID
  file: File;       // 文件对象
  path: string;     // 文件路径
  progress: number; // 上传进度
  message: string;  // 上传信息
  status: 'pending' | 'uploading' | 'success' | 'error'; // 上传状态
}

const props = defineProps({
  width: {
    type: String,
    default: '720px',
  },
  timeout: {
    type: Number,
    default: 300000 // 默认300秒
  },
  lang: {
    type: String as PropType<E_LangType>,
    required: true,
  },
})
const emit = defineEmits(['success'])

const dialogVisible = ref(false)
//上传目录
const uploadPath = ref('');
// 管理拖放状态
const isDragOver = ref(false);
const singleFileInput = ref<HTMLInputElement | null>(null);
const filesWithPath = ref<FileWithPath[]>([]);
//上传文件的id
let fileId = 0;

const init = (_uploadPath: string) => {
  uploadPath.value = _uploadPath;
  dialogVisible.value = false;
  isDragOver.value = false;
  filesWithPath.value = [];
  fileId = 0;
}

const triggerFileInput = () => {
  singleFileInput.value?.click();
};

const createFileWithPath = (file: File, path: string): FileWithPath => {
  const id = fileId++;
  return { id: id, file, path, progress: 0, message: '', status: 'pending' };
}

// 处理单文件选择
const handleFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement;
  if (!target.files) return;

  const file = target.files[0];
  // 假设路径为当前目录（可根据需求修改）
  const newPath = `/${file.name}`;
  // 检查路径是否已存在
  if (!filesWithPath.value.some(f => f.path === newPath)) {
    filesWithPath.value.push(createFileWithPath(file, newPath));
    target.value = ''; // 重置输入框
  } else {
    console.warn('文件路径已存在:', newPath);
  }
};

const selectFolder = async () => {
  try {
    if (!('showDirectoryPicker' in window)) {
      console.error('不支持 File System Access API');
      return;
    }

    const directoryHandle = await (window as any).showDirectoryPicker();
    if (!directoryHandle) return;

    // 获取所有文件及其路径
    const allFilesWithPaths = await traverseDirectory(directoryHandle);
    // 根据路径去重
    const uniqueFiles = allFilesWithPaths.filter(file =>
      !filesWithPath.value.some(existing => existing.path === file.path)
    );
    filesWithPath.value.push(...uniqueFiles);
    console.log('所有文件路径:', allFilesWithPaths);
  } catch (error) {
    console.error('选择目录失败:', error);
  }
};

const traverseDirectory = async (
  handle: FileSystemDirectoryHandle,
  currentPath: string = '',
  rootDirName: string = ''
): Promise<FileWithPath[]> => {
  if (rootDirName == '') {
    rootDirName = handle.name;
  }
  const filesWithPaths: FileWithPath[] = [];
  const entries = await (handle as any).values();
  for await (const entry of entries) {
    if (entry.kind === 'file') {
      const file = await (entry as FileSystemFileHandle).getFile();
      const fullPath = currentPath
        ? `${rootDirName}/${currentPath}/${file.name}`
        : `${rootDirName}/${file.name}`
      filesWithPaths.push(createFileWithPath(file, fullPath));
    } else if (entry.kind === 'directory') {
      const subdirPath = currentPath
        ? `${currentPath}/${entry.name}`
        : entry.name;
      const subdirFiles = await traverseDirectory(
        entry as FileSystemDirectoryHandle,
        subdirPath,
        rootDirName
      );
      filesWithPaths.push(...subdirFiles);
    }
  }
  return filesWithPaths;
};

const deleteFile = (index: number) => {
  filesWithPath.value.splice(index, 1);
};

// 拖拽进入时
const handleDragOver = (e: DragEvent) => {
  e.preventDefault();
  e.dataTransfer!.dropEffect = 'copy';
  isDragOver.value = true;
};

// 拖拽离开时
const handleDragLeave = () => {
  isDragOver.value = false;
};

// 文件释放时
// 修改后的 handleDrop 函数
const handleDrop = async (e: DragEvent) => {
  e.preventDefault();
  isDragOver.value = false;

  if (!e.dataTransfer) return;

  // 处理拖拽的文件和目录
  const items = e.dataTransfer.items;
  const files: File[] = [];
  const directories: FileSystemDirectoryHandle[] = [];

  // 使用 Promise.all 等待所有异步操作完成
  await Promise.all(Array.from(items).map(async (item) => {
    try {
      const handle = await (item as any).getAsFileSystemHandle();
      if (handle.kind === 'directory') {
        directories.push(handle);
      } else {
        const theFile = await handle.getFile();
        files.push(theFile);
      }
    } catch (error) {
      // 处理普通文件
      const file = item.getAsFile();
      if (file) files.push(file);
    }
  }));

  // 处理文件
  files.forEach((file) => {
    const newPath = `/${file.name}`;
    if (!filesWithPath.value.some(f => f.path === newPath)) {
      filesWithPath.value.push(createFileWithPath(file, newPath));
    } else {
      console.warn('文件路径已存在:', newPath);
    }
  });

  // 处理目录（递归遍历子文件）
  for (const dir of directories) {
    try {
      const filesWithPaths = await traverseDirectory(dir);
      const uniqueFiles = filesWithPaths.filter(file =>
        !filesWithPath.value.some(existing => existing.path === file.path)
      );
      filesWithPath.value.push(...uniqueFiles);
    } catch (error) {
      console.error('目录遍历失败:', error);
    }
  }
};

// 进度条颜色逻辑
const progressColor = (progress: number) => {
  if (progress === 100) return '#67C23A'; // 成功绿色
  return '#409EFF'; // 默认蓝色
};

// 开始上传
const startUpload = async () => {
  const concurrency = 3;
  const filesToUpload = filesWithPath.value
    .filter(file => file.status === 'pending') // 过滤待上传状态
    .map(file => uploadFile(file));

  // 分批次执行
  for (let i = 0; i < filesToUpload.length; i += concurrency) {
    const batch = filesToUpload.slice(i, i + concurrency);
    await Promise.all(batch);
  }

  message(sfmLang('uploadCompleted'), 'success');
  emit('success');
};

// 上传单个文件的函数
const uploadFile = async (fileItem: FileWithPath) => {
  const uploadApi = apiList.uploadFile;
  return new Promise<void>((resolve, reject) => {
    const formData = new FormData();
    formData.append('file', fileItem.file);
    formData.append('file_path', fileItem.path);
    formData.append('upload_path', uploadPath.value);

    const xhr = new XMLHttpRequest();
    let timeoutId: any;

    // 设置超时时间
    const timeout = props.timeout;
    timeoutId = setTimeout(() => {
      xhr.abort();
      reject(new Error(`${sfmLang('uploadTimeout')} (${timeout}ms)`));
    }, timeout);

    xhr.open('POST', uploadApi, true);

    // 更新进度
    xhr.upload.onprogress = (event) => {
      if (event.lengthComputable) {
        const progress = Math.round((event.loaded / event.total) * 100);
        const targetFile = filesWithPath.value.find(f => f.id === fileItem.id);
        if (targetFile) {
          targetFile.progress = progress;
          targetFile.status = 'uploading';
        }
      }
    };

    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        const result = JSON.parse(xhr.response);
        if (!result.status) {
          fileItem.message = result.msg;
          fileItem.status = 'error';
        } else {
          fileItem.status = 'success';
        }
        resolve();
      } else {
        fileItem.status = 'error';
        reject(new Error(`${sfmLang('uploadFailed')}: ${xhr.statusText}`));
      }
    };

    xhr.onerror = () => {
      fileItem.status = 'error';
      reject(new Error(sfmLang('uploadError')));
    };

    xhr.send(formData);
  });
};

const open = (_uploadPath: string) => {
  init(_uploadPath);
  dialogVisible.value = true;
}
const close = () => {
  dialogVisible.value = false
}

defineExpose({ open })
</script>
<style lang="scss" scoped>
.upload-container {
  display: flex;
  gap: 10px;

  .upload-files-block {
    flex-shrink: 0;
  }

  .upload-btn {
    margin-bottom: 10px;
    display: flex;
    justify-content: center;
  }

  .upload-drag {
    // 基础样式
    background: linear-gradient(135deg, var(--el-bg-color), var(--el-fill-secondary));
    border: 1px dashed var(--el-border-color);
    border-radius: 8px;
    padding: 40px 20px;
    text-align: center;
    transition: all 0.3s ease;
    cursor: pointer;

    // 添加内阴影提升立体感
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.05);

    // hover 状态增强反馈
    &:hover {
      background: linear-gradient(135deg, var(--el-fill-secondary), var(--el-bg-color));
      border-color: var(--el-color-primary);
      box-shadow: 0 6px 12px rgba(0, 0, 255, 0.15);
    }

    // 拖拽悬停时的视觉反馈
    &.drag-over {
      background: linear-gradient(135deg, var(--el-bg-color), var(--el-fill-secondary-dark));
      border-color: var(--el-color-primary);
      animation: bounce 0.3s ease-in-out;
    }

    // 内容布局
    .upload-icon {
      font-size: 48px;
      color: #409eff;
      margin-bottom: 16px;
      pointer-events: none;
    }

    .upload-text {
      color: #606266;
      font-size: 14px;
      line-height: 1.5;
      pointer-events: none;
    }
  }

  // 添加动画效果
  @keyframes bounce {

    0%,
    100% {
      transform: translateY(0);
    }

    50% {
      transform: translateY(-4px);
    }
  }

  .upload-files-list-container {
    flex: 1;
    overflow: hidden;

    .upload-files-list {
      padding: 10px;
      border-radius: 8px;
      height: 300px;
      overflow-y: auto;
      border: 1px dashed #CDD0D6;

      .file-item {
        display: flex;
        flex-direction: column;
        padding: 8px 12px;
        margin-bottom: 8px;
        background: var(--el-bg-color);
        border-radius: 6px;
        transition: all 0.3s ease;

        &:hover {
          background: var(--el-fill-secondary);
          transform: translateY(-2px);
          box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
        }

        .file-path {
          width: calc(100% - 24px);
          font-size: 12px;
          color: var(--el-text-color-secondary);
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
          margin-bottom: 4px;
        }

        .file-name {
          display: flex;
          justify-content: space-between;
          align-items: center;
          font-size: 14px;
          color: var(--el-text-color-primary);
          margin-top: 4px;

          .delete-btn {
            color: #ff4949;
            cursor: pointer;
            margin-left: 10px;
            transition: transform 0.2s;

            &:hover {
              transform: scale(1.2);
            }
          }
        }

        .el-progress {
          margin-top: 5px;

          // 进度条颜色动态处理
          .el-progress__text {
            color: #606266;
          }

          // 根据进度颜色变化
          .el-progress-bar__inner {
            background-color: var(--el-color-primary);

            &.success {
              background-color: var(--el-color-success);
            }

            &.error {
              background-color: var(--el-color-danger);
            }
          }
        }

        .el-alert {
          :deep(.el-alert__title) {
            font-size: 12px;
            line-height: 16px;
          }

        }

      }
    }
  }
}
</style>
