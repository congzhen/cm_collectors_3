package serverfilemanagement

import "errors"

// 错误常量定义
var (
	Err_ServerFileManagement_NoRootPaths     = errors.New("未设置根路径")
	Err_ServerFileManagement_InvalidPath     = errors.New("路径无效或超出允许范围")
	Err_ServerFileManagement_NotAbsolutePath = errors.New("路径必须是绝对路径")
	Err_InvalidPermissionsFormat             = errors.New("无效权限格式")
	Err_MoveDeleteFailed                     = errors.New("移动成功但删除原文件失败")
	Err_InvalidZipEntryPath                  = errors.New("无效的ZIP条目路径")
	Err_InvalidZipFile                       = errors.New("无效的压缩文件")
	Err_SetPermissionsFailed                 = errors.New("设置权限失败")
	Err_SaveFile_TargetIsDirectory           = errors.New("目标路径是目录，无法写入文件")
	Err_CopySamePath                         = errors.New("无法复制到相同路径")
	Err_MoveSamePath                         = errors.New("无法移动到相同路径")
)
