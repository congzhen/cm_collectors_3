import { ElMessage, ElMessageBox } from "element-plus";
/**
 * 获取文件扩展名（不带点号，小写格式）
 * @param path 文件路径
 * @returns 扩展名字符串（无扩展名返回空字符串）
 */
export function getFileExtension(path: string): string {
  // 1. 提取文件名部分（兼容不同系统路径分隔符）
  const filename = path
    .split('/')
    .pop()
    ?.split('\\')
    .pop() || '';

  // 2. 分割扩展名（保留最后一个点后的部分）
  const lastDotIndex = filename.lastIndexOf('.');
  if (lastDotIndex <= 0) return ''; // 无扩展名或隐藏文件

  return filename
    .substring(lastDotIndex + 1)
    .toLowerCase()
    .trim();
}

//根据文件后缀名返回文件mimeType
export function getExtMimeType(ext: string) {
  switch (ext) {
    case 'jpg':
    case 'jpeg':
      return 'image/jpeg';
    case 'png':
      return 'image/png';
    case 'gif':
      return 'image/gif';
    case 'bmp':
      return 'image/bmp';
    case 'webp':
      return 'image/webp';
    case 'tiff':
      return 'image/tiff';
    case 'svg':
      return 'image/svg+xml';
  }
}



export const format = (date: Date, format: string, isUTC = false): string => {
  //检测日期是否为无效值
  if (date.getTime() < 0 || date == new Date(0)) {
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

export const dateFormat = (date: string | number | Date | null, desiredFormat = 'Y-m-d', utcFormat = false): string => {
  if (!date) return '';
  let now: Date;
  if (typeof date === 'string') {
    now = new Date(date);
  } else if (typeof date === 'number') {
    now = new Date(date);
  } else {
    now = date;
  }
  return format(now, desiredFormat, utcFormat);
}


export const sizeFormat = (size: number | string | null, precision = 2): string => {
  if (!size) return '';

  // 统一转换为数字类型
  const numSize = typeof size === 'string' ? parseInt(size, 10) : size;

  // 处理无效数值
  if (isNaN(numSize) || numSize < 0) return '';

  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let index = 0;
  let formattedSize = numSize;

  // 循环计算合适单位（每次除以1024）
  while (formattedSize >= 1024 && index < units.length - 1) {
    formattedSize /= 1024;
    index++;
  }

  // 处理小数精度（B单位时不显示小数）
  const decimalDigits = index === 0 ? 0 : precision;

  return `${formattedSize.toFixed(decimalDigits)} ${units[index]}`;
}

/**
 * 生成指定位数的随机数字
 * @param digits 需要生成的随机数的位数
 * @returns 生成的随机数字（整数）
 */
export const generateRandomNumber = (digits: number): number => {
  if (digits <= 0) {
    throw new Error("The number of digits must be greater than 0");
  }

  const min = Math.pow(10, digits - 1); // 最小值，例如：digits=3 → 100
  const max = Math.pow(10, digits) - 1; // 最大值，例如：digits=3 → 999

  return Math.floor(Math.random() * (max - min + 1)) + min;
}
/**
 * 将路径字符串转换为数组
 * @param path 路径字符串
 * @returns 路径分段数组
 */
export const pathToArray = (path: string): string[] => {
  if (!path) return [];

  // 统一分隔符为正斜杠 '/'
  const normalizedPath = path.replace(/\\/g, '/');

  // 检测是否是 Windows 根路径（如 C:/）
  const windowsRootRegex = /^[a-zA-Z]:\//;
  let segments: string[];

  if (windowsRootRegex.test(normalizedPath)) {
    // 提取 Windows 根路径（如 C:/）并保留
    const root = normalizedPath.substring(0, 3); // 如 "C:/"
    const rest = normalizedPath.slice(3).replace(/^\/+|\/+$/g, ''); // 去除多余的斜杠
    segments = [root, ...rest.split('/')];
  } else if (normalizedPath.startsWith('/')) {
    // 处理 Linux 根路径（以 / 开头）
    const rest = normalizedPath.slice(1).replace(/^\/+|\/+$/g, ''); // 去除开头和结尾的多余斜杠
    segments = ['/', ...rest.split('/')];
  } else {
    // 普通路径处理
    const trimmedPath = normalizedPath.replace(/^\/+|\/+$/g, '');
    segments = trimmedPath.split('/');
  }

  // 过滤掉空字符串（如多余的斜杠导致的空段）
  return segments.filter(segment => segment.trim() !== '');
};

export interface IMessageBox {
  text: string,
  title?: string,
  type?: 'success' | 'info' | 'warning' | 'error',
  successCallBack?: () => void,
  failCallBack?: () => void,
  ok?: string,
  cancel?: string
}
export const message = function (message: string, type: 'success' | 'info' | 'warning' | 'error') {
  ElMessage({
    message,
    type,
  })
}
export const messageBoxAlert = function (obj: IMessageBox) {
  const confirmButtonText = obj.ok ? obj.ok : 'OK';
  ElMessageBox.alert(
    obj.text,
    obj.title,
    {
      confirmButtonText,
      type: obj.type,
    },
  ).then(obj.successCallBack).catch(obj.failCallBack);
}
export const messageBoxConfirm = function (obj: IMessageBox) {
  const confirmButtonText = obj.ok ? obj.ok : 'OK';
  const cancelButtonText = obj.cancel ? obj.cancel : 'Cancel';
  const title = obj.title ? obj.title : 'Warning';
  ElMessageBox.confirm(
    obj.text,
    title,
    {
      confirmButtonText,
      cancelButtonText,
      type: 'warning',
    }
  ).then(() => {
    if (obj.successCallBack) {
      obj.successCallBack();
    }

  }).catch(() => {
    if (obj.failCallBack) {
      obj.failCallBack();
    }
  })
}
export const messageBoxPrompt = function (obj: IMessageBox, inputPattern: RegExp, inputErrorMessage: string) {
  const confirmButtonText = obj.ok ? obj.ok : 'OK';
  const cancelButtonText = obj.cancel ? obj.cancel : 'Cancel';
  const title = obj.title ? obj.title : 'Warning';
  ElMessageBox.prompt(
    obj.text,
    title,
    {
      confirmButtonText,
      cancelButtonText,
      type: 'warning',
      inputPattern,
      inputErrorMessage,
    }
  ).then(() => {
    if (obj.successCallBack) {
      obj.successCallBack();
    }
  }).catch(() => {
    if (obj.failCallBack) {
      obj.failCallBack();
    }
  })
}
