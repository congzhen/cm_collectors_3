package config

type Config struct {
	General              General              `yaml:"general"`
	Sqlite3              Sqlite3              `yaml:"sqlite3"`
	System               System               `yaml:"system"`
	Scraper              Scraper              `yaml:"scraper"`
	Play                 Play                 `yaml:"play"`
	Jwt                  Jwt                  `yaml:"jwt"`
	Mysql                Mysql                `yaml:"mysql"`
	Cache                Cache                `yaml:"cache"`
	ServerFileManagement ServerFileManagement `yaml:"serverFileManagement"`
	TaryMenu             []TaryMenu           `yaml:"taryMenu"`
}
