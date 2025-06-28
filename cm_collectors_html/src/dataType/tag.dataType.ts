export interface I_tagData {
  tag: I_tag[];
  tagClass: I_tagClass[];
}

export interface I_tag {
  id: string;
  name: string;
  tagClass_id: string;
  hot: number;
  sort: number;
  status: boolean;
}

export interface I_tagClass {
  id: string;
  name: string;
  filesBases_id: string;
  leftShow: boolean;
  sort: number;
  status: boolean;
}

export interface I_tagSort {
  id: string;
  sort: number;
}
