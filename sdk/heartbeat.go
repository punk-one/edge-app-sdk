package sdk

import (
	"time"
)

// initHeartbeat 初始化心跳模块
func (c *Client) initHeartbeat() error {
	// 心跳模块在 startHeartbeat 中启动
	return nil
}

// startHeartbeat 启动心跳发送
func (c *Client) startHeartbeat(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// 立即发送一次心跳
	c.sendHeartbeat()

	for {
		select {
		case <-ticker.C:
			if c.isRunning() {
				c.sendHeartbeat()
			}
		case <-c.heartbeatStop:
			return
		}
	}
}

// sendHeartbeat 发送心跳
func (c *Client) sendHeartbeat() {
	if !c.isRunning() || !c.nats.IsConnected() {
		return
	}

	// 构建基础心跳数据
	heartbeat := HeartbeatData{
		AppKey:    c.opts.AppKey,
		Version:   c.opts.AppVersion,
		Status:    "running",
		Timestamp: time.Now().Unix(),
	}

	// 添加默认指标
	metrics := map[string]interface{}{
		"uptime": c.GetUptime(),
	}

	// 如果有自定义回调，获取额外数据
	c.mu.RLock()
	callback := c.heartbeatCallback
	c.mu.RUnlock()

	if callback != nil {
		customData := callback()
		for k, v := range customData {
			metrics[k] = v
		}
	}

	heartbeat.Metrics = metrics

	// 发布心跳
	if err := c.nats.Publish(c.topics.Heartbeat(), heartbeat); err != nil {
		c.logger.Errorf("Failed to publish heartbeat: %v", err)
	}
}

// StartHeartbeat 手动启动心跳（兼容旧 API）
func (c *Client) StartHeartbeat(interval time.Duration) {
	go c.startHeartbeat(interval)
}

