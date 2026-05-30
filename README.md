# xhc2_for_studying

一个用于学习 C2 通信架构的 Go 项目。它把 HTTP C2 通道、Age 密钥交换、加密消息、Beacon 任务轮询、gRPC 远程管理、implant 生成和可选 MySQL 持久化放在同一个小型代码库里，便于逐层阅读和实验。

> [!WARNING]
> 本项目仅用于授权环境下的安全研究与学习。不要在未授权系统、网络或账号中运行、部署或测试。

## 架构概览

```text
cmd/client  -- gRPC + protobuf -->  server  -- HTTP C2 -->  implant
```

- `cmd/client`: 远程操作客户端，通过 gRPC 管理 Beacon、任务和 implant 生成。
- `server`: HTTP C2 listener、gRPC 服务、本地交互控制台、配置加载和存储层。
- `implant`: 嵌入式配置加载、Age key exchange、注册、周期 check-in 和任务执行。
- `protocol`: C2 profile、消息类型、编码器、加密、nonce、任务模型和 protobuf 定义。

## 关键能力

- HTTP C2 catch-all 路由，非 C2 请求返回嵌入式 decoy 页面。
- Age 密钥交换 + ChaCha20-Poly1305 加密的 C2 消息通道。
- C2 profile 控制 URL 片段、扩展名、User-Agent、session cookie、nonce 和编码策略。
- gRPC + protobuf 远程管理接口，定义位于 `protocol/rpc/c2.proto`。
- 通过 `generate` 命令按目标平台生成带嵌入配置的 implant。
- 存储层支持默认内存存储和 `C2_MYSQL_DSN` 驱动的 MySQL 持久化。

## 目录结构

```text
.
├── cmd/client          # gRPC 远程控制客户端
├── implant             # implant 入口、HTTP client、runtime、task handlers
├── protocol            # 通信协议、加密、C2 profile、protobuf
└── server              # HTTP C2、gRPC、console、生成器、存储层
```

## 快速开始

### 1. 运行服务端

```bash
go run ./server
```

服务端会启动：

- HTTP C2 listener: `:8024`
- gRPC listener: `:8025`
- 前台本地控制台: `c2>`

也可以指定监听地址：

```bash
go run ./server -addr :9090 -rpc-addr :9091
```

`C2_TO_STUDY_ADDR` 可作为 HTTP C2 地址的环境变量默认值，但 `-addr` 参数优先级更高。

### 2. 连接远程客户端

另开一个终端：

```bash
go run ./cmd/client -addr 127.0.0.1:8025
```

常用命令：

```text
beacons
tasks
task <type> <beacon_id> [payload]
result <task_id>
generate [options]
help
exit
```

### 3. 生成 implant

在 `cmd/client` 的 `c2>` 提示符中执行：

```text
generate -server-url http://127.0.0.1:8024 -path-prefix /api/v1 -os linux -arch amd64 -out ./implant-linux
```

常用生成参数：

- `-server-url`: 写入 implant 的 C2 服务端地址。
- `-path-prefix`: 加到所有 implant C2 请求 URL 前面的固定路径前缀。
- `-interval`: Beacon check-in 间隔，单位秒。
- `-jitter`: Beacon 抖动，单位秒。
- `-os` / `-arch`: 目标 `GOOS` / `GOARCH`。
- `-out`: 本地输出路径。

未指定的生成参数会从 `server/config/server.json` 的 `generate_defaults` 读取。

### 4. 运行 implant

生成后运行对应二进制即可。开发时也可以直接运行源码版 implant，但它需要 `implant/config/implant.json` 被嵌入到二进制中；仓库里的 `implant/config/implant.example.json` 只是示例。

```bash
./implant-linux
```

## 服务端配置

服务端配置嵌入在 `server/config/server.json` 中，主要包含：

- `generate_defaults`: implant 生成默认值，包括 `server_url`、`path_prefix`、`interval`、`jitter`、`goos`、`goarch`。
- `database`: 存储后端，默认 `memory`。
- `c2_profile`: URL 片段、消息扩展名池、User-Agent、cookie 名、路径长度和编码参数。

启用 MySQL：

```bash
C2_MYSQL_DSN='user:pass@tcp(127.0.0.1:3306)/xhc2?parseTime=true' go run ./server
```

Age 服务端密钥由 `server/cryptography` 在启动时确保存在；服务端会打印当前 public key。

## 调试接口

```bash
curl http://127.0.0.1:8024/healthz
curl http://127.0.0.1:8024/debug/beacons
```

创建调试任务：

```bash
curl -X POST http://127.0.0.1:8024/debug/tasks \
  -H "Content-Type: application/json" \
  -d '{"implant_id":"<beacon_id>","type":"whoami"}'
```
