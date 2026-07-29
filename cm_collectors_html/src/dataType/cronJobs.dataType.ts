
import type { I_filesBases } from './filesBases.dataType';

export type jobsType = 'import' | 'scraperResource' | 'scraperPerformer' | 'clear' | 'videoFingerprint' | 'aiTag' | 'videoMetadata' | 'clearPerformerAvatarCache';

export interface I_cronJobs {
  id: string;
  filesBases_id: string;
  filesBasesIds: string[];
  filesBasesList?: I_filesBases[];
  scopeMode: 'selected' | 'all';
  configJsonData: string;
  jobs_type: jobsType;
  cron_expression: string;
}

export interface I_cronJobs_info extends I_cronJobs {
  created_at: string;
  last_exec_at: string;
  last_exec_error: string;
  last_exec_status: boolean;
  status: boolean;
  running: boolean;
}
