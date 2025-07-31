import type { I_performer } from "@/dataType/performer.dataType";
import type { I_resource, I_resource_base } from "@/dataType/resource.dataType";

export const getResourceCoverPoster = (resource: I_resource | I_resource_base | undefined) => {
  if (!resource || !resource.coverPoster || resource.coverPoster == '') return '';
  return `/api/resCoverPoster/${resource.filesBases_id}/${resource.coverPoster}`
}

export const getPerformerPhoto = (performer: I_performer | undefined) => {
  if (!performer || !performer.photo || performer.photo == '') return '';
  return `/api/performerFace/${performer.performerBases_id}/${performer.photo}`
}


export const getFileImageByDramaSeriesId = (dramaSeriesId: string, filesName: string, thumbWidth: number = 0, thumbLevel: number = 0) => {
  if (dramaSeriesId == '' || filesName == '') return '';
  const encodedFileName = btoa(encodeURIComponent(filesName));

  // 构建基础URL
  let url = `/api/files/image/${dramaSeriesId}/${encodedFileName}`;

  // 添加查询参数
  const params = new URLSearchParams();
  if (thumbWidth > 0) {
    params.append('thumbWidth', thumbWidth.toString());
  }
  if (thumbLevel > 0) {
    params.append('thumbLevel', thumbLevel.toString());
  }

  // 拼接查询参数
  const queryString = params.toString();
  return url + (queryString ? `?${queryString}` : '');
}
