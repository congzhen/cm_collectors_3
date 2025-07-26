import type { I_resource } from "@/dataType/resource.dataType";
import { messageBoxConfirm } from "./messageBox";
import { resourceServer } from "@/server/resource.server";
import { ElMessage } from "element-plus";

export const resourceDelete = (resource: I_resource, callBackFn: () => void) => {
  messageBoxConfirm({
    text: `确定要删除【${resource.title}】吗？`,
    type: 'warning',
    successCallBack: () => {
      // 删除
      resourceDeleteExec(resource, callBackFn);
    },
    failCallBack: () => {
      // 取消
    }
  })
}

const resourceDeleteExec = async (resource: I_resource, callBackFn: () => void) => {
  const result = await resourceServer.delete(resource.id);
  if (!result || !result.status) {
    ElMessage.error(result.msg);
  } else {
    ElMessage.success('删除成功');
    callBackFn();
  }
}
