import { tagServer } from "@/server/tag.server";
import { messageBoxAlert } from "./messageBox";
import { filesBasesStoreData } from "@/storeData/filesBases.storeData";
import { currentFormatDate } from "@/assets/timer";
import { loadTextFile, saveTextFile } from "@/assets/file";

export const tagExport = async (filesBasesId: string) => {
  const result = await tagServer.tagDataByFilesBasesId(filesBasesId);
  if (!result || !result.status) {
    messageBoxAlert({
      text: '获取标签数据失败',
      type: 'error'
    })
    return
  }
  let exportStr = '// 如需修改标签名称，可在原标签后添加"=>新标签名"，例如: "-旧标签名=>新标签名"\n'
  result.data.tagClass.forEach(tagClass => {
    exportStr += `${tagClass.name}\n`
    result.data.tag.forEach(tag => {
      if (tag.tagClass_id === tagClass.id) {
        exportStr += `-${tag.name}\n`
      }
    })
  })
  const store = filesBasesStoreData()
  const filesBasesName = store.getFilesBasesNameById(filesBasesId)
  const timestamp = currentFormatDate('Ymd_His')
  await saveTextFile(exportStr, `${filesBasesName}-标签数据-${timestamp}.txt`)
}

export const tagImport = async (filesBasesId: string) => {
  const content = await loadTextFile()
  const slc = content.split('\n')

  // 创建一个对象来存储结果
  const import_data: Record<string, string[]> = {}
  let currentKey = ''

  slc.forEach((item: string) => {
    if (item === '' || item.startsWith('//') || item.startsWith('\n') || item.startsWith('\r')) {
      return
    }
    if (item.startsWith('-')) {
      // 如果行以 - 开头，则将其作为当前键的值添加到数组中
      const value = item.substring(1) // 移除开头的 -
      if (currentKey && import_data[currentKey]) {
        import_data[currentKey].push(value)
      }
    } else {
      // 如果行不以 - 开头，则将其作为新的键
      currentKey = item
      import_data[currentKey] = []
    }
  })
  const result = await tagServer.import(filesBasesId, import_data)
  if (!result || !result.status) {
    messageBoxAlert({
      text: result.msg,
      type: 'error'
    })
    return
  } else {
    messageBoxAlert({
      text: '标签数据导入成功，请刷新页面查看最新数据',
      type: 'success'
    })
  }
  return
}
