import request from "@/assets/request";
const routerGroupUri = '';
export const ffmpegServer = {
  getVideoThumbnails: async (videoPath: string, frameCount: number) => {
    return await request<string[]>({
      url: `${routerGroupUri}/FFmpeg/getVideoThumbnails`,
      method: 'post',
      data: {
        videoPath,
        frameCount
      }
    });
  }
}
