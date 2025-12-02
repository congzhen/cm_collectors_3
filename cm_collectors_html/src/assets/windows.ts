export const getTopWindow = (): Window => {
  try {
    // 尝试访问最顶层窗口
    let targetWindow: Window = window;
    let current: Window = window;

    // 遍历所有可访问的父级窗口，直到到达顶层或遇到跨域限制
    while (current !== current.parent) {
      // 检查是否同源
      if (current.parent.location.host !== current.location.host) {
        break;
      }
      current = current.parent;
      targetWindow = current;
    }

    // 如果top对象可用且同源，则使用top
    if (window.top && window.top !== window) {
      if (window.top.location.host === window.location.host) {
        targetWindow = window.top;
      }
    }

    return targetWindow;
  } catch (e) {
    console.error(e);
    // 如果由于跨域策略无法访问父级窗口，则使用当前窗口
    return window;
  }
};

// 新增函数：检查是否允许使用文件选择器API
export const canUseFilePickerAPI = (): boolean => {
  try {
    // 获取合适的窗口对象
    const targetWindow = getTopWindow();

    // 检查是否支持showOpenFilePicker API
    if (!('showOpenFilePicker' in window)) {
      return false;
    }

    // 检查当前窗口是否与目标窗口同源
    // 如果不同源，则不能使用文件选择器API
    if (targetWindow !== window) {
      const targetOrigin = new URL((targetWindow as Window).location.href).origin;
      const currentOrigin = new URL(window.location.href).origin;

      if (targetOrigin !== currentOrigin) {
        return false;
      }
    }

    return true;
  } catch (e) {
    // 出现任何错误都返回false，表示不能使用API
    console.error('Error checking file picker API availability:', e);
    return false;
  }
};
