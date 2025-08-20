package core

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

func initJwtCert(privateKeyPath, publicKeyPath string) error {
	if !jwtCertExist(privateKeyPath, publicKeyPath) {
		return generateJwtRSAKeyPair(privateKeyPath, publicKeyPath)
	}
	return nil
}

func jwtCertExist(privateKeyPath, publicKeyPath string) bool {
	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat(publicKeyPath); os.IsNotExist(err) {
		return false
	}
	return true
}

func generateJwtRSAKeyPair(privateKeyPath, publicKeyPath string) error {
	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %v", err)
	}
	// 通过私钥获取公钥
	publicKey := &privateKey.PublicKey
	// 确保私钥文件的目录存在
	if err := os.MkdirAll(filepath.Dir(privateKeyPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for private key: %v", err)
	}

	// 保存私钥到文件
	if err := savePEMKey(privateKeyPath, privateKey); err != nil {
		return fmt.Errorf("failed to save private key: %v", err)
	}

	// 确保公钥文件的目录存在
	if err := os.MkdirAll(filepath.Dir(publicKeyPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for public key: %v", err)
	}

	// 保存公钥到文件
	if err := savePublicPEMKey(publicKeyPath, publicKey); err != nil {
		return fmt.Errorf("failed to save public key: %v", err)
	}
	return nil
}

// savePEMKey 将私钥保存到PEM文件
func savePEMKey(fileName string, key *rsa.PrivateKey) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// 将私钥编码为PEM格式
	privateKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)
	_, err = file.Write(privateKeyPEM)
	return err
}

// savePublicPEMKey 将公钥保存到PEM文件
func savePublicPEMKey(fileName string, pubkey *rsa.PublicKey) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// 将公钥编码为PEM格式
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(pubkey)
	if err != nil {
		return err
	}

	publicKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: publicKeyBytes,
		},
	)
	_, err = file.Write(publicKeyPEM)
	return err
}
