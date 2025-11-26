
export type jobsType = 'import' | 'scraperResource' | 'scraperPerformer' | 'clear';

export interface I_cronJobs {
  id: string;
  filesBases_id: string;
  jobs_type: jobsType;
  cron_expression: string;
}

export interface I_cronJobs_info extends I_cronJobs {
  created_at: string;
  last_exec_at: string;
  last_exec_error: string;
  last_exec_status: boolean;
  status: boolean;
}
