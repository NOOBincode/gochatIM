package websocket

import (
	"GochatIM/internal/infrastructure/messaging"
	"GochatIM/pkg/logger"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gorilla/websocket"
)

const (
	MessageTypeText     = 1
	MessageTypeImage    = 2
	MessageTypeVoice    = 3
	MessageTypeVideo    = 4
	MessageTypeFile     = 5
	MessageTypeLocation = 6
)

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

const (
	MessageSendTopic    = "message_send"
	MessageReceiveTopic = "message_receive"
)

type Message struct {
	ID              string `json:"id"`
	Operation       int    `json:"operation"`
	Sequence        uint64 `json:"sequence"`
	SenderID        uint64 `json:"sender_id"`
	ReceiverID      uint64 `json:"receiver_id"`
	ConvearsationID string `json:"conversation_id,omitempty"`
	ContentType     int    `json:"content_type,omitempty"`
	Content         string `json:"content,omitempty"`
	Timestamp       int64  `json:"timestamp,omitempty"`
	Extra           string `json:"extra,omitempty"`
}

type Connection struct {
	UserID     uint64
	DeviceID   string
	Conn       *websocket.Conn
	Send       chan []byte
	Closed     chan struct{}
	LastActive time.Time
	mu         sync.Mutex
}

type Gateway struct {
	connections     map[uint64]map[string]*Connection
	ConnectionMutex sync.RWMutex
	//升级器
	Upgrater *websocket.Upgrader

	KafkaProducer *messaging.KafkaProducer
	//上下文
	ctx    context.Context
	cancel context.CancelFunc
}

func NewGateway(kafkaProducer *messaging.KafkaProducer) *Gateway {
	ctx, cancel := context.WithCancel(context.Background())
	return &Gateway{
		connections: make(map[uint64]map[string]*Connection),
		Upgrater: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		KafkaProducer: kafkaProducer,
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (g *Gateway) HandleWebsocket(w http.ResponseWriter, r *http.Request, userID uint64, deviceID string) {
	conn, err := g.Upgrater.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "无法升级到WebSocket连接", http.StatusInternalServerError)
		return
	}

	connection := &Connection{
		UserID:     userID,
		DeviceID:   deviceID,
		Conn:       conn,
		Send:       make(chan []byte, 256),
		Closed:     make(chan struct{}),
		LastActive: time.Now(),
	}
	g.addConnection(connection)
	go g.readPump(connection)
	go g.writePump(connection)
	connectAck := Message{
		Operation: OperationConnectAck,
		Timestamp: time.Now().Unix(),
	}
	data, _ := sonic.Marshal(connectAck)
	connection.Send <- data

	logger.Infof("用户连接成功: %d,设备ID: %s", userID, deviceID)
}

func (g *Gateway) addConnection(conn *Connection) {
	g.ConnectionMutex.Lock()
	defer g.ConnectionMutex.Unlock()
	if _, ok := g.connections[conn.UserID]; !ok {
		g.connections[conn.UserID] = make(map[string]*Connection)
	}
	if oldconn, ok := g.connections[conn.UserID][conn.DeviceID]; ok {
		close(oldconn.Send)
	}
	g.connections[conn.UserID][conn.DeviceID] = conn
}

func (g *Gateway) removeConnection(conn *Connection) {
	g.ConnectionMutex.Lock()
	defer g.ConnectionMutex.Unlock()
	if devices, ok := g.connections[conn.UserID]; ok {
		if c, ok := devices[conn.DeviceID]; ok && c == conn {
			delete(devices, conn.DeviceID)
			logger.Infof("用户断开连接: %d,设备ID: %s", conn.UserID, conn.DeviceID)
		}
		if len(devices) == 0 {
			delete(g.connections, conn.UserID)
		}
	}
}

// 在Connection结构体中，需要使用锁保护LastActive字段的访问
func (c *Connection) UpdateLastActive() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastActive = time.Now()
}

func (c *Connection) GetLastActive() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.LastActive
}

// 在readPump方法中使用UpdateLastActive
func (g *Gateway) readPump(conn *Connection) {
	defer func() {
		g.removeConnection(conn)
		conn.Conn.Close()
		close(conn.Send)
	}()

	conn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.Conn.SetPongHandler(func(string) error {
		conn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.UpdateLastActive()
		return nil
	})
	for {
		_, message, err := conn.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Errorf("读取消息错误: %v", err)
			}
			break
		}
		//处理消息
		g.handleMessage(conn, message)
		conn.UpdateLastActive()
	}
}

// 添加一个安全的发送消息方法
func (g *Gateway) SendToUser(userID uint64, message []byte) bool {
	g.ConnectionMutex.RLock()
	devices, ok := g.connections[userID]
	if !ok {
		g.ConnectionMutex.RUnlock()
		return false
	}

	// 复制设备列表，避免长时间持有读锁
	devicesCopy := make(map[string]*Connection)
	for id, conn := range devices {
		devicesCopy[id] = conn
	}
	g.ConnectionMutex.RUnlock()

	sent := false
	for _, conn := range devicesCopy {
		select {
		case conn.Send <- message:
			sent = true
		default:
			// 通道已满，可能需要关闭连接
			logger.Warnf("用户消息队列已满，无法发送消息: userID=%d", userID)
		}
	}

	return sent
}

// 添加检查用户是否在线的方法
func (g *Gateway) IsUserOnline(userID uint64) bool {
	g.ConnectionMutex.RLock()
	defer g.ConnectionMutex.RUnlock()
	devices, ok := g.connections[userID]
	return ok && len(devices) > 0
}

// 添加获取在线用户数的方法
func (g *Gateway) GetOnlineUserCount() int {
	g.ConnectionMutex.RLock()
	defer g.ConnectionMutex.RUnlock()
	return len(g.connections)
}

// 添加获取用户连接数的方法
func (g *Gateway) GetUserConnectionCount(userID uint64) int {
	g.ConnectionMutex.RLock()
	defer g.ConnectionMutex.RUnlock()
	if devices, ok := g.connections[userID]; ok {
		return len(devices)
	}
	return 0
}

func (g *Gateway) writePump(conn *Connection) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		conn.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-conn.Send:
			conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				conn.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			err := conn.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				logger.Errorf("发送消息错误: %v", err)
				return
			}
		case <-ticker.C:
			conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-conn.Closed:
			return
		}
	}
}

func (g *Gateway) handleMessage(conn *Connection, data []byte) {
	var message Message
	if err := sonic.Unmarshal(data, &message); err != nil {
		logger.Errorf("解析消息错误: %v", err)
		return
	}
	switch message.Operation {
	case OperationHeartbeat:
		HeartbeatAck := Message{
			Operation: OperationHeartbeatAck,
			Timestamp: time.Now().Unix(),
		}
		data, _ := sonic.Marshal(HeartbeatAck)
		conn.Send <- data
	case OperationSendMessage:
		message.SenderID = conn.UserID
		message.Timestamp = time.Now().Unix()

		messageKey := message.ID
		if messageKey == "" {
			messageKey = fmt.Sprintf("%d_%d_%d", message.SenderID, message.ReceiverID, message.Timestamp)
		}
		if err := g.KafkaProducer.SendMessage(MessageSendTopic, messageKey, data); err != nil {
			logger.Errorf("发送消息到Kafka失败: %v", err)
			errorMsg := Message{
				Operation: OperationMessageAck,
				ID:        message.ID,
				Sequence:  message.Sequence,
				Timestamp: time.Now().Unix(),
				Content:   "发送消息失败",
			}
			errData, _ := sonic.Marshal(errorMsg)
			conn.Send <- errData
			return
		}
		ackMsg := Message{
			Operation: OperationMessageAck,
			ID:        message.ID,
			Sequence:  message.Sequence,
			Timestamp: time.Now().Unix(),
			Content:   "消息已发送",
		}
		ackData, _ := sonic.Marshal(ackMsg)
		conn.Send <- ackData
	}
}
