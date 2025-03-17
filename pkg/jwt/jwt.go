package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
)

// 定义错误类型
var (
	ErrTokenExpired     = errors.New("令牌已过期")
	ErrTokenNotValidYet = errors.New("令牌尚未生效")
	ErrTokenMalformed   = errors.New("令牌格式错误")
	ErrTokenInvalid     = errors.New("无法解析的令牌")
)

// CustomClaims 自定义JWT声明结构
type CustomClaims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	// 是否为刷新token
	IsRefresh bool `json:"is_refresh"`
	TokenID   string  `json:"token_id"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT访问令牌和刷新令牌
func GenerateTokens(userID uint64, username string) (accessToken, refreshToken string, err error) {
	// 从配置中获取JWT设置
	secretKey := viper.GetString("jwt.secret_key")
	accessExpireTime := viper.GetInt64("jwt.expires_time")
	refreshExpireTime := viper.GetInt64("jwt.refresh_expires_time")
	issuer := viper.GetString("jwt.issuer")

	// 如果刷新token过期时间未设置，则默认为访问token的7倍
	if refreshExpireTime == 0 {
		refreshExpireTime = accessExpireTime * 7
	}

	// 创建访问令牌
	accessClaims := CustomClaims{
		UserID:   userID,
		Username: username,
		IsRefresh: false,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(accessExpireTime) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
			Subject:   username,
		},
	}

	// 创建刷新令牌
	refreshClaims := CustomClaims{
		UserID:   userID,
		Username: username,
		IsRefresh: true,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(refreshExpireTime) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
			Subject:   username,
		},
	}

	// 使用指定的签名方法创建签名对象
	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(secretKey))
	if err != nil {
		return "", "", err
	}

	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(secretKey))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ParseToken 解析JWT令牌
func ParseToken(tokenString string) (*CustomClaims, error) {
	secretKey := viper.GetString("jwt.secret_key")

	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, ErrTokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// 令牌过期
				return nil, ErrTokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, ErrTokenNotValidYet
			} else {
				return nil, ErrTokenInvalid
			}
		}
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}

// RefreshToken 刷新访问令牌
func RefreshToken(refreshToken string) (newAccessToken string, err error) {
	// 解析刷新令牌
	claims, err := ParseToken(refreshToken)
	if err != nil {
		return "", err
	}

	// 验证是否为刷新令牌
	if !claims.IsRefresh {
		return "", errors.New("提供的不是刷新令牌")
	}

	// 生成新的访问令牌
	accessExpireTime := viper.GetInt64("jwt.expires_time")
	issuer := viper.GetString("jwt.issuer")
	secretKey := viper.GetString("jwt.secret_key")

	// 创建新的访问令牌
	newClaims := CustomClaims{
		UserID:   claims.UserID,
		Username: claims.Username,
		IsRefresh: false,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(accessExpireTime) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
			Subject:   claims.Username,
		},
	}

	// 签名并获取完整的编码后的字符串令牌
	newAccessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims).SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return newAccessToken, nil
}

// IsTokenExpired 检查令牌是否过期
func IsTokenExpired(tokenString string) bool {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return true
	}
	return claims.ExpiresAt.Time.Before(time.Now())
}

// IsTokenAboutToExpire 检查令牌是否即将过期
func IsTokenAboutToExpire(tokenString string) bool {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return true
	}

	bufferTime := viper.GetInt64("jwt.buffer_time")
	return claims.ExpiresAt.Time.Before(time.Now().Add(time.Duration(bufferTime) * time.Second))
}

// GetUserIDFromToken 从令牌中获取用户ID
func GetUserIDFromToken(tokenString string) (uint64, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

// GetAccessTokenExpiration 获取访问令牌过期时间（秒）
func GetAccessTokenExpiration() int64 {
	return viper.GetInt64("jwt.expires_time")
}

// GetRefreshTokenExpiration 获取刷新令牌过期时间（秒）
func GetRefreshTokenExpiration() int64 {
	refreshExpireTime := viper.GetInt64("jwt.refresh_expires_time")
	accessExpireTime := viper.GetInt64("jwt.expires_time")
	
	// 如果刷新token过期时间未设置，则默认为访问token的7倍
	if refreshExpireTime == 0 {
		refreshExpireTime = accessExpireTime * 7
	}
	
	return refreshExpireTime
}

// GenerateTokenWithID 生成带有TokenID的JWT令牌
func GenerateTokenWithID(userID uint64, username string, tokenID string, isRefresh bool) (string, error) {
	// 从配置中获取JWT设置
	secretKey := viper.GetString("jwt.secret_key")
	issuer := viper.GetString("jwt.issuer")
	
	var expireTime int64
	if isRefresh {
		expireTime = GetRefreshTokenExpiration()
	} else {
		expireTime = GetAccessTokenExpiration()
	}
	
	// 创建令牌
	claims := CustomClaims{
		UserID:    userID,
		Username:  username,
		IsRefresh: isRefresh,
		TokenID:   tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireTime) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
			Subject:   username,
		},
	}
	
	// 使用指定的签名方法创建签名对象
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	
	return token, nil
}