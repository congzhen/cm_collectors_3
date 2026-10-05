import { nextTick, onActivated, onBeforeUnmount, onDeactivated, onMounted, ref } from 'vue';

type ShellWindow = Window & { __cmPhoneShell?: boolean; __cmPhoneShellSettings?: boolean; __cmPhoneShellDownload?: boolean };

export function downloadMobileShell(url: string, filename: string): boolean {
  if (!(window as ShellWindow).__cmPhoneShellDownload) return false;
  const headers: Record<string, string> = {};
  for (const name of ['token', 'adminToken']) {
    const value = sessionStorage.getItem(name);
    if (value) headers[name] = value;
  }
  window.dispatchEvent(new CustomEvent('cm-phone-download', {
    detail: { url: new URL(url, window.location.href).href, filename, headers },
  }));
  return true;
}
const activeMenus = new Set<symbol>();
function announceMenus() {
  void nextTick(() => {
    if ((window as ShellWindow).__cmPhoneShellSettings) {
      window.dispatchEvent(new CustomEvent('cm-phone-settings-ready', { detail: { available: activeMenus.size > 0 } }));
    }
  });
}
export function openMobileShellSettings() {
  if (!(window as ShellWindow).__cmPhoneShellSettings) return;
  window.dispatchEvent(new Event('cm-phone-settings'));
}
export function useMobileShellSettings() {
  const available = ref(false);
  const id = Symbol('shell-menu');
  let active = false;
  const refresh = () => {
    available.value = Boolean((window as ShellWindow).__cmPhoneShellSettings);
    if (active && available.value) activeMenus.add(id);
    else activeMenus.delete(id);
    announceMenus();
  };
  const activate = () => { active = true; refresh(); };
  const deactivate = () => { active = false; activeMenus.delete(id); announceMenus(); };
  onMounted(() => { window.addEventListener('cm-phone-shell-ready', refresh); activate(); });
  onActivated(activate);
  onDeactivated(deactivate);
  onBeforeUnmount(() => { window.removeEventListener('cm-phone-shell-ready', refresh); deactivate(); });
  return { available, openSettings: openMobileShellSettings };
}

// 此标记只由新版 APK 向其 WebView 注入；普通浏览器不会接管原生窗口。
export function notifyMobileShell(full: boolean, landscape = false): boolean {
  if (!(window as Window & { __cmPhoneShell?: boolean }).__cmPhoneShell) return false;
  window.dispatchEvent(new CustomEvent('cm-phone-fullscreen', { detail: { full, landscape } }));
  return true;
}
