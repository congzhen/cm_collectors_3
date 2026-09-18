export type PlayerMode = 'desktop' | 'mobile';
export interface LargeMobilePlayerConfig {
  largeMobilePlayer?: PlayerMode;
  largeMobileShortSide?: number;
  largeMobileLongSide?: number;
}

export function normalizePlayerConfig(config: LargeMobilePlayerConfig) {
  const size = (value: number | undefined, fallback: number) =>
    typeof value === 'number' && Number.isInteger(value) && value >= 320 && value <= 4096 ? value : fallback;
  const short = size(config.largeMobileShortSide, 768);
  const long = size(config.largeMobileLongSide, 1024);
  return {
    largeMobilePlayer: config.largeMobilePlayer === 'mobile' ? 'mobile' as const : 'desktop' as const,
    largeMobileShortSide: Math.min(short, long),
    largeMobileLongSide: Math.max(short, long),
  };
}

// 页面布局继续使用原 mobile.ts；这里独立判断播放策略，不受页面开关影响。
export function resolvePlayerMode(config: LargeMobilePlayerConfig): PlayerMode {
  const settings = normalizePlayerConfig(config);
  const ua = navigator.userAgent || navigator.vendor || '';
  const mobileDevice = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini|Mobile|CriOS|FxiOS|EdgiOS/i.test(ua)
    || (navigator.platform === 'MacIntel' && navigator.maxTouchPoints > 1);
  const short = Math.min(window.innerWidth, window.innerHeight);
  const long = Math.max(window.innerWidth, window.innerHeight);
  return mobileDevice && (short >= settings.largeMobileShortSide || long >= settings.largeMobileLongSide)
    ? settings.largeMobilePlayer : 'desktop';
}
