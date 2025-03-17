package asyncq

import (
	"context"
	"time"

	"github.com/bytedance/sonic"
	"github.com/hibiken/asynq"
)

type Client struct {
	client *asynq.Client
}

type Server struct{
	server *asynq.Server
	mux    *asynq.ServeMux
}

func NewClient(redisAddr string,redisDB int) *Client{
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr: redisAddr,
		DB: redisDB,
	})
	return &Client{client: client}
}

func NewServer(redisAddr string,redisDB int,concorrency int) *Server{
	server := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr: redisAddr,
			DB: redisDB,
		},
		asynq.Config{
			Concurrency: concorrency,
		},
	)
	return &Server{server: server,mux: asynq.NewServeMux(), 
	}
}

type Task interface{
	GetTypeName() string
	GetPayload()  interface{}
}

func (c *Client) EnqueueTask(ctx context.Context,task Task,opts ...asynq.Option) error {
	payload,err := sonic.Marshal(task.GetPayload())
	if err != nil {
		return err
	}
	_,err = c.client.EnqueueContext(ctx,asynq.NewTask(task.GetTypeName(),payload),opts...)
	if err != nil{
		return err
	}
	return nil
}


func (c *Client) EnqueueTaskIn(ctx context.Context,task Task,delay time.Duration,opts ...asynq.Option) error {
	opts = append(opts, asynq.ProcessIn(delay))
	err := c.EnqueueTask(ctx,task,opts...)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

type HandlerFunc func(context.Context,*asynq.Task) error

func (s *Server) RegisterHandler(taskType string,handler HandlerFunc) {
	s.mux.HandleFunc(taskType,handler)
}

func (s *Server) Start() error {
	return s.server.Run(s.mux)
}

func (s *Server) Shutdown(){
	s.server.Shutdown()
}