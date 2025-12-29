import { getTopWindow, canUseFilePickerAPI } from "./windows";

export const saveFile = async (content: string, fileName: string, fileType: { description: string; accept: Record<string, string[]> } = { description: 'All Files', accept: { '*/*': ['.txt'] } }) => {
  // 检查是否可以使用文件选择器API
  if (canUseFilePickerAPI()) {
    try {
      // 获取合适的窗口对象
      const targetWindow = getTopWindow();

      // 使用现代文件系统API
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const fileHandle = await (targetWindow as any).showSaveFilePicker({
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

export const loadFile = async (fileTypes: { description: string; accept: Record<string, string[]> }[] = []): Promise<{ content: string; file: File } | null> => {
  try {
    // 检查是否可以使用文件选择器API
    if (canUseFilePickerAPI()) {
      // 获取合适的窗口对象
      const targetWindow = getTopWindow();

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
      const fileHandles = await (targetWindow as any).showOpenFilePicker(options);

      const file = await fileHandles[0].getFile();
      console.log('file', file);
      const content = await file.text();
      return {
        content,
        file
      };
    }
  } catch (err) {
    // 忽略错误，继续使用传统方法
    console.warn('Failed to use showOpenFilePicker, falling back to traditional method.', err);
  }

  // 降级到传统 input 方法
  return new Promise<{ content: string; file: File } | null>((resolve, reject) => {
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
          resolve({
            content,
            file
          });
        } else {
          resolve(null);
        }
      } catch (error) {
        reject(error);
      }
    };

    input.onerror = (error) => {
      reject(error);
    };

    document.body.appendChild(input);
    input.click();
    document.body.removeChild(input);
  });
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
export const loadTextFile = async (): Promise<{ content: string; file: File } | null> => {
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
export const loadJsonFile = async (): Promise<{ content: string; file: File } | null> => {
  return loadFile([{
    description: 'JSON File',
    accept: {
      'application/json': ['.json']
    }
  }]);
};
