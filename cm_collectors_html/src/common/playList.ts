import { resourceServer } from '@/server/resource.server';
import { appStoreData } from '@/storeData/app.storeData';
import { ElMessage } from 'element-plus'

const playListKey = 'CM_PlayList'; // 播放列表存储的键
const playListMax = 100; // 最大播放列表数量

const getPlayListKey = () => {
  const store = {
    appStoreData: appStoreData(),
  }
  return playListKey + "-" + store.appStoreData.currentFilesBases.id;
}

export const getPlayListResource = async () => {
  const ids = playList()
  const result = await resourceServer.dataListByIds(ids);
  if (!result || !result.status) {
    ElMessage.error(result.msg);
    return [];
  }
  // 根据 ids 的顺序对结果进行排序
  const dataMap = new Map(result.data.map(item => [item.id, item]));
  return ids.map(id => dataMap.get(id)).filter(item => item !== undefined);
}

export const playList = (): string[] => {
  try {
    const stored = localStorage.getItem(getPlayListKey())
    if (stored) {
      return JSON.parse(stored)
    }
    return []
  } catch (error) {
    ElMessage.error('获取播放列表失败')
    console.error('获取播放列表失败:', error)
    return []
  }
}

export const playListAdd = (resourceId: string): boolean => {
  try {
    const list = playList()
    if (list.length >= playListMax) {
      ElMessage.warning('播放列表已满')
      return false
    }
    // 检查是否已存在
    if (!list.includes(resourceId)) {
      //添加到数组头部
      list.unshift(resourceId)
      localStorage.setItem(getPlayListKey(), JSON.stringify(list))
      ElMessage.success('已添加到播放列表')
      return true
    } else {
      ElMessage.warning('已在播放列表中')
      return false
    }
  } catch (error) {
    ElMessage.error('添加到播放列表失败')
    console.error('添加到播放列表失败:', error)
    return false
  }
}

export const playListRemove = (resourceId: string): boolean => {
  try {
    const list = playList()
    const index = list.indexOf(resourceId)
    if (index > -1) {
      list.splice(index, 1)
      localStorage.setItem(getPlayListKey(), JSON.stringify(list))
      ElMessage.success('已从播放列表删除')
      return true
    } else {
      ElMessage.warning('未在播放列表中找到该项')
      return false
    }
  } catch (error) {
    ElMessage.error('从播放列表删除失败')
    console.error('从播放列表删除失败:', error)
    return false
  }
}

export const playListClear = (): boolean => {
  try {
    localStorage.removeItem(getPlayListKey())
    ElMessage.success('播放列表已清空')
    return true
  } catch (error) {
    ElMessage.error('清空播放列表失败')
    console.error('清空播放列表失败:', error)
    return false
  }
}
