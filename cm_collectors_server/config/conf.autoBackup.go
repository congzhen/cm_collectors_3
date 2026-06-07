package config

type AutoBackup struct {
	Enabled                 bool   `yaml:"enabled" json:"enabled"`
	BackupPath              string `yaml:"backupPath" json:"backupPath"`
	IntervalHours           int    `yaml:"intervalHours" json:"intervalHours"`
	ResourceChangeThreshold int    `yaml:"resourceChangeThreshold" json:"resourceChangeThreshold"`
	MaxBackups              int    `yaml:"maxBackups" json:"maxBackups"`
}
