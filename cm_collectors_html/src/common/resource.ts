import type { I_resource } from "@/dataType/resource.dataType";
import { messageBoxConfirm } from "./messageBox";
import { resourceServer } from "@/server/resource.server";
import { ElMessage } from "element-plus";
import { LoadingService } from "@/assets/loading";
import { appStoreData } from "@/storeData/app.storeData";

export const resourceDelete = (resource: I_resource, callBackFn: () => void = () => { }) => {
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

export const resourceBatchDelete = (resources: I_resource[], callBackFn: () => void = () => { }) => {
  if (resources.length == 0) {
    ElMessage.error('请选择要删除的资源');
  } else {
    messageBoxConfirm({
      text: `确定要删除当前选中的 ${resources.length} 个资源吗？`,
      type: 'warning',
      successCallBack: async () => {
        // 删除
        try {
          LoadingService.show();
          let refreshTagData = false;
          let successCount = 0;
          for (let i = 0; i < resources.length; i++) {
            let deleted = false;
            try {
              deleted = await resourceDeleteExec(resources[i], () => { }, false, false);
            } catch (error) {
              console.error(error);
            }
            if (deleted) successCount++;
            if (deleted && resources[i].filesBases_id === appStoreData().currentFilesBases.id) {
              refreshTagData = true;
            }
          }
          const failedCount = resources.length - successCount;
          if (failedCount === 0) {
            ElMessage.success(`成功删除 ${successCount} 项资源`);
          } else if (successCount === 0) {
            ElMessage.error(`删除失败，共 ${failedCount} 项资源`);
          } else {
            ElMessage.warning(`成功删除 ${successCount} 项资源，失败 ${failedCount} 项`);
          }
          callBackFn();
          if (refreshTagData) {
            void appStoreData().refreshCurrentTagData();
          }
        } catch (error) {
          console.log(error)
        } finally {
          LoadingService.hide();
        }

      },
    })
  }
}

const resourceDeleteExec = async (
  resource: I_resource,
  callBackFn: () => void = () => { },
  refreshTagData = true,
  showMessage = true,
): Promise<boolean> => {
  const result = await resourceServer.delete(resource.id);
  if (!result || !result.status) {
    if (showMessage) ElMessage.error(result?.msg || '删除失败');
    return false;
  } else {
    if (showMessage) ElMessage.success('删除成功');
    callBackFn();
    if (refreshTagData && resource.filesBases_id === appStoreData().currentFilesBases.id) {
      void appStoreData().refreshCurrentTagData();
    }
    return true;
  }
}


export const resourcePinToTop = async (resource: I_resource, callBackFn: () => void = () => { }, pinToTopStatus?: boolean) => {
  const status = pinToTopStatus !== undefined ? pinToTopStatus : !resource.pinToTop;
  const result = await resourceServer.pinToTop(resource.id, status);
  if (!result || !result.status) {
    ElMessage.error(result.msg);
  } else {
    ElMessage.success(status ? '置顶成功' : '取消置顶成功');
    callBackFn();
  }
}
