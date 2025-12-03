import { getFileNameFromPath } from "@/assets/path";
import { getPlayVideoURL } from "@/common/play";
import type { T_VideoPlayMode } from "@/dataType/app.dataType";
import { resourcesDramaSeriesServer } from "@/server/resource.server";

export const playCloud = async (dramaSeriesId: string, mode: T_VideoPlayMode) => {

  const result = await resourcesDramaSeriesServer.infoById(dramaSeriesId);
  let fileName = 'v'
  if (result && result.status && result.data && result.data.src != '') {
    fileName = getFileNameFromPath(result.data.src, false)
  }
  console.log(fileName);

  // 获取当前服务器地址
  const serverAddress = window.location.origin;
  const path = getPlayVideoURL(dramaSeriesId, mode, fileName);
  // 创建云播放协议链接
  const url = `cmcollectorsvideoplay://${serverAddress}${path}?playCloud=true`;
  console.log('尝试打开云播放协议链接:', url);

  try {
    // 检查是否有访问父窗口的权限（即是否在 iframe 中）
    let hasParentAccess = false;
    try {
      hasParentAccess = !!(window.top && window.top !== window.self);
    } catch {
      // 如果访问被拒绝（安全错误），则认为在 iframe 中但无访问权限
      hasParentAccess = false;
    }

    if (hasParentAccess) {
      // 在 iframe 中且可以访问父窗口，通过父窗口打开
      if (window.top) {
        window.top.location.href = url;
      } else {
        window.location.href = url;
      }
    } else {
      // 不在 iframe 中或无法访问父窗口，直接打开
      window.location.href = url;
    }
  } catch (error) {
    console.error('打开云播放器失败:', error);
    alert('无法打开云播放器，请确保已正确安装并配置了相关组件。');
  }
};
