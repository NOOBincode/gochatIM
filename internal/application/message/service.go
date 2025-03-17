package message

import (
	"GochatIM/internal/domain/entity"
	"GochatIM/internal/infrastructure/websocket"
	"context"
)

type Service interface {
	//发送消息
	SendMessage(ctx context.Context,senderID,receiverID uint64,ContentType int,content string)([]*entity.Message,error)
	//获取消息列表
	GetMessages(ctx context.Context,ConvearsationID string,lastID string,limit int)([]*entity.Message,error)
	//处理接收到的消息
	ProcessIncomingMessage(ctx context.Context,message *websocket.Message)error
}