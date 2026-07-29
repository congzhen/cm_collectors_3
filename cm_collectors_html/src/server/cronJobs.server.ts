import request from "@/assets/request";
import type { I_cronJobs, I_cronJobs_info } from "@/dataType/cronJobs.dataType";
const routerGroupUri = '/cronJobs';
export const cronJobsServer = {
  list: async () => {
    return await request<I_cronJobs_info[]>({
      url: `${routerGroupUri}/list`,
      method: 'get',
    })
  },
  exec: async (id: string) => {
    return await request<boolean>({
      url: `${routerGroupUri}/exec/${id}`,
      method: 'get',
    })
  },
  create: async (data: I_cronJobs) => {
    return await request<I_cronJobs_info>({
      url: `${routerGroupUri}/create`,
      method: 'post',
      data,
    });
  },
  update: async (data: I_cronJobs) => {
    return await request<I_cronJobs_info>({
      url: `${routerGroupUri}/update`,
      method: 'put',
      data,
    });
  },
  delete: async (id: string) => {
    return await request<boolean>({
      url: `${routerGroupUri}/delete/${id}`,
      method: 'delete',
    });
  },
}
