export const format = (date: Date, format: string, isUTC = false): string => {
  // 检测日期是否为无效值
  if (date.getTime() < 0 || date.getTime() === new Date(0).getTime()) {
    return '';
  }
  const year = isUTC ? date.getUTCFullYear() : date.getFullYear();
  const month = ('0' + (isUTC ? date.getUTCMonth() + 1 : date.getMonth() + 1)).slice(-2);
  const day = ('0' + (isUTC ? date.getUTCDate() : date.getDate())).slice(-2);
  const hour = ('0' + (isUTC ? date.getUTCHours() : date.getHours())).slice(-2);
  const minute = ('0' + (isUTC ? date.getUTCMinutes() : date.getMinutes())).slice(-2);
  const second = ('0' + (isUTC ? date.getUTCSeconds() : date.getSeconds())).slice(-2);

  return format
    .replace('Y', year.toString())
    .replace('y', year.toString().slice(-2))
    .replace('m', month)
    .replace('d', day)
    .replace('H', hour)
    .replace('i', minute)
    .replace('s', second);
}

// 优化后的函数名和实现
export const formatDate = (date: string | number | Date | null, formatPar = 'Y-m-d', utcFormat = false): string => {
  if (!date) return '';

  let now: Date;
  if (typeof date === 'string') {
    now = new Date(date);
  } else if (typeof date === 'number') {
    // 假设数字是毫秒时间戳
    now = new Date(date);
  } else {
    now = date;
  }

  // 使用 formatStr 避免与函数名冲突
  return format(now, formatPar, utcFormat);
}
