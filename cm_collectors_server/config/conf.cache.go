package config

type Cache struct {
	Mode      string     `yaml:"mode"`
	Redis     RedisCache `yaml:"redis"`
	FreeCache FreeCache  `yaml:"freeCache"`
}
type RedisCache struct {
	Host             string `yaml:"host"`
	Port             int    `yaml:"port"`
	Password         string `yaml:"password"`
	Db               int    `yaml:"db"`
	DefaultExpireSec int    `yaml:"defaultExpireSec"`
	MinIdleConns     int    `yaml:"minIdleConns"`
	PoolSize         int    `yaml:"poolSize"`
	ConnMaxIdleTime  int    `yaml:"connMaxIdleTime"`
}
type FreeCache struct {
	MaxMemoryMB      int `yaml:"maxMemoryMB"`
	DefaultExpireSec int `yaml:"defaultExpireSec"`
}
