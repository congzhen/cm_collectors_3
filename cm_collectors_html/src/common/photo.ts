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


export const getFileImageByDramaSeriesId = (dramaSeriesId: string, filesName: string, thumbWidth: number = 0) => {
  if (dramaSeriesId == '' || filesName == '') return '';
  const encodedFileName = btoa(encodeURIComponent(filesName));
  return `/api/files/image/${dramaSeriesId}/${encodedFileName}` + (thumbWidth > 0 ? `?thumbWidth=${thumbWidth}` : '')
}
