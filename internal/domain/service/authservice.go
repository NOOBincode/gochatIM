package service

import (
	"GochatIM/internal/domain/repo/user"
	"GochatIM/internal/infrastructure/auth"
	"context"
	"errors"
)

// AuthService 认证服务
type AuthService struct {
	userRepo     user.Repository
	tokenService *auth.TokenService
}

// NewAuthService 创建认证服务
func NewAuthService(userRepo user.Repository, tokenService *auth.TokenService) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		tokenService: tokenService,
	}
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, username, password string) (accessToken, refreshToken string, err error) {
	// 验证用户凭据
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return "", "", err
	}

	// 验证密码
	if !user.VerifyPassword(password) {
		return "", "", errors.New("密码错误")
	}

	// 生成并存储Token
	return s.tokenService.GenerateAndStoreTokens(ctx, user.ID, user.Username)
}

// Logout 用户登出
func (s *AuthService) Logout(ctx context.Context, userID uint64, tokenID string) error {
	return s.tokenService.RevokeToken(ctx, userID, tokenID)
}

// RefreshAccessToken 刷新访问令牌
func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	// 验证刷新令牌
	claims, err := s.tokenService.ValidateToken(ctx, refreshToken)
	if err != nil {
		return "", err
	}

	// 确保是刷新令牌而不是访问令牌
	if !claims.IsRefresh {
		return "", errors.New("令牌类型错误，不是刷新令牌")
	}

	// 检查用户是否存在
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return "", err
	}

	// 生成新的访问令牌，但保留原来的刷新令牌
	newAccessToken, err := s.tokenService.RefreshAccessToken(ctx, claims.UserID, user.Username, claims.TokenID)
	if err != nil {
		return "", err
	}

	return newAccessToken, nil
}
