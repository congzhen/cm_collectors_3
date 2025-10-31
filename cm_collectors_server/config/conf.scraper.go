package config

type Scraper struct {
	UseBrowserPath bool   `yaml:"useBrowserPath"`
	BrowserPath    string `yaml:"browserPath"`
	LogStatus      bool   `yaml:"logStatus"`
	LogPath        string `yaml:"logPath"`
	Headless       bool   `yaml:"headless"`
	VisitHome      bool   `yaml:"visitHome"`
	Timeout        int    `yaml:"timeout"`
}
