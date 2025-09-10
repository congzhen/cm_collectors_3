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
	ScanDiskPaths                      []string            `json:"scanDiskPaths"`
	VideoSuffixName                    []string            `json:"videoSuffixName"`
	AutoGetVideoDefinition             bool                `json:"autoGetVideoDefinition"`
	ResourceNamingMode                 ResourceNamingMode  `json:"resourceNamingMode"`
	CoverPosterMatchName               []string            `json:"coverPosterMatchName"`
	CoverPosterFuzzyMatch              bool                `json:"coverPosterFuzzyMatch"`
	CoverPosterUseRandomImageIfNoMatch bool                `json:"coverPosterUseRandomImageIfNoMatch"`
	CoverPosterSuffixName              []string            `json:"coverPosterSuffixName"`
	CoverPosterType                    int                 `json:"coverPosterType"`
	CoverPosterWidth                   int                 `json:"coverPosterWidth"`
	CoverPosterHeight                  int                 `json:"coverPosterHeight"`
	AutoCreatePoster                   bool                `json:"autoCreatePoster"`
	CheckPath                          bool                `json:"checkPath"`
	FolderToSeries                     bool                `json:"folderToSeries"`
	FolderToSeriesSort                 bool                `json:"folderToSeriesSort"`
	Nfo                                Config_ScanDisk_Nfo `json:"nfo"`
}

type Config_ScanDisk_Nfo struct {
	NfoStatus               bool     `json:"nfoStatus"`
	Roots                   []string `json:"roots"`
	Titles                  []string `json:"titles"`
	IssueNumbers            []string `json:"issueNumbers"`
	IssuingDates            []string `json:"issuingDates"`
	Abstracts               []string `json:"abstracts"`
	Tags                    []string `json:"tags"`
	TagAutoCreate           bool     `json:"tagAutoCreate"`
	PerformerNames          []string `json:"performerNames"`
	PerformerMatchAliasName bool     `json:"performerMatchAliasName"`
	PerformerAutoCreate     bool     `json:"performerAutoCreate"`
	PerformerThumbs         []string `json:"performerThumbs"`
}
