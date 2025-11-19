package main

import (
	"fmt"
	"log"
	"time"

	"github.com/punk-one/edge-app-sdk/sdk"
)

func main() {
	// 初始化 SDK 客户端
	client, err := sdk.NewClient(sdk.Options{
		AppKey:           "app.simple",
		AppVersion:       "1.0.0",
		NatsURL:          "nats://127.0.0.1:4222",
		HeartbeatInterval: 30 * time.Second, // 心跳间隔，默认 30 秒
		LogLevel:         "Info",            // 日志级别，默认 Info
	})
	if err != nil {
		log.Fatalf("Failed to create SDK client: %v", err)
	}
	defer client.Close()

	fmt.Println("Simple App started successfully!")

	// 设置心跳回调（可选）
	client.SetHeartbeatCallback(func() map[string]interface{} {
		return map[string]interface{}{
			"custom_metric": 42,
			"custom_status": "healthy",
		}
	})

	// 注册命令处理
	client.OnCommand(func(cmd sdk.Command) sdk.CommandResult {
		fmt.Printf("Received command: %s\n", cmd.Action)

		switch cmd.Action {
		case "start":
			return sdk.CommandResult{
				Success: true,
				Message: "App started",
			}
		case "stop":
			return sdk.CommandResult{
				Success: true,
				Message: "App stopped",
			}
		case "restart":
			return sdk.CommandResult{
				Success: true,
				Message: "App restarted",
			}
		case "snapshot":
			return sdk.CommandResult{
				Success: true,
				Message: "Snapshot taken",
				Data: map[string]interface{}{
					"uptime":  client.GetUptime(),
					"version": "1.0.0",
				},
			}
		default:
			return sdk.CommandResult{
				Success: false,
				Message: fmt.Sprintf("Unknown command: %s", cmd.Action),
			}
		}
	})

	// 注册配置更新处理
	client.OnConfig(func(cfg map[string]interface{}) error {
		fmt.Printf("Config updated: %+v\n", cfg)
		// 在这里应用新配置
		return nil
	})

	// 上报一些日志
	client.LogInfo("Simple app initialized")
	client.LogInfo("Ready to receive commands")

	// 上报初始状态
	client.ReportStatus(map[string]interface{}{
		"initialized": true,
		"ready":       true,
	})

	// 模拟应用运行
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 定期上报状态
			client.ReportStatus(map[string]interface{}{
				"uptime":     client.GetUptime(),
				"status":     "running",
				"last_check": time.Now().Unix(),
			})

			// 上报一个事件
			client.EmitEvent("app.heartbeat", map[string]interface{}{
				"timestamp": time.Now().Unix(),
			})

			client.LogInfo("App is running normally")
		}
	}
}

