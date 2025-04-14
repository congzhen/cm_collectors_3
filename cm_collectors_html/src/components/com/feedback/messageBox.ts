import { ElMessageBox } from 'element-plus'
export interface IMessageBox {
  text: string,
  title?: string,
  type?: 'success' | 'info' | 'warning' | 'error',
  successCallBack?: () => void,
  failCallBack?: () => void,
  ok?: string,
  cancel?: string
}
export const messageBox = function (obj: IMessageBox) {
  const confirmButtonText = obj.ok ? obj.ok : 'OK';
  const cancelButtonText = obj.cancel ? obj.cancel : 'Cancel';
  ElMessageBox.confirm(
    obj.text,
    obj.title,
    {
      confirmButtonText,
      cancelButtonText,
      type: obj.type,
    },
  ).then(obj.successCallBack).catch(obj.failCallBack);
}

export const messageBoxAlert = function (obj: IMessageBox) {
  const confirmButtonText = obj.ok ? obj.ok : 'OK';
  ElMessageBox.alert(
    obj.text,
    obj.title,
    {
      confirmButtonText,
      type: obj.type,
    },
  ).then(obj.successCallBack).catch(obj.failCallBack);
}

export const messageBoxConfirm = function (obj: IMessageBox) {
  const confirmButtonText = obj.ok ? obj.ok : 'OK';
  const cancelButtonText = obj.cancel ? obj.cancel : 'Cancel';
  const title = obj.title ? obj.title : 'Warning';
  ElMessageBox.confirm(
    obj.text,
    title,
    {
      confirmButtonText,
      cancelButtonText,
      type: 'warning',
    }
  ).then(() => {
    if (obj.successCallBack) {
      obj.successCallBack();
    }

  }).catch(() => {
    if (obj.failCallBack) {
      obj.failCallBack();
    }
  })
}
