package sdk

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// Client SDK 客户端
type Client struct {
	opts          Options
	nats          *NATSClient
	topics        *TopicBuilder
	startTime     time.Time
	mu            sync.RWMutex
	running       bool
	heartbeatStop chan struct{}
	logger        *logrus.Logger
	minLogLevel   LogLevel // 最小日志级别，只有大于等于此级别的日志才上报到 NATS

	// 回调函数
	heartbeatCallback HeartbeatCallback
	commandHandler    CommandHandler
	configHandler     ConfigHandler
}

// NewClient 创建新的 SDK 客户端
func NewClient(opts Options) (*Client, error) {
	if opts.NatsURL == "" {
		opts.NatsURL = "nats://127.0.0.1:4222"
	}

	// 设置默认心跳间隔
	if opts.HeartbeatInterval == 0 {
		opts.HeartbeatInterval = 30 * time.Second
	}

	// 设置默认日志级别
	if opts.LogLevel == "" {
		opts.LogLevel = "Info"
	}

	// 连接 NATS
	natsClient, err := NewNATSClient(opts.NatsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create NATS client: %w", err)
	}

	// 创建主题构建器
	topics := NewTopicBuilder(opts.AppKey)

	// 初始化 logrus
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(stringToLogrusLevel(opts.LogLevel))

	// 设置最小日志级别
	minLogLevel := LogLevel(opts.LogLevel)

	client := &Client{
		opts:          opts,
		nats:          natsClient,
		topics:        topics,
		startTime:     time.Now(),
		running:       true,
		heartbeatStop: make(chan struct{}),
		logger:        logger,
		minLogLevel:   minLogLevel,
	}

	// 初始化各个模块
	if err := client.initHeartbeat(); err != nil {
		return nil, fmt.Errorf("failed to init heartbeat: %w", err)
	}
	if err := client.initCommands(); err != nil {
		return nil, fmt.Errorf("failed to init commands: %w", err)
	}
	if err := client.initConfig(); err != nil {
		return nil, fmt.Errorf("failed to init config: %w", err)
	}

	// 启动心跳
	go client.startHeartbeat(opts.HeartbeatInterval)

	// 优雅关闭
	go client.handleShutdown()

	return client, nil
}

// SetHeartbeatCallback 设置心跳回调函数
func (c *Client) SetHeartbeatCallback(callback HeartbeatCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.heartbeatCallback = callback
}

// OnCommand 注册命令处理函数
func (c *Client) OnCommand(handler CommandHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.commandHandler = handler
}

// OnConfig 注册配置更新处理函数
func (c *Client) OnConfig(handler ConfigHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.configHandler = handler
}

// GetLogger 获取 logrus logger 实例
func (c *Client) GetLogger() *logrus.Logger {
	return c.logger
}

// SetMinLogLevel 设置最小日志级别（只有大于等于此级别的日志才上报到 NATS）
func (c *Client) SetMinLogLevel(level LogLevel) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.minLogLevel = level
}

// LogTrace 上报 Trace 级别日志
func (c *Client) LogTrace(message string) {
	c.logger.Trace(message)
	c.log(LogLevelTrace, message)
}

// LogDebug 上报 Debug 级别日志
func (c *Client) LogDebug(message string) {
	c.logger.Debug(message)
	c.log(LogLevelDebug, message)
}

// LogInfo 上报 Info 级别日志
func (c *Client) LogInfo(message string) {
	c.logger.Info(message)
	c.log(LogLevelInfo, message)
}

// LogWarn 上报 Warn 级别日志
func (c *Client) LogWarn(message string) {
	c.logger.Warn(message)
	c.log(LogLevelWarn, message)
}

// LogError 上报 Error 级别日志
func (c *Client) LogError(message string) {
	c.logger.Error(message)
	c.log(LogLevelError, message)
}

// LogFatal 上报 Fatal 级别日志
func (c *Client) LogFatal(message string) {
	c.logger.Fatal(message)
	c.log(LogLevelFatal, message)
}

// LogPanic 上报 Panic 级别日志
func (c *Client) LogPanic(message string) {
	c.logger.Panic(message)
	c.log(LogLevelPanic, message)
}

// log 内部日志上报方法（只有大于等于配置级别的日志才上报到 NATS）
func (c *Client) log(level LogLevel, message string) {
	if !c.isRunning() {
		return
	}

	// 检查日志级别，只有大于等于配置的级别才上报
	c.mu.RLock()
	minLevel := c.minLogLevel
	c.mu.RUnlock()

	if !shouldReportLog(level, minLevel) {
		return
	}

	logData := LogData{
		Level:     string(level),
		Message:   message,
		Timestamp: time.Now().Unix(),
	}

	if err := c.nats.Publish(c.topics.Logs(), logData); err != nil {
		c.logger.Errorf("Failed to publish log to NATS: %v", err)
	}
}

// EmitEvent 上报事件
func (c *Client) EmitEvent(event string, data map[string]interface{}) {
	if !c.isRunning() {
		return
	}

	eventData := EventData{
		Event:     event,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	if err := c.nats.Publish(c.topics.Events(), eventData); err != nil {
		c.logger.Errorf("Failed to publish event: %v", err)
	}
}

// ReportStatus 上报状态
func (c *Client) ReportStatus(data map[string]interface{}) {
	if !c.isRunning() {
		return
	}

	statusData := StatusData{
		AppKey:    c.opts.AppKey,
		Status:    "running",
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	if err := c.nats.Publish(c.topics.Status(), statusData); err != nil {
		c.logger.Errorf("Failed to publish status: %v", err)
	}
}

// Close 关闭客户端
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return nil
	}

	c.running = false
	close(c.heartbeatStop)

	if c.nats != nil {
		c.nats.Close()
	}

	return nil
}

// isRunning 检查是否运行中
func (c *Client) isRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

// GetUptime 获取运行时长（秒）
func (c *Client) GetUptime() int64 {
	return int64(time.Since(c.startTime).Seconds())
}

// handleShutdown 处理优雅关闭
func (c *Client) handleShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	c.logger.Info("Received shutdown signal, closing client...")
	c.Close()
	os.Exit(0)
}
