package config

import "strconv"

type Mysql struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Db           string `yaml:"db"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	Config       string `yaml:"config"`       //高级配置，例如charset
	MaxIdleConns int    `yaml:"maxIdleConns"` //空闲中的最大链接数
	MaxOpenConns int    `yaml:"maxOpenConns"` //打开数据库的最大大链接数
	LoggerLevel  string `yaml:"loggerLevel"`
}

/*生产环境下，mysql dsn*/
func (m Mysql) Dsn() string {
	//dsn := "root:@tcp(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local"
	return m.User + ":" + m.Password + "@tcp(" + m.Host + ":" + strconv.Itoa(m.Port) + ")/" + m.Db + "?" + m.Config
}
