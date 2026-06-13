import { tagServer } from "@/server/tag.server";
import { messageBoxAlert } from "./messageBox";
import { filesBasesStoreData } from "@/storeData/filesBases.storeData";
import { currentFormatDate } from "@/assets/timer";
import { loadTextFile, saveTextFile } from "@/assets/file";
import type { I_tagImportItem } from "@/dataType/tag.dataType";

const escapeTagText = (value: string) => {
  return (value || '')
    .replace(/\\/g, '\\\\')
    .replace(/\r/g, '\\r')
    .replace(/\n/g, '\\n');
}

const unescapeTagText = (value: string) => {
  let result = '';
  for (let i = 0; i < value.length; i++) {
    const ch = value[i];
    if (ch !== '\\' || i === value.length - 1) {
      result += ch;
      continue;
    }
    const next = value[++i];
    if (next === 'n') result += '\n';
    else if (next === 'r') result += '\r';
    else if (next === '\\') result += '\\';
    else result += next;
  }
  return result;
}

const parseAiEnabled = (value: string): boolean | undefined => {
  const normalized = value.trim().toLowerCase();
  if (['true', '1', 'yes', 'y', 'on', '是', '启用'].includes(normalized)) return true;
  if (['false', '0', 'no', 'n', 'off', '否', '禁用'].includes(normalized)) return false;
  return undefined;
}

const parseTagImportText = (content: string) => {
  const importData: Record<string, I_tagImportItem[]> = {};
  let currentClass = '';
  let currentTag: I_tagImportItem | null = null;

  content.split(/\r?\n/).forEach((rawLine) => {
    const line = rawLine.trimEnd();
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith('//')) {
      return;
    }

    if (/^\s+/.test(line) && currentTag) {
      const splitIndex = trimmed.indexOf(':');
      if (splitIndex === -1) return;
      const key = trimmed.substring(0, splitIndex).trim();
      const value = trimmed.substring(splitIndex + 1).trimStart();
      if (key === 'aiDescription') {
        currentTag.aiDescription = unescapeTagText(value);
      } else if (key === 'aiEnabled') {
        const parsed = parseAiEnabled(value);
        if (parsed !== undefined) {
          currentTag.aiEnabled = parsed;
        }
      }
      return;
    }

    if (trimmed.startsWith('-')) {
      if (!currentClass) return;
      currentTag = {
        name: trimmed.substring(1).trim(),
      };
      if (currentTag.name) {
        importData[currentClass].push(currentTag);
      }
      return;
    }

    currentClass = trimmed;
    currentTag = null;
    if (!importData[currentClass]) {
      importData[currentClass] = [];
    }
  });

  return importData;
}

export const tagExport = async (filesBasesId: string) => {
  const result = await tagServer.tagDataByFilesBasesId(filesBasesId);
  if (!result || !result.status) {
    messageBoxAlert({
      text: '获取标签数据失败',
      type: 'error'
    })
    return
  }
  let exportStr = '// CM Collectors 标签数据 v2\n';
  exportStr += '// 分类名独占一行，标签以 - 开头，标签属性使用缩进行。\n';
  exportStr += '// 如需重命名，可写成 分类旧名=>分类新名 或 -旧标签名=>新标签名。\n\n';
  result.data.tagClass.forEach(tagClass => {
    exportStr += `${tagClass.name}\n`;
    result.data.tag.forEach(tag => {
      if (tag.tagClass_id === tagClass.id) {
        exportStr += `-${tag.name}\n`;
        exportStr += `  aiEnabled: ${tag.aiEnabled ? 'true' : 'false'}\n`;
        exportStr += `  aiDescription: ${escapeTagText(tag.aiDescription || '')}\n`;
      }
    })
  })
  const store = filesBasesStoreData()
  const filesBasesName = store.getFilesBasesNameById(filesBasesId)
  const timestamp = currentFormatDate('Ymd_His')
  await saveTextFile(exportStr, `${filesBasesName}-标签数据-${timestamp}.txt`)
}

export const tagImport = async (filesBasesId: string) => {
  const fileData = await loadTextFile()
  if (!fileData) {
    messageBoxAlert({
      text: '未选择文件',
      type: 'error'
    })
    return
  }

  const import_data = parseTagImportText(fileData.content);
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
