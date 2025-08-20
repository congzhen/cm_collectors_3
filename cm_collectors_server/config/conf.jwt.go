package config

import "os"

type Jwt struct {
	ExpiresAt      int    `yaml:"expiresAt"`
	Issuer         string `yaml:"issuer"`
	Audience       string `yaml:"audience"`
	Subject        string `yaml:"subject"`
	PrivateKeyPath string `yaml:"privateKeyPath"`
	PublicKeyPath  string `yaml:"publicKeyPath"`

	privateKey string
	publicKey  string
}

/*获取私钥*/
func (j *Jwt) GetPrivateKey() string {
	if j.privateKey != "" {
		return j.privateKey
	}
	content, err := os.ReadFile(j.PrivateKeyPath)
	if err != nil {
		panic(err)
	}
	j.privateKey = string(content)
	return j.privateKey
}

/*获取公钥*/
func (j *Jwt) GetPublicKey() string {
	if j.publicKey != "" {
		return j.publicKey
	}
	content, err := os.ReadFile(j.PublicKeyPath)
	if err != nil {
		panic(err)
	}
	j.publicKey = string(content)
	return j.publicKey
}
