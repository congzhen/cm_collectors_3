package datatype

type ParPaging struct {
	Page       int  `json:"page" binding:"required"`
	Limit      int  `json:"limit" binding:"required"`
	FetchCount bool `json:"fetchCount"`
}

// 请求参数 - 获取资源列表
type ReqParam_ResourcesList struct {
	ParPaging
	FilesBasesId string              `json:"filesBasesId"`
	SearchData   ReqParam_SearchData `json:"searchData"`
}

type ReqParam_Resource struct {
	Resource    ReqParam_ResourceBase               `json:"resource"`
	PhotoBase64 string                              `json:"photoBase64"`
	Performers  []string                            `json:"performers"`
	Directors   []string                            `json:"directors"`
	Tags        []string                            `json:"tags"`
	DramaSeries []ReqParam_resourceDramaSeries_Base `json:"dramaSeries"`
}
type ReqParam_ResourceBase struct {
	ID                string         `json:"id"`
	FilesBasesID      string         `json:"filesBases_id" binding:"required"`
	Mode              E_resourceMode `json:"mode" binding:"required"`
	Title             string         `json:"title" binding:"required"`
	IssueNumber       string         `json:"issueNumber"`
	CoverPoster       string         `json:"coverPoster"`
	CoverPosterMode   int            `json:"coverPosterMode"`
	CoverPosterWidth  int            `json:"coverPosterWidth"`
	CoverPosterHeight int            `json:"coverPosterHeight"`
	IssuingDate       string         `json:"issuingDate"`
	Country           string         `json:"country"`
	Definition        string         `json:"definition"`
	Stars             int            `json:"stars"`
	Abstract          string         `json:"abstract"`
}
type ReqParam_resourceDramaSeries_Base struct {
	ID  string `json:"id"`
	Src string `json:"src"`
}

// 请求参数 - 创建filesBases
type ReqParam_CreateFilesBases struct {
	Name                     string   `json:"name" binding:"required"`
	MainPerformerBasesId     string   `json:"mainPerformerBasesId" binding:"required"`
	RelatedPerformerBasesIds []string `json:"relatedPerformerBasesIds"`
}

// 请求参数 - 设置filesBases
type ReqParam_SetFilesBases struct {
	ID                    string              `json:"id"`
	Info                  ReqParam_FielsBases `json:"info"`
	Config                string              `json:"config"`
	MainPerformerBasesId  string              `json:"mainPerformerBasesId"`
	RelatedPerformerBases []string            `json:"relatedPerformerBases"`
}

// 请求参数 - filesBases排序
type ReqParam_FilesBasesSort struct {
	SortData []FilesBasesSort `json:"sortData"`
}

type FilesBasesSort struct {
	ID   string `json:"id"`
	Sort int    `json:"sort"`
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
	CareerPerformer   bool     `json:"careerPerformer"`
	CareerDirector    bool     `json:"careerDirector"`
}

// 请求参数 - 获取喜爱的演员列表
type ReqParam_TopPreferredPerformers struct {
	PreferredIds           []string `json:"preferredIds"`           //喜欢演员的ids
	MainPerformerBasesId   string   `json:"mainPerformerBasesId"`   //主演员集id
	ShieldNoPerformerPhoto bool     `json:"shieldNoPerformerPhoto"` //屏蔽无头像演员
	Limit                  int      `json:"limit"`                  //获取数量
}

type ReqParam_PerformerData struct {
	Performer   ReqParam_Performer `json:"performer"`
	PhotoBase64 string             `json:"photoBase64"`
}
type ReqParam_Performer struct {
	ID               string `json:"id"`
	PerformerBasesID string `json:"performerBases_id" binding:"required"`
	Name             string `json:"name" binding:"required"`
	AliasName        string `json:"aliasName"`
	Birthday         string `json:"birthday"`
	Nationality      string `json:"nationality"`
	Photo            string `json:"photo"`
	CareerPerformer  bool   `json:"careerPerformer"`
	CareerDirector   bool   `json:"careerDirector"`
	Introduction     string `json:"introduction"`
	Cup              string `json:"cup"`
	Bust             string `json:"bust"`
	Waist            string `json:"waist"`
	Hip              string `json:"hip"`
	Stars            int    `json:"stars"`
	RetreatStatus    bool   `json:"retreatStatus"`
	Status           bool   `json:"status"`
}

type ReqParam_PerformerStatus struct {
	ID     string `json:"id"`
	Status bool   `json:"status"`
}

type ReqParam_CreatePerformerBases struct {
	Name string `json:"name" binding:"required"`
}

// 请求参数 - 修改演员集
type ReqParam_UpdatePerformerBases struct {
	ID     string `json:"id" binding:"required"`
	Name   string `json:"name" binding:"required"`
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
