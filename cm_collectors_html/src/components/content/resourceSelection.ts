import type { ComputedRef, InjectionKey } from 'vue';
import type { I_resource } from '@/dataType/resource.dataType';

export interface ResourceSelection {
  active: ComputedRef<boolean>;
  selectedIds: ComputedRef<Set<string>>;
  groupLabels?: ComputedRef<Map<string, string>>;
  toggle: (resource: I_resource) => void;
}

export const resourceSelectionKey: InjectionKey<ResourceSelection> = Symbol('resource-selection');

// 本页组号按此列表循环配色，保持卡片和两种表格中的标记一致。
export const RESOURCE_GROUP_MARK_COLORS = [
  { background: '#dbeafe', text: '#1e40af', border: '#93c5fd' },
  { background: '#d1fae5', text: '#065f46', border: '#6ee7b7' },
  { background: '#fef3c7', text: '#92400e', border: '#fcd34d' },
  { background: '#ede9fe', text: '#5b21b6', border: '#c4b5fd' },
  { background: '#ffe4e6', text: '#9f1239', border: '#fda4af' },
  { background: '#cffafe', text: '#155e75', border: '#67e8f9' },
  { background: '#ffedd5', text: '#9a3412', border: '#fdba74' },
  { background: '#ecfccb', text: '#3f6212', border: '#bef264' },
] as const;
