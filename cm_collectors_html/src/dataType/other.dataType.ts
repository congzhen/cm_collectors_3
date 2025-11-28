export interface I_scraperOneResource {
  title: string;
  issueNumber: string;
  dramaSeriesSrc: string;
}

export type I_databaseCleanupFormClearItem = 'resource' | 'performer' | 'tags' | 'tagClass' | 'fileDatabaseConfig' | 'importConfig' | 'resourceScraperConfig' | 'performerScraperConfig' | 'generalConfig' | 'cronJobs';
export interface I_databaseCleanupForm {
  filesBases_ids: string[];
  clear_items: I_databaseCleanupFormClearItem[];
}
