package sdk

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nats-io/nats.go"
	"gopkg.in/yaml.v3"
)

// initConfig 初始化配置模块
func (c *Client) initConfig() error {
	// 订阅配置下发主题
	_, err := c.nats.Subscribe(c.topics.ConfigSet(), func(msg *nats.Msg) {
		c.handleConfigUpdate(msg)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to config topic: %w", err)
	}

	return nil
}

// handleConfigUpdate 处理配置更新
func (c *Client) handleConfigUpdate(msg *nats.Msg) {
	var configData ConfigData
	if err := json.Unmarshal(msg.Data, &configData); err != nil {
		c.LogError(fmt.Sprintf("Failed to unmarshal config: %v", err))
		return
	}

	// 保存配置文件
	configPath := c.getConfigPath()
	if err := c.saveConfig(configPath, configData.Config); err != nil {
		c.LogError(fmt.Sprintf("Failed to save config: %v", err))
		// 发送失败确认
		c.sendConfigAck(false, fmt.Sprintf("Failed to save config: %v", err))
		return
	}

	// 调用配置处理函数
	c.mu.RLock()
	handler := c.configHandler
	c.mu.RUnlock()

	if handler != nil {
		if err := handler(configData.Config); err != nil {
			c.LogError(fmt.Sprintf("Failed to apply config: %v", err))
			c.sendConfigAck(false, fmt.Sprintf("Failed to apply config: %v", err))
			return
		}
	}

	// 更新日志级别（如果配置中有）
	// 支持 sdk.log_level 和 log_level 两种格式
	if sdkConfig, ok := configData.Config["sdk"].(map[string]interface{}); ok {
		if logLevelStr, ok := sdkConfig["log_level"].(string); ok {
			c.SetMinLogLevel(LogLevel(logLevelStr))
			c.logger.SetLevel(stringToLogrusLevel(logLevelStr))
			c.LogInfo(fmt.Sprintf("Log level updated to: %s", logLevelStr))
		}
	} else if logLevelStr, ok := configData.Config["log_level"].(string); ok {
		// 兼容旧格式
		c.SetMinLogLevel(LogLevel(logLevelStr))
		c.logger.SetLevel(stringToLogrusLevel(logLevelStr))
		c.LogInfo(fmt.Sprintf("Log level updated to: %s", logLevelStr))
	}

	// 发送成功确认
	c.sendConfigAck(true, "Config updated successfully")
	c.LogInfo("Config updated successfully")
}

// saveConfig 保存配置到文件（YAML 格式）
func (c *Client) saveConfig(path string, config map[string]interface{}) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 序列化为 YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 写入文件（原子写入）
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// 原子替换
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to rename config file: %w", err)
	}

	return nil
}

// sendConfigAck 发送配置确认
func (c *Client) sendConfigAck(success bool, message string) {
	ack := map[string]interface{}{
		"success":   success,
		"message":   message,
		"timestamp": 0, // 可以添加时间戳
	}

	if err := c.nats.Publish(c.topics.ConfigAck(), ack); err != nil {
		c.LogError(fmt.Sprintf("Failed to send config ack: %v", err))
	}
}

// getConfigPath 获取配置文件路径
func (c *Client) getConfigPath() string {
	// 默认路径
	return fmt.Sprintf("/usr/local/edge/apps/%s/config.yaml", c.opts.AppKey)
}

// LoadConfig 加载配置文件（YAML 格式）
func (c *Client) LoadConfig() (map[string]interface{}, error) {
	configPath := c.getConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

