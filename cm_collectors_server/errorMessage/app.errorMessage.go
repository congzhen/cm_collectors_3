package errorMessage

var (
	Err_Resources_Not_Found           = NewErrorData(2000, "资源不存在")
	Err_Resources_ID_Empty            = NewErrorData(2001, "资源ID为空")
	Err_Resources_Save_Photo_Failed   = NewErrorData(2002, "保存资源封面失败")
	Err_Resources_Delete_Photo_Failed = NewErrorData(2003, "删除资源封面失败")

	Err_Resources_Play_DramaSeries_Not_Found = NewErrorData(2100, "源不存在")
	Err_Resources_Play_Src_Error             = NewErrorData(2101, "路径源错误")

	Err_performer_Not_Found           = NewErrorData(3000, "演员不存在")
	Err_performer_Save_Photo_Failed   = NewErrorData(3002, "保存演员照片失败")
	Err_performer_Delete_Photo_Failed = NewErrorData(3003, "删除演员照片失败")

	Err_Tag_Not_Found      = NewErrorData(4001, "标签不存在")
	Err_TagClaSS_Not_Found = NewErrorData(4002, "标签分类不存在")

	Err_No_Config_ScanDisk = NewErrorData(5000, "请先创建导入配置")

	Err_Current_Server_Has_Been_Set_To_Disallow_This_Peration = NewErrorData(9000, "当前服务器已禁止此操作")
)
