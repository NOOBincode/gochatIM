package auth

import (
	"GochatIM/internal/infrastructure/cache"
	"GochatIM/pkg/jwt"
	"GochatIM/pkg/utils"
	"context"
	"time"
)

// TokenService Token管理服务
type TokenService struct {
	userCache cache.UserCache
}

// NewTokenService 创建Token服务
func NewTokenService(userCache cache.UserCache) *TokenService {
	return &TokenService{
		userCache: userCache,
	}
}

// GenerateAndStoreTokens 生成并存储用户Token
func (s *TokenService) GenerateAndStoreTokens(ctx context.Context, userID uint64, username string) (accessToken, refreshToken string, err error) {
	// 生成唯一的tokenID
	tokenID := utils.GenerateUUID() // 使用UUID作为唯一ID

	// 生成访问令牌
	accessToken, err = jwt.GenerateTokenWithID(userID, username, tokenID, false)
	if err != nil {
		return "", "", err
	}

	// 生成刷新令牌
	refreshToken, err = jwt.GenerateTokenWithID(userID, username, tokenID, true)
	if err != nil {
		return "", "", err
	}

	// 存储访问令牌到Redis
	accessExpiration := time.Duration(jwt.GetAccessTokenExpiration()) * time.Second
	if err := s.userCache.SetUserToken(ctx, userID, tokenID, accessToken, accessExpiration); err != nil {
		return "", "", err
	}

	// 存储刷新令牌到Redis
	refreshExpiration := time.Duration(jwt.GetRefreshTokenExpiration()) * time.Second
	if err := s.userCache.SetUserToken(ctx, userID, tokenID+":refresh", refreshToken, refreshExpiration); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ValidateToken 验证Token是否有效
func (s *TokenService) ValidateToken(ctx context.Context, token string) (*jwt.CustomClaims, error) {
	// 解析JWT获取声明
	claims, err := jwt.ParseToken(token)
	if err != nil {
		return nil, err
	}

	// 检查Redis中是否存在该Token
	storedToken, err := s.userCache.GetUserToken(ctx, claims.UserID, claims.TokenID)
	if err != nil || storedToken != token {
		return nil, err
	}

	return claims, nil
}

// RevokeToken 撤销Token
func (s *TokenService) RevokeToken(ctx context.Context, userID uint64, tokenID string) error {
	return s.userCache.RevokeUserToken(ctx, userID, tokenID)
}

// RevokeAllTokens 撤销用户所有Token
func (s *TokenService) RevokeAllTokens(ctx context.Context, userID uint64) error {
	return s.userCache.RevokeAllUserTokens(ctx, userID)
}

// RefreshAccessToken 刷新访问令牌但保留刷新令牌
func (s *TokenService) RefreshAccessToken(ctx context.Context, userID uint64, username string, tokenID string) (string, error) {
	// 生成新的访问令牌
	accessToken, err := jwt.GenerateTokenWithID(userID, username, tokenID, false)
	if err != nil {
		return "", err
	}

	// 存储新的访问令牌到Redis
	accessExpiration := time.Duration(jwt.GetAccessTokenExpiration()) * time.Second
	if err := s.userCache.SetUserToken(ctx, userID, tokenID, accessToken, accessExpiration); err != nil {
		return "", err
	}

	return accessToken, nil
}
