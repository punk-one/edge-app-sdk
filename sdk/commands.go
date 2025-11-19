package sdk

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// initCommands 初始化命令处理模块
func (c *Client) initCommands() error {
	// 订阅命令主题
	_, err := c.nats.Subscribe(c.topics.Command(), func(msg *nats.Msg) {
		c.handleCommand(msg)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to command topic: %w", err)
	}

	return nil
}

// handleCommand 处理接收到的命令
func (c *Client) handleCommand(msg *nats.Msg) {
	var cmd Command
	if err := json.Unmarshal(msg.Data, &cmd); err != nil {
		c.LogError(fmt.Sprintf("Failed to unmarshal command: %v", err))
		return
	}

	// 执行命令处理函数
	c.mu.RLock()
	handler := c.commandHandler
	c.mu.RUnlock()

	var result CommandResult
	if handler != nil {
		result = handler(cmd)
	} else {
		// 默认处理
		result = c.defaultCommandHandler(cmd)
	}

	// 设置命令 ID 和时间戳
	result.CommandID = cmd.CommandID
	if result.Timestamp == 0 {
		result.Timestamp = time.Now().Unix()
	}

	// 如果有回复主题，发送回复（RPC 模式）
	if msg.Reply != "" {
		if err := c.nats.Respond(msg.Reply, result); err != nil {
			c.LogError(fmt.Sprintf("Failed to respond to command: %v", err))
		}
	} else {
		// 否则发布到结果主题
		if err := c.nats.Publish(c.topics.CommandResult(), result); err != nil {
			c.LogError(fmt.Sprintf("Failed to publish command result: %v", err))
		}
	}
}

// defaultCommandHandler 默认命令处理
func (c *Client) defaultCommandHandler(cmd Command) CommandResult {
	switch cmd.Action {
	case "start":
		return CommandResult{
			Success: true,
			Message: "App is already running",
		}
	case "stop":
		// 注意：实际停止操作应该由外部 systemd 管理
		return CommandResult{
			Success: true,
			Message: "Stop command received (managed by systemd)",
		}
	case "restart":
		return CommandResult{
			Success: true,
			Message: "Restart command received (managed by systemd)",
		}
	case "snapshot":
		return CommandResult{
			Success: true,
			Message: "Snapshot command received",
			Data: map[string]interface{}{
				"uptime":  c.GetUptime(),
				"version": c.opts.AppVersion,
				"status":  "running",
			},
		}
	default:
		return CommandResult{
			Success: false,
			Message: fmt.Sprintf("Unknown command: %s", cmd.Action),
		}
	}
}

