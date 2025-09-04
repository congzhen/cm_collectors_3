package config

type General struct {
	LogoName            string `yaml:"logoName"`
	IsAdminLogin        bool   `yaml:"isAdminLogin"`
	AdminPassword       string `yaml:"adminPassword"`
	IsAutoCreateM3u8    bool   `yaml:"isAutoCreateM3u8"`
	Language            string `yaml:"language"`
	AllowServerOpenFile bool   `yaml:"allowServerOpenFile"`
}
