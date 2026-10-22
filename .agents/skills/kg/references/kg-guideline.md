# KG Guideline

## 目标

统一 **cache key** 命名: 可读, 可共享, 可按租户隔离.

## 结构

| 段       | 来源            | 含义                    |
|----------|-----------------|-------------------------|
| prefix   | `WithPrefix`    | site/tenant (可选)      |
| creator  | `KeyGen.Init`   | 创建 key 的服务         |
| audience | 方法或 instance | 可读范围                |
| domain   | 调用方          | 业务域                  |
| biz      | 调用方          | 域内资源 id (`KeyType`) |

**biz** 为域内资源 id, 类型 `KeyType`: **uint64 或 string**. 数值 id 用无符号整型, 不用有符号整型.

wire 示例 (无 prefix):

```text
SVC_ALPHA:PEER:SESSION:42
```

带 prefix:

```text
1050:SVC_ALPHA:PEER:SESSION:42
```

## 规范化

1. `SanitizeSegment` 处理 domain, biz, instance id
2. `NormalizeSegment` 处理 creator, 具名 audience, prefix
3. sanitize 后必填段为空 → panic

段内 `:` 与 `-` 会变为 `_`, 设计 string biz 时需知晓.

## 初始化

- 每进程一个 `KeyGen`, 启动 `KeyGen.Init(creator, opts...)`
- 多租户: `WithPrefix(site_id)`, **不要**把 site 写进 creator, 除非团队明确放弃跨站同一 creator 语义

## 生成策略

| 场景            | API                                             |
|-----------------|-------------------------------------------------|
| 进程独占        | `Key`                                           |
| 同 creator 共享 | `SharedKey`                                     |
| 定向共享        | `ShareTo`                                       |
| 跨服务          | `GlobalKey`                                     |
| 读他方          | `SharedKeyFrom` / `ShareFrom` / `GlobalKeyFrom` |

## 迁移

1. 定 creator 与 prefix 策略
2. 段内冲突字符交给 `SanitizeSegment`; biz 使用 uint64/string 资源 id
3. domain/biz 尽量保持不变
4. 必要时双写双读
