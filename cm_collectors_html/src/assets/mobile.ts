const Mobile_Show_Key = 'mobile_show';

// 是否关闭移动端显示，默认为 false，表示不关闭移动端显示
let closeMobileDisplay = false;

export const setCloseMobileDisplay = (close: boolean) => {
  closeMobileDisplay = close === true;
};

export const setMobileShow = (show: boolean) => {
  localStorage.setItem(Mobile_Show_Key, show ? '1' : '0');
};

export const getMobileShow = () => {
  return localStorage.getItem(Mobile_Show_Key) === '1';
};

export const goToMobileOrPC = () => {
  if (isMobile()) {
    location.href = '/mobile';
  } else {
    location.href = '/';
  }
};

export const isMobile = () => {
  return isMobileDevice() && !getMobileShow() && !closeMobileDisplay;
};

export const isMobileDevice = () => {
  try {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const userAgent = navigator.userAgent || navigator.vendor || (window as any).opera || '';
    const viewportShortSide = Math.min(window.innerWidth, window.innerHeight);
    const viewportLongSide = Math.max(window.innerWidth, window.innerHeight);
    const isCompactViewport = viewportShortSide < 768 && viewportLongSide < 1024;
    const isIPadOS = navigator.platform === 'MacIntel' && navigator.maxTouchPoints > 1;
    const isMobileUA = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini|Mobile|CriOS|FxiOS|EdgiOS/i.test(userAgent);
    const isTabletUA = /Tablet|iPad/i.test(userAgent) || (/Android/i.test(userAgent) && !/Mobile/i.test(userAgent));

    return isCompactViewport && (isMobileUA || isTabletUA || isIPadOS);
  } catch (e) {
    console.error('Error in mobile detection:', e);
    return false;
  }
};
