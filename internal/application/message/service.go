package message

import (
	"context"
	"time"

	"GochatIM/internal/domain/entity"
	"GochatIM/internal/interfaces/websocket"
	"GochatIM/pkg/logger"
	"GochatIM/pkg/snowflake"
	"strconv"

	"github.com/bytedance/sonic"
)

// Service 消息服务
type Service interface {
	// SendMessage 发送消息
	SendMessage(ctx context.Context, senderID, receiverID uint64, contentType int, content string) (*entity.Message, error)

	// GetMessages 获取消息列表
	GetMessages(ctx context.Context, conversationID string, lastID string, limit int) ([]*entity.Message, error)

	// ProcessIncomingMessage 处理接收到的消息
	ProcessIncomingMessage(ctx context.Context, message *websocket.Message) error
}

// ServiceImpl 消息服务实现
type ServiceImpl struct {
	messageRepo message.Repository
	gateway     *websocket.Gateway
	idGenerator *snowflake.Generator
}

// NewService 创建消息服务
func NewService(messageRepo message.Repository, gateway *websocket.Gateway, idGenerator *snowflake.Generator) Service {
	return &ServiceImpl{
		messageRepo: messageRepo,
		gateway:     gateway,
		idGenerator: idGenerator,
	}
}

// SendMessage 发送消息
func (s *ServiceImpl) SendMessage(ctx context.Context, senderID, receiverID uint64, contentType int, content string) (*entity.Message, error) {
	// 生成消息ID
	messageID := s.idGenerator.Generate()

	// 创建会话ID (较小ID在前)
	var conversationID string
	if senderID < receiverID {
		conversationID = strconv.FormatUint(senderID, 10) + "_" + strconv.FormatUint(receiverID, 10)
	} else {
		conversationID = strconv.FormatUint(receiverID, 10) + "_" + strconv.FormatUint(senderID, 10)
	}

	// 创建消息实体
	msg := &entity.Message{
		MsgID:          messageID,
		ConversationID: conversationID,
		SenderID:       senderID,
		ReceiverType:   entity.ReceiverTypeUser, // 默认为个人消息
		ReceiverID:     receiverID,
		ContentType:    contentType,
		Content:        content,
		Extra:          "",
		SendTime:       time.Now(),
		Status:         entity.MessageStatusSent,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 保存消息
	if err := s.messageRepo.Save(ctx, msg); err != nil {
		logger.Errorf("保存消息失败: %v", err)
		return nil, err
	}

	// 构造WebSocket消息
	wsMessage := &websocket.Message{
		ID:              messageID,
		Operation:       websocket.OperationReceiveMessage,
		SenderID:        senderID,
		ReceiverID:      receiverID,
		ReceiverType:    entity.ReceiverTypeUser,
		ConvearsationID: conversationID,
		ContentType:     contentType,
		Content:         content,
		Timestamp:       time.Now().Unix(),
	}

	// 序列化消息
	data, err := sonic.Marshal(wsMessage)
	if err != nil {
		logger.Errorf("序列化消息失败: %v", err)
		return msg, nil
	}

	// 如果接收者在线，直接发送
	if s.gateway.IsUserOnline(receiverID) {
		s.gateway.SendToUser(receiverID, data)
	}

	return msg, nil
}

// GetMessages 获取消息列表
func (s *ServiceImpl) GetMessages(ctx context.Context, conversationID string, lastID string, limit int) ([]*entity.Message, error) {
	return s.messageRepo.FindByConversation(ctx, conversationID, lastID, limit)
}

// ProcessIncomingMessage 处理接收到的消息
func (s *ServiceImpl) ProcessIncomingMessage(ctx context.Context, wsMessage *websocket.Message) error {
	// 创建会话ID (较小ID在前)
	var conversationID string
	if wsMessage.SenderID < wsMessage.ReceiverID {
		conversationID = strconv.FormatUint(wsMessage.SenderID, 10) + "_" + strconv.FormatUint(wsMessage.ReceiverID, 10)
	} else {
		conversationID = strconv.FormatUint(wsMessage.ReceiverID, 10) + "_" + strconv.FormatUint(wsMessage.SenderID, 10)
	}

	// 创建消息实体
	msg := &entity.Message{
		MsgID:          wsMessage.ID,
		ConversationID: conversationID,
		SenderID:       wsMessage.SenderID,
		ReceiverType:   wsMessage.ReceiverType,
		ReceiverID:     wsMessage.ReceiverID,
		ContentType:    wsMessage.ContentType,
		Content:        wsMessage.Content,
		Extra:          "",
		SendTime:       time.Unix(wsMessage.Timestamp, 0),
		Status:         entity.MessageStatusSent,
		CreatedAt:      time.Unix(wsMessage.Timestamp, 0),
		UpdatedAt:      time.Unix(wsMessage.Timestamp, 0),
	}

	// 保存消息
	if err := s.messageRepo.Save(ctx, msg); err != nil {
		logger.Errorf("保存消息失败: %v", err)
		return err
	}

	// 序列化消息
	data, err := sonic.Marshal(wsMessage)
	if err != nil {
		logger.Errorf("序列化消息失败: %v", err)
		return err
	}

	// 如果接收者在线，直接发送
	if s.gateway.IsUserOnline(wsMessage.ReceiverID) {
		s.gateway.SendToUser(wsMessage.ReceiverID, data)
	}

	return nil
}
