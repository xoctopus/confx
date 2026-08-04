# APPx指南

本文档说明 `appx` 的运行和组合方式

以下均以 `exmaple/appx/cmd/example/main.go` 为例

## APPx 组成


- 构建信息 
  + `example/appx/cmd/example.Name` 服务或应用的名称
  + `example/appx/cmd/example.Feature` 特性 构建时对应 `git branch`
  + `example/appx/cmd/example.Version` 版本 构建时对应 `git tag`
  + `example/appx/cmd/example.CommitID` 提交hash
  + `example/appx/cmd/example.Date` 构建时间戳
  + `example/appx/cmd/example.meta` `Mete`

- 组件配置
  + `example/appx/cmd/example.config`: 通常匿名(不可导出)结构体去组合需要的组件
  + 需要的配置可以使用 `pkg/conf...` 或自行定义
  + 组件的生命周期
    * `SetDefault`: 只赋予默认值, 不做其他转换和操作. 如果组件实现该接口则被调用.
    * `Init`: 初始化组件. 比如 `confrdb.EndpointMySQL.Init(ctx)` 会根据配置初始化数据库连接并进行活跃探测
    * `Close` / `Shutdown`: 优雅关闭组件

- 应用实体 `example/appx/cmd/example.app`

- 全局上下文 `example/appx/cmd/example.ctx + cancel` 控制整个 `app` 的生命周期

- 主入口 `example/appx/cmd/example.Main` 

## APPx 搭建

- 强行重写默认配置或外部配置: 这个部分非必须, 是临时 `hack` 或者 hard-coding 的部分
- 初始化 `Meta`. 这个部分必须.
- 新建 `AppCtx`实例. 传入选项:
  + 主入口: `example/appx/cmd/example.Main`
  + `WithBuildMeta`: 构建信息选项.
  + `WithMainRoot`: 主目录选项. 传入主目录路径, 默认是当前工作目录.
  + `WithPreRunner`: 这一系列方法会同步顺序执行, 通常作为 `app` 初始化行为. 比如, 全局配置初始化, 全局上下文注入等.
  + `WithServes`: 这一系列方法会并行执行, 通常作为 `app` 的长生命周期运行能力. 比如, http服务, 定时任务等.
  + `WithCloseFns`: 这一系列方法会在执行 `AppCtx.Close` 被顺序调用, 去优雅关闭和资源回收.

## APPx 生命周期

- `app = appx.NewAppContext(MainEntry, options...)`: 实例化 `appx`.
- `app.Conf`: 初始化依赖组件. 如: 数据库, 缓存, 消息队列等.
- `app.Execute`: 命令行 `<app_name> run` 为执行入口
- 执行 `PreRunners`: 顺序同步执行.
- 执行 `Serves`: 执行 `PreRunners` 之后, 并行异步执行(不会阻塞)
- 运行 `app` 实例. 即 `MainEntry`
- `app.Close(ctx)`: 顺序执行 **组件关闭** 和 `CloseFns`. 必须在 `MainEntry` 调用

## 其他补充说明

- `WithCloseFns`: 无需注册组件的关闭和回收, 只需要做业务其他的回收操作. 所有关闭会在 `AppCtx.Close` 被调用
- 主入口必须执行关闭操作: `defer app.Close(ctx)` `defer cancel()` 
- `<app_name> run`: 为执行入口.
- `<app_name> version`: 用于输出版本信息, 会被默认添加.
- 如果还需要其他指令入口, 可以使用 `AppCtx.AddCommand` 注册子命令

