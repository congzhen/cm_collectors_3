/**
 * 从给定的文件路径中提取文件名
 * @param path 文件的完整路径
 * @param withExtension 是否包含扩展名，true表示包含，false表示不包含
 * @returns 文件名
 */
export function getFileNameFromPath(path: string, withExtension: boolean = true): string {
  // 兼容不同系统的路径分隔符，提取文件名部分
  const filename = path.replace(/\\/g, '/').split('/').pop() || '';

  if (withExtension) {
    return filename;
  }

  // 移除扩展名
  const lastDotIndex = filename.lastIndexOf('.');
  if (lastDotIndex <= 0) {
    // 没有扩展名或以点号开头的文件（隐藏文件）
    return filename;
  }

  return filename.substring(0, lastDotIndex);
}
/**
 * 从给定的文件路径中提取目录路径
 * @param path 文件的完整路径
 * @returns 目录路径
 */
export function getDirectoryFromPath(path: string): string {
  // 兼容不同系统的路径分隔符
  const normalizedPath = path.replace(/\\/g, '/');

  // 找到最后一个路径分隔符的位置
  const lastSeparatorIndex = normalizedPath.lastIndexOf('/');

  // 如果没有找到分隔符，返回空字符串（表示当前目录）
  if (lastSeparatorIndex === -1) {
    return '';
  }

  // 返回目录部分（不包括最后的路径分隔符）
  return normalizedPath.substring(0, lastSeparatorIndex);
}
