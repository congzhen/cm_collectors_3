import type { ComputedRef, InjectionKey, Ref } from 'vue';
import type { I_resource, I_resourceGroupSummary } from '@/dataType/resource.dataType';

export interface ResourceGroupPage {
  key: string;
  summaries: I_resourceGroupSummary[];
}
export function resourceGroupPageKey(library: string, resources: I_resource[]) {
  return JSON.stringify([library, resources.map(resource => [resource.id, resource.associationGroupId || ''])]);
}

// 布局提供组内资源的播放动作；未提供时由卡片统一使用 playResource。
export const resourceGroupSelectKey: InjectionKey<(resource: I_resource) => void> = Symbol('resource-group-select');
export const resourceGroupKeepOpenKey: InjectionKey<boolean> = Symbol('resource-group-keep-open');
let activeFanCloser: (() => void) | undefined;
export function claimResourceGroupFan(close: () => void) {
  if (activeFanCloser !== close) activeFanCloser?.();
  activeFanCloser = close;
}
export function releaseResourceGroupFan(close: () => void) {
  if (activeFanCloser === close) activeFanCloser = undefined;
}

// 切库时父列表可能尚未更新，旧页和页外候选均不能重新进入播放器。
export function resourceGroupPlaybackList(resources: I_resource[], selected: I_resource | null, library: string): I_resource[] {
  const current = resources.every(resource => resource.filesBases_id === library)
    ? resources : resources.filter(resource => resource.filesBases_id === library);
  return selected?.filesBases_id === library && !current.some(resource => resource.id === selected.id)
    ? [...current, selected] : current;
}
export const resourceGroupContextKey: InjectionKey<{
  enabled: ComputedRef<boolean>;
  pageIds: ComputedRef<Set<string>>;
  summaries: Ref<Map<string, I_resourceGroupSummary>>;
}> = Symbol('resource-group-context');

// 只合并本次返回的数据，不改变排序、资源计数或服务端分页。
export function mergeResourceGroupPage(resources: I_resource[], enabled: boolean): I_resource[] {
  if (!enabled) return resources;
  const groups = new Set<string>();
  return resources.filter(resource => {
    const group = resource.associationGroupId;
    if (!group) return true;
    if (groups.has(group)) return false;
    groups.add(group);
    return true;
  });
}
