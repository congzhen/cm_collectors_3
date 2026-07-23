import request, { requestBlob } from '@/assets/request';
import type {
  I_performerAvatarBatchPreview,
  I_performerAvatarBatchProgress,
  I_performerAvatarBatchActorPage,
  I_performerAvatarCandidate,
  I_performerAvatarLibraryStatus,
  PerformerAvatarStrategy,
} from '@/dataType/performerAvatarLibrary.dataType';

export const performerAvatarLibraryServer = {
  status: async () => request<I_performerAvatarLibraryStatus>({
    url: '/performerAvatarLibrary/status',
  }),
  updateDataFile: async () => request<I_performerAvatarLibraryStatus>({
    url: '/performerAvatarLibrary/updateDataFile',
    method: 'post',
  }),
  candidates: async (performerId: string) => request<I_performerAvatarCandidate[]>({
    url: `/performerAvatarLibrary/candidates/${performerId}`,
  }),
  batchActors: async (performerBasesId: string, page: number, limit: number, search: string, photoFilter: string) => request<I_performerAvatarBatchActorPage>({
    url: `/performerAvatarLibrary/batchActors/${performerBasesId}/${page}/${limit}`,
    params: { search, photoFilter },
  }),
  batchActorIds: async (performerBasesId: string, search: string, photoFilter: string) => request<string[]>({
    url: `/performerAvatarLibrary/batchActorIds/${performerBasesId}`,
    params: { search, photoFilter },
  }),
  apply: async (performerId: string, candidateId: string, overwrite: boolean) => request<boolean>({
    url: '/performerAvatarLibrary/apply',
    method: 'post',
    data: { performerId, candidateId, overwrite },
  }),
  batchPreview: async (performerBasesId: string, allPerformers: boolean, performerIds: string[], strategy: PerformerAvatarStrategy, overwrite: boolean) => request<I_performerAvatarBatchPreview>({
    url: '/performerAvatarLibrary/batchPreview',
    method: 'post',
    data: { performerBasesId, allPerformers, performerIds, strategy, overwrite },
  }),
  batchApply: async (performerBasesId: string, allPerformers: boolean, performerIds: string[], strategy: PerformerAvatarStrategy, overwrite: boolean) => request<I_performerAvatarBatchProgress>({
    url: '/performerAvatarLibrary/batchApply',
    method: 'post',
    data: { performerBasesId, allPerformers, performerIds, strategy, overwrite },
  }),
  batchProgress: async (batchId: string) => request<I_performerAvatarBatchProgress>({
    url: `/performerAvatarLibrary/batchProgress/${batchId}`,
  }),
  previewImage: async (performerId: string, candidateId: string) => requestBlob({
    url: `/performerAvatarLibrary/preview/${performerId}/${candidateId}`,
  }),
};
