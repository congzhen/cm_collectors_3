export interface I_performerTagClass {
  id: string;
  performerBases_id: string;
  name: string;
  sort: number;
  status: boolean;
  createdAt?: string;
}

export interface I_performerTag {
  id: string;
  performerTagClass_id: string;
  name: string;
  sort: number;
  status: boolean;
  performerCount: number;
  createdAt?: string;
}

export interface I_performerTagData {
  tagClasses: I_performerTagClass[];
  tags: I_performerTag[];
}

export type PerformerTagMatchMode = 'any' | 'all';

export interface I_performerTagFilter {
  tagIds: string[];
  tagMatchMode: PerformerTagMatchMode;
}
