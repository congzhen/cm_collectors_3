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



/**
 * 生成一个随机的RGBA颜色字符串
 * @returns 返回格式为rgba(r, g, b, a)的随机颜色字符串，其中r、g、b为0-255之间的整数，a为0-1之间保留两位小数的小数
 */
export const getRandomColor = (): string => {
  // 生成0-255之间的随机整数作为RGB颜色值
  const r = Math.floor(Math.random() * 256);
  const g = Math.floor(Math.random() * 256);
  const b = Math.floor(Math.random() * 256);

  // 生成0-1之间的小数作为透明度，保留两位小数
  const a = Math.random().toFixed(2);

  // 将随机生成的RGBA值组合成颜色字符串并返回
  return `rgba(${r}, ${g}, ${b}, ${a})`;
};
