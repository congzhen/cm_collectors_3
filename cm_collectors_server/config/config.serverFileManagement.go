package config

type ServerFileManagement struct {
	RootPath []SFM_PathEntry `yaml:"rootPath"`
}

type SFM_PathEntry struct {
	RealPath    string `yaml:"realPath"`
	VirtualPath string `yaml:"virtualPath"`
}
