# APPx指南

本文档说明 `appx` 的职责契约, 运行步骤和组合方式

以下均以 `example/appx/cmd/example/main.go` 为例

与包文档 `pkg/appx/doc.go` 对齐: 职责边界为
`PreInit` → `Init` → `WithContext` → `PreRun` → `Serve` → `Main`

本版约定: **偏 serve**; oneshot 优先用 `AddCommand`. `Main` 自行决定等待/退出方式,
框架不根据 `Serves` 是否为空推断 Main 行为.

## APPx 职能

`appx` 是可执行应用的运行时骨架, 把构建身份、配置组件、命令行和生命周期收口到一个 `AppCtx` 上:

- 构建身份: 持有 `Meta` (Name / Feature / Version / CommitID / Date / Runtime), 提供 `version` 输出
- 配置加载: 合并 `config/local.yml` 与环境变量到配置结构, 并写出 `config/default.yml` 与 `config/.env`
- 组件编排: `Conf` 先跑 `PreInit`, 再按契约对配置树做 `Init`, 并把 `Injectable` 叠进返回的 `ctx`
- 命令行: cobra 根命令, 默认 `run` / `version`, 可用 `AddCommand` 扩展
- 运行与关闭: `PreInit` 在 Init 前准备元数据/hooks; `PreRun` 用完整 ctx 做进程级准备; `Serve` 并行长驻; `Main` 决定进程何时结束; `Close` 关闭 Conf 组件和 `WithClose` 回调

## APPx 组成

- 构建信息
  + `example/appx/cmd/example.Name` 服务或应用的名称
  + `example/appx/cmd/example.Feature` 特性 构建时对应 `git branch`
  + `example/appx/cmd/example.Version` 版本 构建时对应 `git tag`
  + `example/appx/cmd/example.CommitID` 提交hash
  + `example/appx/cmd/example.Date` 构建时间戳
  + `example/appx/cmd/example.meta` `Meta`

- 组件配置
  + `example/appx/cmd/example.config`: 通常匿名(不可导出)结构体去组合需要的组件
  + 需要的配置可以使用 `pkg/conf...` 或自行定义
  + 组件的生命周期
    * `SetDefault`: 只赋予默认值, 不做其他转换和操作. 如果组件实现该接口则被调用.
    * `Init`: 资源就位, 不开业务流量. 比如 `confrdb.EndpointMySQL.Init(ctx)` 会根据配置初始化数据库连接并进行活跃探测
    * `Close` / `Shutdown`: 优雅关闭组件

- 应用实体 `example/appx/cmd/example.app`

- 全局上下文 `example/appx/cmd/example.ctx + cancel` 控制整个 `app` 的生命周期

- 主入口 `example/appx/cmd/example.Main`

## APPx 搭建

- 强行重写默认配置或外部配置: 这个部分非必须, 是临时 `hack` 或者 hard-coding 的部分
- 初始化 `Meta`. 这个部分必须.
- 新建 `AppCtx`实例. 传入选项:
  + 主入口: `example/appx/cmd/example.Main` — 在 `Serve` 之后执行; **自行决定** 是等待 (signal / Serves 结束) 还是 oneshot 退出, 并负责 `Close`
  + `WithMeta`: 构建信息选项.
  + `WithRoot`: 应用根目录选项. 用于解析 `config/` 下文件; 默认是当前工作目录.
  + `WithPreInit`: 在 `Conf` 的组件 `Init` 之前顺序执行. 让组件能 Init (元数据、hooks), 不开业务流量.
  + `WithPreRun`: 在 `run` 上、`Serve` 之前顺序执行. 用完整 ctx 做进程级准备. 比如全局配置初始化、全局上下文注入等.
  + `WithServe`: 在 `PreRun` 之后并行执行. 长驻服务 (推荐主范式). 比如 http 服务、定时任务等.
  + `WithClose`: 这一系列方法会在执行 `AppCtx.Close` 被顺序调用, 去优雅关闭和资源回收.

## APPx 生命周期

职责边界:

- `PreInit`: 让组件能 Init (元数据、hooks)
- `Init`: 资源就位, 不开业务流量
- `WithContext`: 能力进 ctx
- `PreRun`: 用完整 ctx 做进程级准备
- `Serve`: 长驻对外服务 (推荐)
- `Main`: 进程壳 / 入口收口; 行为由调用方决定 (wait 或 oneshot), 必须 `Close`

步骤:

- `app = appx.NewAppContext(MainEntry, options...)`: 实例化 `appx`.
- `app.Conf`: 加载配置 → 写出 defaults → 执行 `PreInit` → `Init` 依赖组件并注入 ctx. 如: 数据库, 缓存, 消息队列等.
- `app.Execute`: 命令行 `<app_name> run` 为执行入口
- 执行 `PreRun` (`PreRunners`): 顺序同步执行.
- 执行 `Serve` (`Serves`): 执行 `PreRun` 之后, 并行异步执行(不会阻塞)
- 运行 `MainEntry`: 框架只调用; 是否等待 Serves / signal、何时退出, 由 Main 自己决定
- `app.Close(ctx)`: 顺序执行 **组件关闭** 和 `WithClose` 回调. 必须在 `MainEntry` 调用

## APPx Conf契约

`AppCtx.Conf(ctx, configurations...) context.Context` 在写完默认配置之后, 先跑 `PreInit`, 再对配置树做前序遍历.

- `AppCtx.Conf` 做三件事情
  + `PreInit`: 让组件能 Init
  + 组件初始化 (`Init`)
  + 全局上下文注入 (`WithContext` / `Injectable`)

- `AppCtx.Conf` 如何编排初始化和注入流程
  + 写出 `config/default.yml` 与 `config/.env` 之后, 顺序执行 `WithPreInit` (`AppOption.PreInit`)
  + 注入 `WithAppMeta`
  + 顺序即编排: 按照参数 `configurations` 顺序; 结构体则按照定义字段顺序 (会忽略未导出字段/类型).
    * 如果能 `Init` 则 `Init(ctx)`. `ErrSkipInitializing`, 无具体值(nil) 视为跳过, 后续步骤不再执行; 其他错误直接 panic.
    * 如果实现了 `Injectable` 则注入上下文. 多类型单例会被后者覆盖 (如 多个 `confredis.Endpoint`)
  + 最终所有组件就绪, 并返回 `全局上下文`.

## 其他补充说明

- `WithClose`: 无需注册组件的关闭和回收, 只需要做业务其他的回收操作. 所有关闭会在 `AppCtx.Close` 被调用
- 主入口必须执行关闭操作: `defer app.Close(ctx)` `defer cancel()`
- `<app_name> run`: 默认长期进程入口 (`PreRun` → `Serve` → `Main`)
- `<app_name> version`: 用于输出版本信息, 会被默认添加.
- oneshot / 工具型入口: 优先 `AppCtx.AddCommand` 注册子命令, 不要依赖「清空 Serves 让 Main 变命令」
