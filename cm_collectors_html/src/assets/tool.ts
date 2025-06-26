/**
 * 从路径字符串中获取最后一个非空的路径片段（文件名或文件夹名）
 * 支持 Windows 和 Unix 风格路径
 */
export const getFinalPathSegment = (src: string): string => {
  if (!src) return ''

  // 统一处理为 Unix 风格路径（兼容 Windows 和 Linux）
  const normalizedPath = src.replace(/\\/g, '/')

  // 分割路径
  const segments = normalizedPath.split('/')

  // 从后往前查找第一个非空片段
  for (let i = segments.length - 1; i >= 0; i--) {
    if (segments[i]) {
      return segments[i]
    }
  }

  return ''
}
