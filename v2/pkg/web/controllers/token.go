package controllers

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"regexp"
	"strings"
	"time"
)

const (
	defaultExpireSeconds = 60 * 60 * 1 // 默认过期时间（s）
)

var (
	mySigningKey = "JWT-ARY-STARK" // 默认秘钥（启动时自动更改）
)

type TokenData struct {
	User      string `json:"user"`
	UserRole  string `json:"userRole"`
	Workspace int    `json:"workspace"`
}

type MyCustomClaims struct {
	TokenData
	jwt.RegisteredClaims
}

func init() {
	// 生成随机的JWTSecretKey
	//if conf.RunMode == conf.Release {
	//	u, _ := uuid.NewUUID()
	//	mySigningKey = u.String()
	//}
}

// GenerateToken 生成新的Token
func GenerateToken(data TokenData) (tokenString string, err error) {
	claims := MyCustomClaims{
		data,
		jwt.RegisteredClaims{
			// Also fixed dates can be used for the NumericDate
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(defaultExpireSeconds * time.Second)),
			Issuer:    "nemo",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString([]byte(mySigningKey))
	return
}

// ValidToken 验证token
func ValidToken(tokenString string) (data *TokenData) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&MyCustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(mySigningKey), nil
		},
		jwt.WithLeeway(5*time.Second),
	)
	if err != nil || token == nil {
		return nil
	}
	if claims, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
		return &claims.TokenData
	}

	return nil
}

// GetTokenValueFromHeader 从header中提取token值
func GetTokenValueFromHeader(tokenString string) string {
	//Bearer <token> or <token>
	if strings.HasPrefix(tokenString, "Bearer") {
		tokenRegExp := regexp.MustCompile(`Bearer (.*)`)
		finalToken := tokenRegExp.FindStringSubmatch(tokenString)
		if len(finalToken) >= 2 {
			return finalToken[1]
		}
	} else {
		return tokenString
	}

	return ""
}

// SetTokenValueToHeader 生成header中的token值
func SetTokenValueToHeader(tokenString string) string {
	//Authorization: Bearer <token>
	return fmt.Sprintf("Bearer %s", tokenString)
}
