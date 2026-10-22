---
name: kg
description: >-
  如何使用 `confx/pkg/types/kg` 生成结构化 cache key.
  覆盖四段约定, `:` 分隔, prefix 与 creator 分离, biz 为 uint64/string 资源 id, 以及 segment 规范化.
  当需要统一缓存 key 命名, 跨服务共享 key, 或迁移旧 key 约定时使用.
---

# KG

`pkg/types/kg` 只负责 **cache key 字符串** 的命名结构, 不负责存储.

## Key 结构

- 基础: `{creator}:{audience}:{domain}:{biz}`
- 带站点/租户: `{prefix}:{creator}:{audience}:{domain}:{biz}`
- 分隔符: `SegmentSeparator` (`:`)

**prefix 与 creator 分开**: prefix 表示 site/tenant; creator 表示服务 identity. 跨站读 key 时 creator 仍用服务名, prefix 随租户变化.

## biz 段

**biz** 为 domain 内 **资源 id**, 类型 `KeyType`: **`uint64` 或 `string`**.

- 数值 id 用无符号整型 (如 `uint64` 主键), 不用有符号整型.
- 字符串 code/ULID 等走 `SanitizeSegment` (空白 / `:` / `-` → `_`).

## 段规范化

- `SanitizeSegment`: trim; 空白 / `:` / `-` → `_`.
- `NormalizeSegment`: sanitize + 大写 (用于 creator, 具名 audience, prefix).

domain / biz 在生成 key 时 sanitize, sanitize 后 **len==0 则 panic**.

creator / 具名 audience 不能为空, 且不能为保留字 `PEER`, `ANY`.

## 快速用法

```go
var g kg.KeyGen
g.Init("svc-alpha", kg.WithPrefix("1050")) // site 1050, 服务 svc-alpha
k := g.SharedKey("SESSION", uint64(42))
```

| 场景             | 方法        |
|------------------|-------------|
| 进程独占 cache   | `Key`       |
| 同服务多实例共享 | `SharedKey` |
| 指定服务可读     | `ShareTo`   |
| 跨服务契约       | `GlobalKey` |
| 读他方 key       | `*From`     |

可选: `WithInstanceID` 固定 `KeyGen.Key` 的 instance audience (默认 ULID).

## 参考

- [kg-guideline.md](references/kg-guideline.md)
- `pkg/types/kg/doc.go`
- `pkg/types/kg/kgen_test.go`

## 注意

- `KeyGen.Init` 只生效一次
- biz 仅 `KeyType` (`uint64` / `string`), 不用有符号整型
