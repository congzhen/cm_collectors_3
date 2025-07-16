package config

type Config struct {
	Sqlite3              Sqlite3              `yaml:"sqlite3"`
	System               System               `yaml:"system"`
	Jwt                  Jwt                  `yaml:"jwt"`
	Mysql                Mysql                `yaml:"mysql"`
	Cache                Cache                `yaml:"cache"`
	ServerFileManagement ServerFileManagement `yaml:"serverFileManagement"`
}
