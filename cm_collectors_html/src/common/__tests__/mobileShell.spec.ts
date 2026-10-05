import { afterEach, describe, expect, it, vi } from 'vitest';
import { downloadMobileShell, notifyMobileShell } from '../mobileShell';
import { useMobileShellSettings } from '../mobileShell';
import { mount, flushPromises } from '@vue/test-utils';
import { defineComponent, h } from 'vue';
const shellWindow = window as Window & { __cmPhoneShell?: boolean };
afterEach(() => {
  delete shellWindow.__cmPhoneShell;
  delete (window as Window & { __cmPhoneShellDownload?: boolean }).__cmPhoneShellDownload;
  sessionStorage.clear();
  delete (window as Window & { __cmPhoneShellSettings?: boolean }).__cmPhoneShellSettings;
});
describe('手机套壳全屏通知', () => {
  it('普通浏览器和旧 APK 下载仍使用播放器原来的方式', () => {
    expect(downloadMobileShell('/api/video/mp4/e/v.mp4', '视频.mp4')).toBe(false);
  });
  it('新版 APK 下载传递绝对地址和登录请求头', () => {
    (window as Window & { __cmPhoneShellDownload?: boolean }).__cmPhoneShellDownload = true;
    sessionStorage.setItem('token', 'user-token');
    sessionStorage.setItem('adminToken', 'admin-token');
    const listener = vi.fn();
    window.addEventListener('cm-phone-download', listener);
    try {
      expect(downloadMobileShell('/api/video/mp4/e/v.mp4', '视频.mp4')).toBe(true);
      expect((listener.mock.calls[0][0] as CustomEvent).detail).toEqual({
        url: new URL('/api/video/mp4/e/v.mp4', window.location.href).href,
        filename: '视频.mp4', headers: { token: 'user-token', adminToken: 'admin-token' },
      });
    } finally { window.removeEventListener('cm-phone-download', listener); }
  });
  it('服务器入口仅在 APK 就绪后出现，卸载时恢复原生入口', async () => {
    const received: boolean[] = [];
    const listener = (event: Event) => received.push((event as CustomEvent).detail.available);
    const open = vi.fn();
    window.addEventListener('cm-phone-settings-ready', listener);
    window.addEventListener('cm-phone-settings', open);
    const wrapper = mount(defineComponent({ setup() {
      const { available, openSettings } = useMobileShellSettings();
      return () => available.value ? h('button', { onClick: openSettings }, '服务器') : null;
    } }));
    expect(wrapper.find('button').exists()).toBe(false);
    (window as Window & { __cmPhoneShellSettings?: boolean }).__cmPhoneShellSettings = true;
    window.dispatchEvent(new Event('cm-phone-shell-ready')); await flushPromises();
    expect(wrapper.find('button').exists()).toBe(true); expect(received.at(-1)).toBe(true);
    await wrapper.find('button').trigger('click'); expect(open).toHaveBeenCalledOnce();
    wrapper.unmount(); await flushPromises(); expect(received.at(-1)).toBe(false);
    window.removeEventListener('cm-phone-settings-ready', listener);
    window.removeEventListener('cm-phone-settings', open);
  });
  it('普通浏览器不启用套壳全屏', () => {
    expect(notifyMobileShell(true, true)).toBe(false);
  });
  it('新版套壳接收进入和退出通知', () => {
    shellWindow.__cmPhoneShell = true;
    const listener = vi.fn();
    window.addEventListener('cm-phone-fullscreen', listener);
    try {
      expect(notifyMobileShell(true, true)).toBe(true);
      expect((listener.mock.calls[0][0] as CustomEvent).detail).toEqual({ full: true, landscape: true });
      notifyMobileShell(false);
      expect((listener.mock.calls[1][0] as CustomEvent).detail.full).toBe(false);
    } finally { window.removeEventListener('cm-phone-fullscreen', listener); }
  });
});
