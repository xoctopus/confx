---
name: appx
description: 
  - 如何用 `github.com/xoctopus/confx/pkg/appx` 组装可执行应用
  - Meta, Conf, PreInit/PreRun/Serve, Close
  - 接入 conf* 基础设施(confotel、confrdb、confredis、confpulsar、confrabbit、confjwt、conftls)
  - 当需要搭建 confx 应用, 编写 NewAppContext/Conf/Execute, 或把基础设施组件接入宿主项目时使用
---

# APPx

本文描述
- 快速组装一个可执行应用
- 基础设施

## 组装

参见 
- `APPx` 指南 [appx-guideline.md](references/appx-guideline.md)
- 代码示例 `example/appx/cmd/example/main.go`

## 基础组件

- `pkg/confjwt`: JWT token 生成+校验
- `pkg/confotel`: OpenTelemetry集成
- `pkg/confpulsar`: pulsar消息队列组件
- `pkg/confrabbit`: rabbit消息队列组件
- `pkg/confrdb`: 关系型数据库组件(目前只支持mysql)
- `pkg/confredis`: redis组件 
- `pkg/conftls`: tls证书接入

## 最小可运行骨架

参见 `example/appx/cmd/example/main.go`

最小步骤:

- NewAppContext
- Conf (PreInit → Init → WithContext)
- Execute (PreRun → Serve → Main; Main 自决 wait/oneshot 并 Close)
- 其它 oneshot 优先 AddCommand
