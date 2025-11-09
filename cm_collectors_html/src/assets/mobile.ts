// 检测是否为移动设备
export const isMobile = () => {
  try {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const userAgent = navigator.userAgent || navigator.vendor || (window as any).opera;

    // 更全面的移动设备检测正则表达式
    const mobileRegex = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini|Mobile|mobile|CriOS|FxiOS|EdgiOS/i;

    // 检查用户代理
    const isMobileUA = mobileRegex.test(userAgent);

    // 检查触摸支持和屏幕尺寸
    const isTouchDevice = 'ontouchstart' in window || navigator.maxTouchPoints > 0;
    const isSmallScreen = window.innerWidth <= 1024 || window.innerHeight <= 1024;

    // 检查特定的移动设备特征
    const isMobileDevice = /Mobi|Tablet|Touch/i.test(userAgent) ||
      (isTouchDevice && isSmallScreen) ||
      Math.min(screen.width, screen.height) < 768;

    // 返回综合判断结果
    return isMobileUA || isMobileDevice;
  } catch (e) {
    // 出现异常时，默认返回false（桌面端）
    console.error('Error in mobile detection:', e);
    return false;
  }
};
