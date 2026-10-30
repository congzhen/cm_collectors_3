import type { ComputedRef, InjectionKey } from 'vue';
import type { I_resource } from '@/dataType/resource.dataType';

export interface ResourceSelection {
  active: ComputedRef<boolean>;
  selectedIds: ComputedRef<Set<string>>;
  toggle: (resource: I_resource) => void;
}

export const resourceSelectionKey: InjectionKey<ResourceSelection> = Symbol('resource-selection');
