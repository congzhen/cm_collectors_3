package processors

import (
	"cm_collectors_server/core"
	"cm_collectors_server/datatype"
	"cm_collectors_server/utils"
	"errors"
	"fmt"
)

type Login struct {
}

func (t Login) AdminLogin(pwd string) (string, error) {
	if pwd != core.Config.General.AdminPassword {
		return "", fmt.Errorf("管理员密码错误")
	}
	return t.generateLoginJWT("admin", datatype.ENUM_UserType_Admin)
}

func (Login) generateLoginJWT(userId string, userType datatype.UserType) (string, error) {
	//创建jwtToken
	jwtToken := utils.JwtUserToken{
		PrivateKey: core.Config.Jwt.GetPrivateKey(),
		PublicKey:  core.Config.Jwt.GetPublicKey(),
		ExpiresAt:  core.Config.Jwt.ExpiresAt,
		Issuer:     core.Config.Jwt.Issuer,
		Audience:   core.Config.Jwt.Audience,
		Subject:    core.Config.Jwt.Subject,
	}
	return jwtToken.CreateToken(userId, userType)
}

func (Login) JWTParseToken(tokenString string) (*utils.UserTokenCustomClaims, error) {
	if tokenString == "" {
		return nil, errors.New("tokenString is empty")
	}
	jwtToken := utils.JwtUserToken{
		PrivateKey: core.Config.Jwt.GetPrivateKey(),
		PublicKey:  core.Config.Jwt.GetPublicKey(),
		ExpiresAt:  core.Config.Jwt.ExpiresAt,
		Issuer:     core.Config.Jwt.Issuer,
		Audience:   core.Config.Jwt.Audience,
		Subject:    core.Config.Jwt.Subject,
	}
	return jwtToken.ParseToken(tokenString)
}
