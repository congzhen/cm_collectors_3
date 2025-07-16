package errorMessage

var (
	Err_Resources_Not_Found           = NewErrorData(2000, "资源不存在")
	Err_Resources_ID_Empty            = NewErrorData(2001, "资源ID为空")
	Err_Resources_Save_Photo_Failed   = NewErrorData(2002, "保存资源封面失败")
	Err_Resources_Delete_Photo_Failed = NewErrorData(2003, "删除资源封面失败")

	Err_performer_Save_Photo_Failed   = NewErrorData(3002, "保存演员照片失败")
	Err_performer_Delete_Photo_Failed = NewErrorData(3003, "删除演员照片失败")
)
