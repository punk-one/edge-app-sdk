package sdk

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// NATSClient NATS 客户端封装
type NATSClient struct {
	conn *nats.Conn
}

// NewNATSClient 创建 NATS 客户端
func NewNATSClient(url string) (*NATSClient, error) {
	conn, err := nats.Connect(url,
		nats.ReconnectWait(time.Second),
		nats.MaxReconnects(-1),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				fmt.Printf("NATS disconnected: %v\n", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			fmt.Printf("NATS reconnected to %v\n", nc.ConnectedUrl())
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return &NATSClient{conn: conn}, nil
}

// Publish 发布消息
func (nc *NATSClient) Publish(subject string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	return nc.conn.Publish(subject, payload)
}

// Subscribe 订阅主题
func (nc *NATSClient) Subscribe(subject string, handler func(*nats.Msg)) (*nats.Subscription, error) {
	return nc.conn.Subscribe(subject, handler)
}

// QueueSubscribe 队列订阅
func (nc *NATSClient) QueueSubscribe(subject, queue string, handler func(*nats.Msg)) (*nats.Subscription, error) {
	return nc.conn.QueueSubscribe(subject, queue, handler)
}

// Request 发送请求（RPC）
func (nc *NATSClient) Request(subject string, data interface{}, timeout time.Duration) (*nats.Msg, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	return nc.conn.Request(subject, payload, timeout)
}

// Respond 响应请求（RPC 回复）
func (nc *NATSClient) Respond(reply string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	return nc.conn.Publish(reply, payload)
}

// Close 关闭连接
func (nc *NATSClient) Close() {
	if nc.conn != nil {
		nc.conn.Close()
	}
}

// IsConnected 检查连接状态
func (nc *NATSClient) IsConnected() bool {
	return nc.conn != nil && nc.conn.IsConnected()
}

