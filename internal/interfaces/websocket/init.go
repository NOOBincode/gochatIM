package websocket

import (
	"GochatIM/internal/infrastructure/auth"
	"GochatIM/internal/infrastructure/messaging"
	"GochatIM/pkg/logger"
)

// InitWebsocket 初始化WebSocket相关组件
func InitWebsocket(kafkaProducer *messaging.KafkaProducer, authService *auth.TokenService) *Handler {
	// 创建Gateway
	gateway := NewGateway(kafkaProducer)

	// 创建Handler
	handler := NewHandler(gateway, authService)

	logger.Info("WebSocket组件初始化完成")

	return handler
}
