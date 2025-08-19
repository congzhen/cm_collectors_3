package utils

import (
	"cm_collectors_server/datatype"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtUserToken struct {
	PrivateKey string
	PublicKey  string
	ExpiresAt  int
	Issuer     string
	Audience   string
	Subject    string
}

type UserTokenCustomClaims struct {
	UserId   string            `json:"userId"`
	UserType datatype.UserType `json:"userType"`
	jwt.RegisteredClaims
}

func (jwtUserToken *JwtUserToken) CreateToken(userId string, userType datatype.UserType) (string, error) {
	if userType == "" {
		userType = datatype.ENUM_UserType_Normal
	}
	if jwtUserToken.ExpiresAt == 0 {
		jwtUserToken.ExpiresAt = 86400
	}
	if jwtUserToken.Issuer == "" {
		jwtUserToken.Issuer = "demo"
	}
	if jwtUserToken.Audience == "" {
		jwtUserToken.Audience = "demo"
	}
	if jwtUserToken.Subject == "" {
		jwtUserToken.Subject = "somebody"
	}

	now := time.Now()
	expirationTime := now.Add(time.Duration(jwtUserToken.ExpiresAt) * time.Second)
	claims := UserTokenCustomClaims{
		userId,
		userType,
		jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    jwtUserToken.Issuer,
			Subject:   jwtUserToken.Subject,
			ID:        "1",
			Audience:  []string{jwtUserToken.Audience},
		},
	}
	then := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	privateKey, errParseRSAPrivateKeyFromPEM := jwt.ParseRSAPrivateKeyFromPEM([]byte(jwtUserToken.PrivateKey))
	if errParseRSAPrivateKeyFromPEM != nil {
		return "", errParseRSAPrivateKeyFromPEM
	}
	token, errSignedString := then.SignedString(privateKey)
	if errSignedString != nil {
		return "", errSignedString
	}
	return token, errSignedString
}

func (jwtUserToken *JwtUserToken) ParseToken(tokenStr string) (*UserTokenCustomClaims, error) {
	myCustomClaims := UserTokenCustomClaims{}
	publickey, errParseRSAPublicKeyFromPEM := jwt.ParseRSAPublicKeyFromPEM([]byte(jwtUserToken.PublicKey))
	if errParseRSAPublicKeyFromPEM != nil {
		return &myCustomClaims, errParseRSAPublicKeyFromPEM
	}
	token, errParseWithClaims := jwt.ParseWithClaims(tokenStr, &myCustomClaims, func(token *jwt.Token) (interface{}, error) {
		return publickey, nil
	})
	/*
		t, _ := token.Claims.GetExpirationTime()
		fmt.Println("Time: ", t)
	*/
	if errParseWithClaims != nil {
		return &myCustomClaims, errParseWithClaims
	}
	if claims, ok := token.Claims.(*UserTokenCustomClaims); ok && token.Valid {
		//fmt.Printf("%v %v", claims.Foo, claims.RegisteredClaims.Issuer)
		return claims, nil
	} else {
		return &myCustomClaims, errParseWithClaims
	}
}
