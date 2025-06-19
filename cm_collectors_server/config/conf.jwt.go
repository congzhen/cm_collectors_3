package config

import "os"

type Jwt struct {
	ExpiresAt      int    `yaml:"expiresAt"`
	Issuer         string `yaml:"issuer"`
	Audience       string `yaml:"audience"`
	Subject        string `yaml:"subject"`
	PrivateKeyPath string `yaml:"privateKeyPath"`
	PublicKeyPath  string `yaml:"publicKeyPath"`
}

/*获取私钥*/
func (j *Jwt) GetPrivateKey() string {
	content, err := os.ReadFile(j.PrivateKeyPath)
	if err != nil {
		panic(err)
	}
	return string(content)
}

/*获取公钥*/
func (j *Jwt) GetPublicKey() string {
	content, err := os.ReadFile(j.PublicKeyPath)
	if err != nil {
		panic(err)
	}
	return string(content)
}
