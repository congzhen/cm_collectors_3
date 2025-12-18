import request from "@/assets/request";
const routerGroupUri = '';
export const loginServer = {
  adminLogin: async (password: string) => {
    const result = await request<string>({
      url: `${routerGroupUri}/login/admin`,
      method: 'post',
      data: {
        password,
      }
    });
    if (result && result.status) {
      //获取返回头中的token，并设置存储
      sessionStorage.setItem('adminToken', result.data);
    }
    return result;
  }
}
