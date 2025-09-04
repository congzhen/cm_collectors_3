package datatype

type ResourceNamingMode string

const (
	ResourceNamingModeFileName     ResourceNamingMode = "fileName"
	ResourceNamingModeDirName      ResourceNamingMode = "dirName"
	ResourceNamingModeDirFileName  ResourceNamingMode = "dirFileName"
	ResourceNamingModeFullPathName ResourceNamingMode = "fullPathName"
)

const CoverPosterMatchName_fileName string = "fileName"

type Config_ScanDisk struct {
	ScanDiskPaths                      []string           `json:"scanDiskPaths"`
	VideoSuffixName                    []string           `json:"videoSuffixName"`
	ResourceNamingMode                 ResourceNamingMode `json:"resourceNamingMode"`
	CoverPosterMatchName               []string           `json:"coverPosterMatchName"`
	CoverPosterFuzzyMatch              bool               `json:"coverPosterFuzzyMatch"`
	CoverPosterUseRandomImageIfNoMatch bool               `json:"coverPosterUseRandomImageIfNoMatch"`
	CoverPosterSuffixName              []string           `json:"coverPosterSuffixName"`
	CoverPosterType                    int                `json:"coverPosterType"`
	AutoCreatePoster                   bool               `json:"autoCreatePoster"`
	CheckPath                          bool               `json:"checkPath"`
}
