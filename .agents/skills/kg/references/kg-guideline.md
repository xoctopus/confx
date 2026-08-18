# KG Guideline

## 目标

`kg` 用于统一 key 命名约定，提供稳定、可读、可共享的 key 结构。

## Key 结构

- 基础结构: `{creator}:{audience}:{domain}:{biz}`
- 带前缀结构: `{prefix}:{creator}:{audience}:{domain}:{biz}`

字段语义:

- `creator`: key 创建方身份，来自 `KeyGen.Init(creator)`
- `audience`: 受众，通常为实例、`PEER`、指定 identity 或 `ANY`
- `domain`: 业务域（由调用方定义）
- `biz`: 业务标识（如 ID、code、组合 key）
- `prefix`: 可选命名空间（如 tenant/site/env），由调用方决定

## 初始化建议

- 每个服务进程持有一个 `KeyGen` 实例
- 启动时尽早 `Init(...)`
- 如需租户隔离，启动时固定 `WithPrefix(...)`

## 生成策略建议

- 进程独占状态: `Key(domain, biz)`
- 同服务共享: `SharedKey(domain, biz)`
- 指定受众共享: `ShareTo(audience, domain, biz)`
- 跨服务契约: `GlobalKey(domain, biz)`

## 规范化与保留字

- 推荐对 identity 使用 `NormalizeCreator` / `NormalizeAudience`
- 保留字:
  - `PEER`
  - `ANY`

## 迁移建议

从旧的业务化 key 迁移到 `kg` 时:

1. 先确定 creator 与 prefix 策略
2. 保持 domain/biz 稳定，避免批量失效
3. 若需兼容旧 key，短期双写双读，逐步切换
