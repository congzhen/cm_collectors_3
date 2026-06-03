import { E_resourceDramaSeriesType } from "@/dataType/app.dataType";
import type { I_resource } from "@/dataType/resource.dataType";

interface ResourceDurationTextOptions {
  // compact 为 true 时只显示总时长，不显示“几个视频 ·”前缀。
  // 移动端短视频封面较窄，使用这个模式可以明显减少角标遮挡面积。
  compact?: boolean;
}

// 将数据库里保存的整数秒格式化成封面角标上的短文本。
// 小于 1 小时显示 mm:ss；达到 1 小时显示 h:mm:ss，尽量避免封面角标过长。
export const formatDurationSeconds = (seconds: number): string => {
  const totalSeconds = Math.floor(seconds);
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const secs = totalSeconds % 60;
  if (hours > 0) {
    return `${hours}:${String(minutes).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
  }
  return `${minutes}:${String(secs).padStart(2, '0')}`;
}

// 根据资源及其分集信息生成封面角标文案。
// 这里只使用接口已经返回的 durationSeconds，不主动请求后端。
// 缺失时长的分集不参与总时长计算，全部未知时返回空字符串，由角标组件隐藏自身。
export const getResourceDurationText = (resource: I_resource, options: ResourceDurationTextOptions = {}): string => {
  // 只有视频类资源显示时长，漫画/图集/普通文件即使意外带有 durationSeconds 也不展示。
  if (resource.mode !== E_resourceDramaSeriesType.Movies && resource.mode !== E_resourceDramaSeriesType.VideoLink) {
    return '';
  }

  // durationSeconds 为 0 表示未知，不等同于真实 0 秒，因此需要过滤掉。
  const knownDurations = resource.dramaSeries
    .map(item => item.durationSeconds || 0)
    .filter(duration => duration > 0);
  if (knownDurations.length === 0) {
    return '';
  }

  const totalDuration = knownDurations.reduce((total, duration) => total + duration, 0);
  const durationText = formatDurationSeconds(totalDuration);
  if (!options.compact && resource.dramaSeries.length > 1) {
    return `${resource.dramaSeries.length}\u4e2a\u89c6\u9891 \u00b7 ${durationText}`;
  }
  return durationText;
}
