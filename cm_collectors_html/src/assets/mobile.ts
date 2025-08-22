// 检测是否为移动设备
export const isMobile = () => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const userAgent = navigator.userAgent || navigator.vendor || (window as any).opera;
  return /android|iPad|iPhone|iPod|windows phone/i.test(userAgent);
};
