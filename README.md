# Edge App SDK

Edge App SDK 是一个用于帮助开发者快速将 App 接入 edge-agent 的 NATS 体系的 Go 语言 SDK。

**GitHub**: [https://github.com/punk-one/edge-app-sdk](https://github.com/punk-one/edge-app-sdk)

## 功能特性

- ✅ **自动 NATS 连接管理** - 自动连接、重连、心跳检测
- ✅ **心跳上报** - 自动定时上报应用状态和指标
- ✅ **命令接收** - 支持 start/stop/restart/config.update/snapshot 等命令
- ✅ **配置管理** - 自动接收配置更新并保存到文件
- ✅ **日志上报** - 基于 logrus，支持多级别日志上报（INFO/WARN/ERROR/DEBUG），可配置最小上报级别
- ✅ **事件上报** - 支持自定义事件上报
- ✅ **状态上报** - 支持应用状态上报
- ✅ **RPC 支持** - 支持 NATS Request-Reply 模式

## 快速开始

### 安装

```bash
go get github.com/punk-one/edge-app-sdk
```

### 基本使用

```go
package main

import (
    "log"
    "github.com/punk-one/edge-app-sdk/sdk"
)

func main() {
    // 初始化 SDK
    client, err := sdk.NewClient(sdk.Options{
        AppKey:     "app.camera",
        AppVersion: "1.0.3",
        NatsURL:    "nats://127.0.0.1:4222",
    })
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()

    // 注册命令处理
    client.OnCommand(func(cmd sdk.Command) sdk.CommandResult {
        switch cmd.Action {
        case "restart":
            // 处理重启命令
            return sdk.CommandResult{
                Success: true,
                Message: "Restarted successfully",
            }
        default:
            return sdk.CommandResult{
                Success: false,
                Message: "Unknown command",
            }
        }
    })

    // 上报日志
    client.LogInfo("App started")

    // 上报事件
    client.EmitEvent("app.ready", map[string]interface{}{
        "timestamp": time.Now().Unix(),
    })

    // 保持运行
    select {}
}
```

## SDK API 文档

### 初始化

SDK 初始化需要提供以下参数：

```go
client, err := sdk.NewClient(sdk.Options{
    AppKey:           string,        // App 标识，如 "app.camera"（必填）
    AppVersion:       string,        // 版本号（必填）
    NatsURL:          string,        // NATS 服务地址，默认 "nats://127.0.0.1:4222"（可选）
    HeartbeatInterval: time.Duration, // 心跳间隔，默认 30 秒（可选）
    LogLevel:         string,        // 日志级别，默认 "Info"（可选）
})
```

**参数说明**：

| 参数 | 类型 | 必填 | 说明 | 示例 |
|------|------|------|------|------|
| `AppKey` | string | 是 | App 的唯一标识符，用于构建 NATS Topic | `"app.camera"` |
| `AppVersion` | string | 是 | App 的版本号，用于心跳上报 | `"1.0.3"` |
| `NatsURL` | string | 否 | NATS 服务器地址 | `"nats://127.0.0.1:4222"` |
| `HeartbeatInterval` | time.Duration | 否 | 心跳间隔，默认 30 秒 | `30 * time.Second` |
| `LogLevel` | string | 否 | 日志级别（参考 logrus），默认 "Info" | `"Info"`, `"Debug"`, `"Warn"`, `"Error"` |

**日志级别说明**（参考 logrus 的日志级别）：

SDK 支持以下日志级别，按严重程度从低到高：

- `Trace` - 最详细的日志，通常用于调试
- `Debug` - 调试信息
- `Info` - 一般信息（默认级别）
- `Warn` - 警告信息
- `Error` - 错误信息
- `Fatal` - 致命错误，会导致程序退出
- `Panic` - 严重错误，会导致程序 panic

只有大于等于配置级别的日志才会上报到 NATS。

### 心跳

SDK 自动发送心跳，默认间隔为 30 秒（可通过 `HeartbeatInterval` 参数配置）。心跳数据包含：

- `app_key` - App 标识
- `version` - App 版本号
- `status` - 运行状态（running/stopped/error）
- `uptime` - 运行时长（秒）
- `timestamp` - 时间戳
- `metrics` - 自定义指标（通过回调函数添加）

可以设置自定义回调添加额外指标：

```go
client.SetHeartbeatCallback(func() map[string]interface{} {
    return map[string]interface{}{
        "custom_metric": 42,
        "device_count": 10,
    }
})
```

### 命令处理

```go
client.OnCommand(func(cmd sdk.Command) sdk.CommandResult {
    switch cmd.Action {
    case "start":
        return sdk.CommandResult{Success: true, Message: "Started"}
    case "stop":
        return sdk.CommandResult{Success: true, Message: "Stopped"}
    case "restart":
        return sdk.CommandResult{Success: true, Message: "Restarted"}
    case "config.update":
        // 处理配置更新
        return sdk.CommandResult{Success: true, Message: "Config updated"}
    case "snapshot":
        return sdk.CommandResult{
            Success: true,
            Data: map[string]interface{}{
                "uptime": client.GetUptime(),
            },
        }
    default:
        return sdk.CommandResult{Success: false, Message: "Unknown command"}
    }
})
```

### 配置管理

```go
client.OnConfig(func(cfg map[string]interface{}) error {
    // 应用新配置
    // 配置文件会自动保存到指定路径
    return nil
})
```

**注意**：日志级别通过 `LogLevel` 参数在初始化时设置，也可以通过 `SetMinLogLevel()` 方法动态修改。

### 日志上报

SDK 使用 logrus 作为日志库，支持多级别日志（参考 logrus 的日志级别）：

```go
client.LogTrace("Trace message")   // 最详细的日志
client.LogDebug("Debug message")   // 调试信息
client.LogInfo("Information message")  // 一般信息
client.LogWarn("Warning message")  // 警告信息
client.LogError("Error message")   // 错误信息
client.LogFatal("Fatal message")  // 致命错误
client.LogPanic("Panic message")  // 严重错误
```

**日志级别配置**：

- 默认只上报 Info 及以上级别的日志到 NATS
- 可通过 `LogLevel` 参数在初始化时设置
- 可通过 `client.SetMinLogLevel()` 方法动态设置

```go
// 设置最小日志级别（只有大于等于此级别的日志才上报到 NATS）
client.SetMinLogLevel(sdk.LogLevelWarn) // 只上报 Warn 和 Error

// 获取 logrus logger 实例
logger := client.GetLogger()
logger.SetLevel(logrus.DebugLevel) // 设置本地日志级别
```

### 事件上报

```go
client.EmitEvent("device.error", map[string]interface{}{
    "code":    502,
    "message": "Connection timeout",
})
```

### 状态上报

```go
client.ReportStatus(map[string]interface{}{
    "connected":   true,
    "device_count": 12,
    "cpu_usage":   45.2,
})
```

## NATS Topic 规范

所有主题遵循以下格式：`app.<app_key>.<type>`

### 心跳

- **Topic**: `app.<app_key>.heartbeat`
- **方向**: App → Edge-Agent
- **频率**: 默认每 30 秒（可通过 `HeartbeatInterval` 配置）
- **数据内容**: 包含 `app_key`、`version`、`status`、`uptime`、`timestamp` 和自定义 `metrics`

### 日志

- **Topic**: `app.<app_key>.logs`
- **方向**: App → Edge-Agent

### 事件

- **Topic**: `app.<app_key>.events`
- **方向**: App → Edge-Agent

### 状态

- **Topic**: `app.<app_key>.status`
- **方向**: App → Edge-Agent

### 命令

- **Topic**: `app.<app_key>.cmd`
- **方向**: Edge-Agent → App
- **模式**: Request-Reply (RPC) 或 Pub/Sub

### 命令结果

- **Topic**: `app.<app_key>.cmd.result`
- **方向**: App → Edge-Agent

### 配置下发

- **Topic**: `app.<app_key>.config.set`
- **方向**: Edge-Agent → App

### 配置确认

- **Topic**: `app.<app_key>.config.ack`
- **方向**: App → Edge-Agent

## 示例应用

完整示例请参考 [examples/simple-app/main.go](examples/simple-app/main.go)

运行示例：

```bash
cd examples/simple-app
go run main.go
```

## 项目结构

```
edge-app-sdk/
├── sdk/                    # SDK 核心代码
│   ├── client.go          # 客户端主入口
│   ├── model.go           # 数据模型定义
│   ├── nats.go            # NATS 客户端封装
│   ├── heartbeat.go       # 心跳模块
│   ├── commands.go        # 命令处理模块
│   ├── config.go          # 配置管理模块
│   ├── logging.go         # 日志模块（logrus 集成）
│   └── events.go          # 事件模块
├── examples/               # 示例应用
│   └── simple-app/        # 简单示例
├── go.mod                  # Go 模块定义
└── README.md              # 本文件
```

## 依赖

- [NATS Go Client](https://github.com/nats-io/nats.go) - NATS 消息总线客户端
- [Logrus](https://github.com/sirupsen/logrus) - 结构化日志库
- [YAML v3](https://github.com/go-yaml/yaml) - YAML 配置文件解析库

## 注意事项

1. **心跳间隔**: 默认心跳间隔为 30 秒，可通过 `HeartbeatInterval` 参数配置
2. **日志级别**: 默认日志级别为 Info，可通过 `LogLevel` 参数配置（支持 Trace/Debug/Info/Warn/Error/Fatal/Panic）
3. **心跳数据**: 心跳自动包含 `app_key` 和 `version` 信息，可通过回调函数添加自定义指标
5. **优雅关闭**: SDK 会自动处理 SIGINT 和 SIGTERM 信号，实现优雅关闭
6. **连接重连**: SDK 自动处理 NATS 连接断开和重连
7. **线程安全**: 所有 SDK 方法都是线程安全的

## 开发建议

在生产环境中，建议补充以下功能：

- TLS/NKeys 认证（NATS 安全连接）
- 消息签名/验证（命令安全）
- 持久化配置存储（原子文件写入）
- 日志批量发送和背压控制
- 优雅关闭和重连策略
- 健康检查机制

## 许可证

本项目采用 [BSD 3-Clause License](LICENSE) 许可证。

## 贡献

欢迎提交 Issue 和 Pull Request！
