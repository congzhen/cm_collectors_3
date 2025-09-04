package datatype

type ResourceNamingMode string

const (
	ResourceNamingModeFileName     ResourceNamingMode = "fileName"
	ResourceNamingModeDirName      ResourceNamingMode = "dirName"
	ResourceNamingModeDirFileName  ResourceNamingMode = "dirFileName"
	ResourceNamingModeFullPathName ResourceNamingMode = "fullPathName"
)

type CoverPosterMatchName string

const (
	CoverPosterMatchName_cover     CoverPosterMatchName = "cover"
	CoverPosterMatchName_poster    CoverPosterMatchName = "poster"
	CoverPosterMatchName_thumb     CoverPosterMatchName = "thumb"
	CoverPosterMatchName_thumbnail CoverPosterMatchName = "thumbnail"
	CoverPosterMatchName_fanart    CoverPosterMatchName = "fanart"
	CoverPosterMatchName_fileName  CoverPosterMatchName = "fileName"
)

type Config_ScanDisk struct {
	ScanDiskPaths                      []string               `json:"scanDiskPaths"`
	VideoSuffixName                    []string               `json:"videoSuffixName"`
	ResourceNamingMode                 ResourceNamingMode     `json:"resourceNamingMode"`
	CoverPosterMatchName               []CoverPosterMatchName `json:"coverPosterMatchName"`
	CoverPosterFuzzyMatch              bool                   `json:"coverPosterFuzzyMatch"`
	CoverPosterUseRandomImageIfNoMatch bool                   `json:"coverPosterUseRandomImageIfNoMatch"`
	CoverPosterSuffixName              []string               `json:"coverPosterSuffixName"`
	CoverPosterType                    int                    `json:"coverPosterType"`
	AutoCreatePoster                   bool                   `json:"autoCreatePoster"`
	CheckPath                          bool                   `json:"checkPath"`
}
