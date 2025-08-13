package datatype

type ResDataList struct {
	DataList any   `json:"dataList"  `
	Total    int64 `json:"total"  `
}

type E_resourceMode string

const (
	E_resourceMode_Movies    E_resourceMode = "movies"
	E_resourceMode_Comic     E_resourceMode = "comic"
	E_resourceMode_Atlas     E_resourceMode = "atlas"
	E_resourceMode_Files     E_resourceMode = "files"
	E_resourceMode_VideoLink E_resourceMode = "videoLink"
	E_resourceMode_NetDisk   E_resourceMode = "netDisk"
)
