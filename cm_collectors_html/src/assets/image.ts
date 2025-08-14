/**
 * 将base64字符串转换为File对象
 * @param base64String 包含base64数据的字符串，可能包含数据URL前缀
 * @param filename 生成的文件名
 * @param mimeType 文件的MIME类型，默认为'image/png'
 * @returns 转换后的File对象
 */
export const base64ToFile = (base64String: string, filename: string, mimeType: string = 'image/png'): File => {
  // 去掉base64数据URL的前缀（如data:image/png;base64,）
  const base64Data = base64String.replace(/^data:image\/\w+;base64,/, '');

  // 将base64字符串解码为二进制数据
  const byteString = atob(base64Data);

  // 创建字节数组
  const ab = new ArrayBuffer(byteString.length);
  const ia = new Uint8Array(ab);

  // 将二进制数据填充到字节数组中
  for (let i = 0; i < byteString.length; i++) {
    ia[i] = byteString.charCodeAt(i);
  }

  // 返回File对象
  return new File([ia], filename, { type: mimeType });
}

/**
 * 按指定最大宽度缩放图像（支持OffscreenCanvas和传统canvas回退）
 * @param imageData 图像数据（File对象、Blob或base64字符串）
 * @param maxWidth 最大宽度（像素）
 * @param options 缩放选项
 * @returns 缩放后的图像base64数据
 */
export const scaleImage = async (
  imageData: File | Blob | string,
  maxWidth: number,
  options?: {
    quality?: number;           // 图像质量 (0-1)
    resizeQuality?: 'pixelated' | 'low' | 'medium' | 'high'; // 缩放算法质量
    outputType?: 'image/jpeg' | 'image/png' | 'image/webp';   // 输出格式
  }
): Promise<string> => {
  return new Promise(async (resolve, reject) => {
    try {
      let blob: Blob;

      // 根据输入类型转换为Blob
      if (typeof imageData === 'string') {
        // 如果是base64字符串，转换为Blob
        const file = base64ToFile(imageData, 'temp.jpg');
        blob = file;
      } else {
        // 如果已经是Blob或File对象
        blob = imageData;
      }

      // 获取图像原始尺寸
      const img = new Image();
      const objectUrl = URL.createObjectURL(blob);

      img.onload = async () => {
        // 计算缩放后的尺寸
        const scale = maxWidth / img.width;
        const newWidth = maxWidth;
        const newHeight = img.height * scale;

        // 检查浏览器是否支持OffscreenCanvas
        const useOffscreenCanvas = typeof OffscreenCanvas !== 'undefined';

        if (useOffscreenCanvas) {
          // 使用OffscreenCanvas（现代浏览器）
          try {
            // 使用createImageBitmap进行高质量缩放
            const resizeQuality = options?.resizeQuality || 'high';
            const imageBitmap = await createImageBitmap(blob, {
              resizeWidth: newWidth,
              resizeHeight: newHeight,
              resizeQuality: resizeQuality
            });

            // 使用OffscreenCanvas替代传统canvas元素
            const offscreenCanvas = new OffscreenCanvas(newWidth, newHeight);
            const ctx = offscreenCanvas.getContext('2d');

            if (!ctx) {
              URL.revokeObjectURL(objectUrl);
              imageBitmap.close();
              throw new Error('无法获取OffscreenCanvas上下文');
            }

            // 绘制图像到OffscreenCanvas
            ctx.drawImage(imageBitmap, 0, 0, newWidth, newHeight);

            // 转换为base64
            const outputType = options?.outputType || 'image/jpeg';
            const quality = options?.quality !== undefined ? options.quality : 0.9;

            // 将OffscreenCanvas转换为Blob，再转换为base64
            const canvasBlob = await offscreenCanvas.convertToBlob({
              type: outputType,
              quality: quality
            });

            const reader = new FileReader();
            reader.onload = () => {
              // 清理资源
              URL.revokeObjectURL(objectUrl);
              imageBitmap.close();

              resolve(reader.result as string);
            };
            reader.onerror = () => {
              // 清理资源
              URL.revokeObjectURL(objectUrl);
              imageBitmap.close();

              reject(new Error('Blob读取失败'));
            };
            reader.readAsDataURL(canvasBlob);
            // eslint-disable-next-line @typescript-eslint/no-unused-vars
          } catch (offscreenError) {
            // 如果OffscreenCanvas失败，回退到传统canvas方法
            scaleWithTraditionalCanvas(img, newWidth, newHeight, objectUrl, options, resolve, reject);
          }
        } else {
          // 回退到传统canvas方法（旧版浏览器）
          scaleWithTraditionalCanvas(img, newWidth, newHeight, objectUrl, options, resolve, reject);
        }
      };

      img.onerror = () => {
        URL.revokeObjectURL(objectUrl);
        reject(new Error('图像加载失败'));
      };

      img.src = objectUrl;
    } catch (error) {
      reject(error);
    }
  });
}

/**
 * 使用传统canvas方法缩放图像
 * @param img 已加载的图像对象
 * @param newWidth 新宽度
 * @param newHeight 新高度
 * @param objectUrl 图像对象URL
 * @param options 缩放选项
 * @param resolve Promise resolve函数
 * @param reject Promise reject函数
 */
const scaleWithTraditionalCanvas = (
  img: HTMLImageElement,
  newWidth: number,
  newHeight: number,
  objectUrl: string,
  options: {
    quality?: number;
    outputType?: 'image/jpeg' | 'image/png' | 'image/webp';
  } | undefined,
  resolve: (value: string | PromiseLike<string>) => void,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  reject: (reason?: any) => void
) => {
  try {
    // 创建canvas用于绘制缩放后的图像
    const canvas = document.createElement('canvas');
    const ctx = canvas.getContext('2d');

    if (!ctx) {
      URL.revokeObjectURL(objectUrl);
      reject(new Error('无法获取canvas上下文'));
      return;
    }

    // 设置canvas尺寸
    canvas.width = newWidth;
    canvas.height = newHeight;

    // 在canvas上绘制缩放后的图像
    ctx.drawImage(img, 0, 0, newWidth, newHeight);

    // 将图像转换为base64数据
    const outputType = options?.outputType || 'image/jpeg';
    const quality = options?.quality !== undefined ? options.quality : 0.9;
    const dataUrl = canvas.toDataURL(outputType, quality);

    // 清理资源
    URL.revokeObjectURL(objectUrl);

    resolve(dataUrl);
  } catch (error) {
    URL.revokeObjectURL(objectUrl);
    reject(error);
  }
}

/**
 * 获取图片的宽度和高度
 * @param imageData 图像数据（File对象、Blob或base64字符串）
 * @returns 包含宽度和高度的对象
 */
export const getImageDimensions = async (imageData: File | Blob | string): Promise<{ width: number; height: number }> => {
  return new Promise(async (resolve, reject) => {
    try {
      let blob: Blob;

      // 根据输入类型转换为Blob
      if (typeof imageData === 'string') {
        // 如果是base64字符串，转换为Blob
        const file = base64ToFile(imageData, 'temp.jpg');
        blob = file;
      } else {
        // 如果已经是Blob或File对象
        blob = imageData;
      }

      // 创建图片对象用于获取尺寸
      const img = new Image();
      const objectUrl = URL.createObjectURL(blob);

      // 设置图片加载完成后的处理
      img.onload = () => {
        // 获取图片的实际宽度和高度
        const width = img.width;
        const height = img.height;

        // 释放创建的对象URL
        URL.revokeObjectURL(objectUrl);

        // 返回尺寸信息
        resolve({ width, height });
      };

      img.onerror = () => {
        URL.revokeObjectURL(objectUrl);
        reject(new Error('图像加载失败'));
      };

      // 开始加载图像
      img.src = objectUrl;
    } catch (error) {
      reject(error);
    }
  });
}
