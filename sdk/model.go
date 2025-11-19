package sdk

import "time"

// Options SDK 初始化选项
type Options struct {
	AppKey           string        // App 标识，如 "app.camera"
	AppVersion       string        // 版本号
	NatsURL          string        // NATS 服务地址，如 "nats://127.0.0.1:4222"
	HeartbeatInterval time.Duration // 心跳间隔，默认 30 秒
	LogLevel         string        // 日志级别（Trace/Debug/Info/Warn/Error/Fatal/Panic），默认 Info
}

// Command 命令结构
type Command struct {
	Action    string                 `json:"action"`     // 命令动作：start, stop, restart, config.update, snapshot, action.xxx
	Payload   map[string]interface{} `json:"payload"`     // 命令负载
	CommandID string                 `json:"command_id"` // 命令 ID
}

// CommandResult 命令执行结果
type CommandResult struct {
	CommandID string                 `json:"command_id"`
	Success   bool                   `json:"success"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp int64                  `json:"timestamp"`
}

// HeartbeatData 心跳数据
type HeartbeatData struct {
	AppKey    string                 `json:"app_key"`
	Version   string                 `json:"version"`
	Status    string                 `json:"status"` // running, stopped, error
	Metrics   map[string]interface{} `json:"metrics,omitempty"`
	Timestamp int64                  `json:"timestamp"`
}

// LogData 日志数据
type LogData struct {
	Level     string `json:"level"`     // INFO, WARN, ERROR, DEBUG
	Message   string `json:"msg"`
	Timestamp int64  `json:"timestamp"`
}

// EventData 事件数据
type EventData struct {
	Event     string                 `json:"event"`
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
}

// StatusData 状态数据
type StatusData struct {
	AppKey    string                 `json:"app_key"`
	Status    string                 `json:"status"`
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
}

// ConfigData 配置数据
type ConfigData struct {
	Config    map[string]interface{} `json:"config"`
	Version   string                 `json:"version,omitempty"`
	Timestamp int64                  `json:"timestamp"`
}


// HeartbeatCallback 心跳回调函数，返回自定义数据
type HeartbeatCallback func() map[string]interface{}

// CommandHandler 命令处理函数
type CommandHandler func(cmd Command) CommandResult

// ConfigHandler 配置更新处理函数
type ConfigHandler func(cfg map[string]interface{}) error

// TopicBuilder 主题构建器
type TopicBuilder struct {
	appKey string
}

// NewTopicBuilder 创建主题构建器
func NewTopicBuilder(appKey string) *TopicBuilder {
	return &TopicBuilder{
		appKey: appKey,
	}
}

// Heartbeat 心跳主题
func (tb *TopicBuilder) Heartbeat() string {
	return "app." + tb.appKey + ".heartbeat"
}

// Logs 日志主题
func (tb *TopicBuilder) Logs() string {
	return "app." + tb.appKey + ".logs"
}

// Events 事件主题
func (tb *TopicBuilder) Events() string {
	return "app." + tb.appKey + ".events"
}

// Status 状态主题
func (tb *TopicBuilder) Status() string {
	return "app." + tb.appKey + ".status"
}

// Command 命令主题
func (tb *TopicBuilder) Command() string {
	return "app." + tb.appKey + ".cmd"
}

// CommandResult 命令结果主题
func (tb *TopicBuilder) CommandResult() string {
	return "app." + tb.appKey + ".cmd.result"
}

// ConfigSet 配置下发主题
func (tb *TopicBuilder) ConfigSet() string {
	return "app." + tb.appKey + ".config.set"
}

// ConfigAck 配置确认主题
func (tb *TopicBuilder) ConfigAck() string {
	return "app." + tb.appKey + ".config.ack"
}

