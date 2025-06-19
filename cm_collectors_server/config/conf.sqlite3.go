package config

type Sqlite3 struct {
	Path        string `yaml:"path"`
	LoggerLevel string `yaml:"loggerLevel"`
}
