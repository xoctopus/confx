---
name: appx
description: >-
  如何用 `confx/pkg/appx` 组装可执行应用, 并接入 `confx/pkg/conf*` 基础设施组件.
  当需要搭建 confx 应用, 编写 NewAppContext/Conf/Execute, 或把基础设施组件接入宿主项目时使用.
---

# APPx

`confx/pkg/appx` 负责把配置, 基础设施组件与命令入口组装成一个可执行应用:

- `Meta`: 应用元信息(名称, 版本, 特性等), 决定配置前缀与默认路径
- `Conf`: 加载配置并驱动各组件的 `PreInit → Init → WithContext`
- `Execute`: 运行主流程 `PreRun → Serve → Main`; `Main` 自决 wait/oneshot 并统一 `Close`
- oneshot 子命令通过 `AddCommand` 挂到根命令, 与 `Serve` 共存

基础设施组件位于 `confx/pkg/conf*`, 每个组件实现 `types.Endpoint[Option]` 契约,
由 `appx` 在 `Conf` 阶段统一初始化, 在退出时统一关闭.

## 组装

参见 
- `APPx` 指南 [appx-guideline.md](references/appx-guideline.md)
- 代码示例 `example/appx/cmd/example/main.go`

## 基础组件

| 包                 | 职责                                                        |
|--------------------|-------------------------------------------------------------|
| `pkg/confethcli`   | 以太坊兼容链客户端, 内置 `EthChainID` 枚举                  |
| `pkg/confjwt`      | JWT token 生成 + 校验, context 注入                         |
| `pkg/confkafka`    | kafka 消息队列组件(占位, 目前仅引入 kafka-go)               |
| `pkg/conflogx`     | logx 日志配置: 级别 / 格式 / std 或 zap 后端 / 文件滚动输出 |
| `pkg/confotel`     | OpenTelemetry 集成: trace / metric / log processor          |
| `pkg/confpulsar`   | pulsar 消息队列组件: producer / consumer endpoint           |
| `pkg/confrabbit`   | rabbitmq 消息队列组件: producer / consumer endpoint         |
| `pkg/confrdb`      | 关系型数据库组件(目前只支持 mysql)                          |
| `pkg/confredis`    | redis 组件                                                  |
| `pkg/conftls`      | tls / x509 证书接入                                         |
| `pkg/confws`       | 可配置 WebSocket 服务端: Endpoint / ClientManager / 回调    |
| `pkg/confxxl`      | xxl-job 执行器接入: 注册 / 心跳 / Job 分发                  |

## 最小可运行骨架

参见 `example/appx/cmd/example/main.go`

最小步骤:

- NewAppContext
- Conf (PreInit → Init → WithContext)
- Execute (PreRun → Serve → Main; Main 自决 wait/oneshot 并 Close)
- 其它 oneshot 优先 AddCommand
