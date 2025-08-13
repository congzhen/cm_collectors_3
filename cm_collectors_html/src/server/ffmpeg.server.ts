import request from "@/assets/request";
const routerGroupUri = '';
export const ffmpegServer = {
  getVideoKeyFramePosters: async (videoPath: string, frameCount: number) => {
    return await request<string[]>({
      url: `${routerGroupUri}/FFmpeg/getVideoKeyFramePosters`,
      method: 'post',
      data: {
        videoPath,
        frameCount
      }
    });
  }
}
