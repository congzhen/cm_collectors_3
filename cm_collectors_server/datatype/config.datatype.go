package datatype

type Config_ScanDisk struct {
	ScanDiskPaths         []string `json:"scanDiskPaths"`
	VideoSuffixName       []string `json:"videoSuffixName"`
	CoverPosterSuffixName []string `json:"coverPosterSuffixName"`
	CoverPosterType       int      `json:"coverPosterType"`
	AutoCreatePoster      bool     `json:"autoCreatePoster"`
	CheckPath             bool     `json:"checkPath"`
}
