package datatype

type ParPaging struct {
	Page       int  `json:"page" binding:"required"`
	Limit      int  `json:"limit" binding:"required"`
	FetchCount bool `json:"fetchCount"`
}

// 请求参数 - 获取资源列表
type ReqParam_ResourcesList struct {
	ParPaging
	FilesBasesId string `json:"filesBasesId"`
}

// 请求参数 - 获取喜爱的演员列表
type ReqParam_TopPreferredPerformers struct {
	PreferredIds           []string `json:"preferredIds"`           //喜欢演员的ids
	MainPerformerBasesId   string   `json:"mainPerformerBasesId"`   //主演员集id
	ShieldNoPerformerPhoto bool     `json:"shieldNoPerformerPhoto"` //屏蔽无头像演员
	Limit                  int      `json:"limit"`                  //获取数量
}

// 请求参数 - 修改演员集
type ReqParam_UpdatePerformerBases struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Sort   int    `json:"sort"`
	Status bool   `json:"status"`
}
