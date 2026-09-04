---
name: kg
description: >-
  如何使用 `confx/pkg/types/kg` 生成结构化 key.
  覆盖 creator/audience/domain/biz 四段约定与 prefix 注入策略.
  当需要统一缓存 key 命名, 跨服务共享 key, 或迁移旧 key 约定时使用.
---

# KG

本文描述
- `pkg/types/kg` 的职责边界与使用方式
- 常见 key 生成模式与注意事项

## 边界

`kg` 只负责 key 命名结构，不负责存储实现。

- key 结构: `{creator}:{audience}:{domain}:{biz}`
- 可选前缀: `{prefix}:{creator}:{audience}:{domain}:{biz}`
- 不内置业务语义，不拼接业务域，不感知站点/租户

## 指南

参见
- `KG` 指南 [kg-guideline.md](references/kg-guideline.md)
- 代码实现 `pkg/types/kg/kgen.go`
- 测试用例 `pkg/types/kg/kgen_test.go`

## 快速用法

1. 初始化 `KeyGen`，传入 creator identity
2. 需要租户隔离时通过 `WithPrefix(...)` 注入 prefix
3. 根据场景选择:
   - `Key`: 进程独占
   - `SharedKey`: 同 creator 共享
   - `ShareTo`: 指定 audience 共享
   - `GlobalKey`: 跨服务契约

## 注意事项

- `Init` 是 once 语义，重复调用保留首次结果
- `creator` / `audience` 不能为保留字 `PEER` / `ANY`
- `NormalizeCreator` / `NormalizeAudience` 只做段规范化与保留字校验
