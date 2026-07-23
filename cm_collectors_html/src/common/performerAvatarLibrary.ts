import { ElMessage, ElMessageBox } from 'element-plus';
import type { I_performerAvatarLibraryStatus } from '@/dataType/performerAvatarLibrary.dataType';
import { performerAvatarLibraryServer } from '@/server/performerAvatarLibrary.server';

// 使用头像库功能前确保数据文件存在。首次使用由用户确认后下载，成功后调用方可继续原操作。
export const ensurePerformerAvatarLibraryReady = async (): Promise<I_performerAvatarLibraryStatus | null> => {
  const statusResult = await performerAvatarLibraryServer.status();
  if (!statusResult.status) {
    ElMessage.error(statusResult.msg);
    return null;
  }
  if (statusResult.data.ready) {
    return statusResult.data;
  }

  try {
    await ElMessageBox.confirm(
      '尚未下载演员头像库数据文件，是否立即下载？',
      '需要下载头像库数据',
      {
        type: 'warning',
        confirmButtonText: '立即下载',
        cancelButtonText: '暂不下载',
        closeOnClickModal: false,
      },
    );
  } catch {
    return null;
  }

  const updateResult = await performerAvatarLibraryServer.updateDataFile();
  if (!updateResult.status) {
    ElMessage.error(updateResult.msg);
    return null;
  }
  ElMessage.success('演员头像库数据文件下载完成');
  return updateResult.data;
};
