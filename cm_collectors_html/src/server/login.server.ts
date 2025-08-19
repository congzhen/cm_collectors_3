import request from "@/assets/request";
const routerGroupUri = '';
export const loginServer = {
  adminLogin: async (password: string) => {
    return await request<string>({
      url: `${routerGroupUri}/login/admin`,
      method: 'post',
      data: {
        password,
      }
    });
  }
}
