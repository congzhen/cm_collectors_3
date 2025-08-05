export interface I_filesBases_base {
  id: string;
  name: string;
  sort: number;
  addTime: string;
  status: boolean;
}

export interface I_filesBases extends I_filesBases_base {
  filesRelatedPerformerBases: I_filesRelatedPerformerBases[];
  filesBasesSetting: I_filesBasesSetting;
}


export interface I_filesBasesSetting {
  filesBases_id: string;
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


export interface I_filesBases_sort {
  id: string;
  sort: number;
}
