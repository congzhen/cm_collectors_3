import request from "@/assets/request";
import type { I_cronJobs_info } from "@/dataType/cronJobs.dataType";
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
  create: async (filesBases_id: string, jobs_type: string, cron_expression: string) => {
    return await request<I_cronJobs_info>({
      url: `${routerGroupUri}/create`,
      method: 'post',
      data: {
        filesBases_id,
        jobs_type,
        cron_expression,
      },
    });
  },
  update: async (id: string, filesBases_id: string, jobs_type: string, cron_expression: string) => {
    return await request<I_cronJobs_info>({
      url: `${routerGroupUri}/update`,
      method: 'put',
      data: {
        id,
        filesBases_id,
        jobs_type,
        cron_expression,
      },
    });
  },
  delete: async (id: string) => {
    return await request<boolean>({
      url: `${routerGroupUri}/delete/${id}`,
      method: 'delete',
    });
  },
}
