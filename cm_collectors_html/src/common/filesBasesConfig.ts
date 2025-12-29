import { loadJsonFile, saveJsonFile } from "@/assets/file";
import { currentFormatDate } from "@/assets/timer";
import type { I_config_app } from "@/dataType/config.dataType";
import { filesBasesStoreData } from "@/storeData/filesBases.storeData";
import { ElMessage } from "element-plus";
import { messageBoxConfirm } from "./messageBox";

export const filesBasesConfigExport = async (filesBasesId: string, configObj: I_config_app) => {
  const store = filesBasesStoreData()
  const filesBasesName = store.getFilesBasesNameById(filesBasesId)
  const timestamp = currentFormatDate('Ymd_His')
  const exportStr = JSON.stringify(configObj, null, 2)
  await saveJsonFile(exportStr, `${filesBasesName}-配置参数-${timestamp}.json`)
}
export const filesBasesConfigImport = (): Promise<I_config_app | null> => {
  return new Promise((resolve) => {
    messageBoxConfirm({
      text: '导入配置时，如果某些项目（如演员、标签等）在当前文件库中不存在，则会自动移除这些项目。为确保配置导入完整，建议先关联好演员集和标签集。',
      title: '导入配置',
      type: 'warning',
      successCallBack: async () => {
        try {
          const fileData = await loadJsonFile()
          if (!fileData) {
            ElMessage.error('未选择文件');
            resolve(null)
            return
          }
          const configObj = JSON.parse(fileData.content) as I_config_app
          resolve(configObj)
        } catch (error) {
          ElMessage.error((error as Error).message || '导入配置失败');
          console.log(error)
          resolve(null)
        }
      },
      failCallBack: () => {
        resolve(null)
      }
    })
  })
}
