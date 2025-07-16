import type { I_sfm_FileEntry } from "./dataType"

const api = '/api/sfm/'

export const apiList = {
  pathDir: api + 'pathDir',
  searchFiles: api + 'searchFiles',
  createFile: api + 'createFile',
  createFolder: api + 'createFolder',
  openFile: api + 'openFile',
  saveFile: api + 'saveFile',
  uploadFile: api + 'uploadFile',
  downloadFile: api + 'downloadFile',
  copyFile: api + 'copyFile',
  moveFile: api + 'moveFile',
  compressFile: api + 'compressFile',
  permissionsFile: api + 'permissionsFile',
  deleteFile: api + 'deleteFile',
  renameFile: api + 'renameFile',
  unCompressFile: api + 'unCompressFile',
}


export const sfm_GetPathDir = async (path: string) => {
  const res = await request<IResponse<I_sfm_FileEntry[]>>(apiList.pathDir, 'POST', { path })
  return res
}
export const sfm_SearchFiles = async (path: string, search: string) => {
  const res = await request<IResponse<I_sfm_FileEntry[]>>(apiList.searchFiles, 'POST', { path, search })
  return res
}
export const sfm_CreateFile = async (name: string, path: string) => {
  const res = await request<IResponse<boolean>>(apiList.createFile, 'POST', { name, path })
  return res
}
export const sfm_CreateFolder = async (name: string, path: string, permissions: string) => {
  const res = await request<IResponse<boolean>>(apiList.createFolder, 'POST', { name, path, permissions })
  return res
}
export const sfm_OpenFile = async <T>(filePath: string, returnType: 'text' | 'base64' | 'blob' = 'text') => {
  const res = await request<IResponse<T>>(apiList.openFile, 'POST', { filePath, returnType })
  return res
}
export const sfm_SaveFile = async (filePath: string, content: string) => {
  const res = await request<IResponse<boolean>>(apiList.saveFile, 'POST', { filePath, content })
  return res
}

export const sfm_DownloadFile = (filePath: string) => {
  const encodedPath = encodeURIComponent(filePath)
  const url = `${apiList.downloadFile}?filePath=${encodedPath}`
  window.location.href = url // 直接触发浏览器下载
}

export const sfm_PasteCopy = async (path: string, files: string[]) => {
  const res = await request<IResponse<boolean>>(apiList.copyFile, 'POST', { path, files })
  return res
}
export const sfm_PasteMove = async (path: string, files: string[]) => {
  const res = await request<IResponse<boolean>>(apiList.moveFile, 'POST', { path, files })
  return res
}

export const sfm_CompressFile = async (path: string, files: string[]) => {
  const res = await request<IResponse<boolean>>(apiList.compressFile, 'POST', { path, files })
  return res
}

export const sfm_Permissions = async (files: string[], permissions: string, sub_files: boolean) => {
  const res = await request<IResponse<boolean>>(apiList.permissionsFile, 'POST', { files, permissions, sub_files })
  return res
}

export const sfm_DeleteFile = async (path: string, files: string[]) => {
  const res = await request<IResponse<boolean>>(apiList.deleteFile, 'POST', { path, files })
  return res
}

export const sfm_RenameFile = async (name: string, path: string) => {
  const res = await request<IResponse<boolean>>(apiList.renameFile, 'POST', { name, path })
  return res
}
export const sfm_UnCompressFile = async (file: string) => {
  const res = await request<IResponse<boolean>>(apiList.unCompressFile, 'POST', { file })
  return res
}

export interface IResponse<T> {
  status: boolean;
  statusCode: number;
  msg: string;
  data: T;
}

type RequestOptions = {
  headers?: Record<string, string>
  timeout?: number
}

interface RequestError extends Error {
  status?: number
  response?: any
}

const request = async <T>(
  url: string,
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' = 'GET',
  data?: any,
  { headers = {}, timeout = 10000 }: RequestOptions = {},
): Promise<T> => {
  const controller = new AbortController()
  const id = setTimeout(() => controller.abort(), timeout)

  try {
    // 处理不同方法的参数
    const options: RequestInit = {
      method,
      headers: {
        'Content-Type': 'application/json',
        ...headers,
      },
      signal: controller.signal,
    }

    if (['POST', 'PUT'].includes(method)) {
      options.body = JSON.stringify(data)
    } else if (method === 'GET' && data) {
      const searchParams = new URLSearchParams()
      for (const [key, value] of Object.entries(data)) {
        searchParams.append(key, String(value))
      }
      url += `?${searchParams}`
    }

    const response = await fetch(url, options)
    clearTimeout(id)

    if (!response.ok) {
      const error: RequestError = new Error(`HTTP Error ${response.status}: ${response.statusText}`)
      error.status = response.status
      try {
        error.response = await response.json()
      } catch {
        error.response = await response.text()
      }
      throw error
    }

    return (await response.json()) as T
  } catch (error) {
    if (error instanceof Error && error.name === 'AbortError') {
      throw new Error('Request timeout')
    } else {
      const unknownError = error as Error;
      console.error('请求失败:', unknownError)
      throw unknownError
    }
  }
}
