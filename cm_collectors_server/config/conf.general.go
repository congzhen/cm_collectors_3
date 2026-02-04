package config

type General struct {
	LogoName               string         `yaml:"logoName"`
	IsAdminLogin           bool           `yaml:"isAdminLogin"`
	AdminPassword          string         `yaml:"adminPassword"`
	Language               string         `yaml:"language"`
	NotAllowServerOpenFile bool           `yaml:"notAllowServerOpenFile"`
	Theme                  string         `yaml:"theme"`
	ClosePlayCloud         bool           `yaml:"closePlayCloud"`
	ClosePlayCloudDialog   bool           `yaml:"closePlayCloudDialog"`
	PlayCloudMode          string         `yaml:"playCloudMode"`
	WindowsStartNotRunApp  bool           `yaml:"windowsStartNotRunApp"`
	TvBoxEnabled           bool           `yaml:"tvBoxEnabled"`
	VideoRateLimit         VideoRateLimit `yaml:"videoRateLimit"`
}
type VideoRateLimit struct {
	// 是否启用限流
	Enabled bool `yaml:"enabled" json:"enabled"`
	// 每秒请求次数
	RequestsPerSecond float64 `yaml:"requestsPerSecond" json:"requestsPerSecond"`
	// 桶容量
	Burst int `yaml:"burst" json:"burst"`
}
