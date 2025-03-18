package model

// Message WebSocket消息结构
type Message struct {
	ID             string `json:"id"`
	Operation      int    `json:"operation"`
	Sequence       uint64 `json:"sequence"`
	SenderID       uint64 `json:"sender_id"`
	ReceiverID     uint64 `json:"receiver_id"`
	ConversationID string `json:"conversation_id,omitempty"`
	ContentType    int    `json:"content_type,omitempty"`
	Content        string `json:"content,omitempty"`
	Timestamp      int64  `json:"timestamp,omitempty"`
	Extra          string `json:"extra,omitempty"`
	ReceiverType   int    `json:"receiver_type,omitempty"`
}

// 消息类型常量
const (
	MessageTypeText     = 1
	MessageTypeImage    = 2
	MessageTypeVoice    = 3
	MessageTypeVideo    = 4
	MessageTypeFile     = 5
	MessageTypeLocation = 6
)

// 操作类型常量
const (
	OperationConnect        = 1
	OperationConnectAck     = 2
	OperationHeartbeat      = 3
	OperationHeartbeatAck   = 4
	OperationSendMessage    = 5
	OperationMessageAck     = 6
	OperationReceiveMessage = 7
	OperationDisconnect     = 8
)

// 主题常量
const (
	MessageSendTopic    = "message_send"
	MessageReceiveTopic = "message_receive"
)
