package config

type PerformerAvatarLibrary struct {
	CachePath       string `yaml:"cachePath" json:"cachePath"`
	CustomBaseURL   string `yaml:"customBaseUrl" json:"customBaseUrl"`
	DefaultStrategy string `yaml:"defaultStrategy" json:"defaultStrategy"`
}
