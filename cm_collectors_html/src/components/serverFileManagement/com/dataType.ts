export enum E_sfm_ToolBar {
  CreateFile = 'createFile',
  CreateFolder = 'createFolder',
  UploadFile = 'uploadFile',
  Download = 'download',
  Copy = 'copy',
  Move = 'move',
  Paste = 'paste',
  Compress = 'compress',
  Permissions = 'permissions',
  Delete = 'delete',
  Search = 'search',
}

export enum E_sfm_Column {
  Name = 'name',
  Permissions = 'permissions',
  Size = 'size',
  ModifiedAt = 'modifiedAt',
  Operate = 'operate',
}

export enum E_sfm_FileOperate {
  Open = 'open',
  Select = 'select',
  Download = 'download',
  Permissions = 'permissions',
  Extract = 'extract',
  Rename = 'rename',
  Delete = 'delete',
}

export enum E_sfm_FileType {
  File = 'file',
  Directory = 'directory',
}


export interface I_sfm_FileEntry {
  name: string;
  type: E_sfm_FileType;
  is_dir: boolean
  permissions: string
  size: number
  modified_at: string
  path: string
}

export interface I_sfm_FilesAction {
  type: E_sfm_ToolBar.Copy | E_sfm_ToolBar.Move;
  files: string[]
}


//文件扩展名
export const imageExtensions = ['jpg', 'jpeg', 'png', 'gif', 'webp'];
export const textExtensions = ['txt', 'log', 'md', 'csv'];
export const documentExtensions = ['pdf', 'doc', 'docx', 'xlsx'];
export const codeExtensions = ['js', 'ts', 'json', 'html', 'css', 'py', 'java', 'c', 'cpp', 'go', 'rb', 'php', 'sql', 'sh', 'bat', 'ps1', 'vbs', 'vb', 'pl', 'm', 'r', 'swift', 'kt', 'scala', 'groovy', 'lua', 'perl', 'ruby', 'rust', 'cmake', 'makefile', 'yaml', 'yml', 'ini'];
