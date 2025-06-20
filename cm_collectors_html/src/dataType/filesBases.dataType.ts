import type { I_config_app, I_config_nfo, I_config_simple } from "./config.dataType";

export interface I_filesBases {
  id: string;
  name: string;
  sort: number;
  addTime: string;
  status: boolean;
  filesRelatedPerformerBases: I_filesRelatedPerformerBases[];
  filesBasesSetting: I_filesBasesSetting;
}

export interface I_filesBasesSetting {
  filesBases_id: string;
  config_app: I_config_app;
  config_nfo: I_config_nfo;
  config_simple: I_config_simple;

  config_json_data: string;
  nfo_json_data: string;
  simple_json_data: string;
}

export interface I_filesRelatedPerformerBases {
  id: string;
  filesBases_id: string;
  performerBases_id: string;
  main: boolean;
}
