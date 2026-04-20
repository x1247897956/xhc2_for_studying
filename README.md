# xhc2_for_studying

一个用于学习 C2 基础通信流程的 Go 示例项目。项目包含一个 HTTP server 和一个 implant 客户端，用于演示 beacon 注册、周期性 check-in、任务下发和任务结果回传。

> 仅用于授权环境下的安全研究与学习。

## 项目结构

```text
.
├── implant/        # 客户端入口、配置、运行循环、任务执行逻辑
├── protocol/       # 通信协议、消息结构和错误定义
├── server/         # HTTP 服务、handler、内存存储和任务管理
├── go.mod
└── go.sum
```

## 环境要求

- Go 1.25.8 或兼容版本

## 快速开始

启动 server：

```bash
go run ./server
```

默认监听地址为 `:8024`。也可以通过参数或环境变量指定：

```bash
go run ./server -addr :8024
C2_TO_STUDY_ADDR=:8024 go run ./server
```

运行 implant：

```bash
go run ./implant
```

implant 的服务端地址配置在 `implant/config/implant.json` 中：

```json
{
  "server_url": "http://127.0.0.1:8024",
  "interval": 5,
  "jitter": 0
}
```

## 调试接口

健康检查：

```bash
curl http://127.0.0.1:8024/healthz
```

查看已注册 beacon：

```bash
curl http://127.0.0.1:8024/debug/beacons
```

创建调试任务：

```bash
curl -X POST http://127.0.0.1:8024/debug/tasks \
  -H "Content-Type: application/json" \
  -d '{"implant_id":"<beacon_id>","type":"whoami"}'
```

当前支持的任务类型包括：

- `noop`
- `whoami`
