import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { normalizePlayerConfig, resolvePlayerMode } from '../playerPreference';

const device = (ua: string, width: number, height: number, platform = '', touches = 0) => {
  vi.stubGlobal('navigator', { userAgent: ua, platform, maxTouchPoints: touches });
  vi.stubGlobal('innerWidth', width);
  vi.stubGlobal('innerHeight', height);
};
describe('大屏移动设备播放策略', () => {
  beforeEach(() => { localStorage.clear(); });
  afterEach(() => vi.unstubAllGlobals());
  it('旧配置让平板保留桌面播放器', () => {
    device('Android Tablet', 800, 1280);
    expect(resolvePlayerMode({})).toBe('desktop');
  });
  it('只替换大屏移动设备，电脑触屏和手机不跟随大屏配置', () => {
    const config = { largeMobilePlayer: 'mobile' as const };
    device('Android Tablet', 800, 1280);
    expect(resolvePlayerMode(config)).toBe('mobile');
    device('Windows NT', 1920, 1080, 'Win32', 10);
    expect(resolvePlayerMode(config)).toBe('desktop');
    device('iPhone', 390, 844);
    expect(resolvePlayerMode(config)).toBe('desktop');
  });
  it('识别桌面 UA 的 iPad，横竖屏结果一致', () => {
    device('Macintosh', 820, 1180, 'MacIntel', 5);
    expect(resolvePlayerMode({ largeMobilePlayer: 'mobile' })).toBe('mobile');
    device('Macintosh', 1180, 820, 'MacIntel', 5);
    expect(resolvePlayerMode({ largeMobilePlayer: 'mobile' })).toBe('mobile');
  });
  it('两条边任意达到阈值即命中，允许后台调整', () => {
    const config = { largeMobilePlayer: 'mobile' as const };
    device('Android', 767, 1023);
    expect(resolvePlayerMode(config)).toBe('desktop');
    device('Android', 768, 900);
    expect(resolvePlayerMode(config)).toBe('mobile');
    device('Android', 700, 1024);
    expect(resolvePlayerMode(config)).toBe('mobile');
    expect(resolvePlayerMode({ ...config, largeMobileShortSide: 800, largeMobileLongSide: 1200 })).toBe('desktop');
  });
  it('只服从后台，忽略之前保存的本设备和页面偏好', () => {
    device('Android', 800, 1280);
    localStorage.setItem('mobile_show', '1');
    localStorage.setItem('video_player_preference', 'desktop');
    expect(resolvePlayerMode({ largeMobilePlayer: 'mobile' })).toBe('mobile');
    localStorage.setItem('video_player_preference', 'mobile');
    expect(resolvePlayerMode({})).toBe('desktop');
  });
  it('异常尺寸配置回退并规范短长边顺序', () => {
    expect(normalizePlayerConfig({ largeMobileShortSide: -1, largeMobileLongSide: NaN }))
      .toEqual({ largeMobilePlayer: 'desktop', largeMobileShortSide: 768, largeMobileLongSide: 1024 });
    expect(normalizePlayerConfig({ largeMobileShortSide: 1200, largeMobileLongSide: 800 }).largeMobileShortSide).toBe(800);
  });
});
