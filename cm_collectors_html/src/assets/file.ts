export const saveFile = async (content: string, fileName: string, fileType: { description: string; accept: Record<string, string[]> } = { description: 'All Files', accept: { '*/*': ['.txt'] } }) => {
  // 检查浏览器是否支持 showSaveFilePicker API
  if ('showSaveFilePicker' in window) {
    try {
      // 使用现代文件系统API
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const fileHandle = await (window as any).showSaveFilePicker({
        suggestedName: fileName,
        types: [fileType]
      });

      const writable = await fileHandle.createWritable();
      await writable.write(content);
      await writable.close();
      return;
    } catch (err) {
      // 用户取消保存或出现错误，回退到传统方法
      console.warn('Failed to use showSaveFilePicker, falling back to traditional method.', err);
    }
  }

  // 传统的下载方式（兼容所有浏览器）
  const blob = new Blob([content], { type: Object.keys(fileType.accept)[0] || 'text/plain' });
  const a = document.createElement('a');
  a.download = fileName;
  a.href = URL.createObjectURL(blob);
  a.click();
  URL.revokeObjectURL(a.href);
}

export const loadFile = async (fileTypes: { description: string; accept: Record<string, string[]> }[] = []) => {
  try {
    // 检查浏览器是否支持文件选择器 API
    if ('showOpenFilePicker' in window) {
      // 使用现代文件系统 API
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const options: any = {
        multiple: false
      };

      // 只有当提供了fileTypes时才设置types选项
      if (fileTypes.length > 0) {
        options.types = fileTypes;
      }

      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const fileHandles = await (window as any).showOpenFilePicker(options);

      const file = await fileHandles[0].getFile();
      const content = await file.text();
      return content;
    } else {
      // 降级到传统 input 方法
      return new Promise<string>((resolve, reject) => {
        const input = document.createElement('input');
        input.type = 'file';

        // 设置accept属性
        if (fileTypes.length > 0) {
          const acceptExtensions = Object.values(fileTypes[0]?.accept || {}).flat();
          if (acceptExtensions.length > 0) {
            input.accept = acceptExtensions.join(',');
          }
        }

        input.onchange = async (event: Event) => {
          try {
            const target = event.target as HTMLInputElement;
            const file = target.files?.[0];
            if (file) {
              const content = await file.text();
              resolve(content);
            }
          } catch (error) {
            reject(error);
          }
        };

        input.onerror = (error) => {
          reject(error);
        };

        input.click();
      });
    }
  } catch (error) {
    console.error('Error:', error);
    throw error;
  }
}

// 便捷函数：保存文本文件
export const saveTextFile = async (content: string, fileName: string) => {
  return saveFile(content, fileName, {
    description: 'Text File',
    accept: {
      'text/plain': ['.txt']
    }
  });
};

// 便捷函数：加载文本文件
export const loadTextFile = async () => {
  return loadFile([{
    description: 'Text File',
    accept: {
      'text/plain': ['.txt']
    }
  }]);
};

// 便捷函数：保存JSON文件
export const saveJsonFile = async (content: string, fileName: string) => {
  return saveFile(content, fileName, {
    description: 'JSON File',
    accept: {
      'application/json': ['.json']
    }
  });
};

// 便捷函数：加载JSON文件
export const loadJsonFile = async () => {
  return loadFile([{
    description: 'JSON File',
    accept: {
      'application/json': ['.json']
    }
  }]);
};
