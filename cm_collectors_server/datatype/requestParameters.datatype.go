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

// 请求参数 - 设置filesBases
type ReqParam_SetFilesBases struct {
	ID                    string              `json:"id"`
	Info                  ReqParam_FielsBases `json:"info"`
	Config                string              `json:"config"`
	MainPerformerBasesId  string              `json:"mainPerformerBasesId"`
	RelatedPerformerBases []string            `json:"relatedPerformerBases"`
}

// 请求参数 - filesBases信息
type ReqParam_FielsBases struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Sort   int    `json:"sort"`
	Status bool   `json:"status"`
}

// 请求参数 - 获取演员列表
type ReqParam_PerformersList struct {
	PerformerBasesIds []string `json:"performerBasesIds"`
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

// 请求参数 - 修改TagClass
type ReqParam_TagClass struct {
	ID           string `json:"id"`
	FilesBasesID string `json:"filesBases_id"`
	Name         string `json:"name"`
	LeftShow     bool   `json:"leftShow"`
	Sort         int    `json:"sort"`
	Status       bool   `json:"status"`
}

// 请求参数 - 修改Tag
type ReqParam_Tag struct {
	ID         string `json:"id"`
	TagClassID string `json:"tagClass_id"`
	Name       string `json:"name"`
	Sort       int    `json:"sort"`
	Status     bool   `json:"status"`
}

// 请求参数 - 修改TagData排序
type ReqParam_UpdateTagDataSort struct {
	TagClassSort []TagSort `json:"tagClassSort"`
	TagSort      []TagSort `json:"tagSort"`
}
type TagSort struct {
	ID   string `json:"id"`
	Sort int    `json:"sort"`
}
