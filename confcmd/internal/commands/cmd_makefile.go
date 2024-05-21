package commands

type GoCmdMakefileGenerator struct {
	EnableLinkOptimize    bool     `name:"enable-link-optimize"    help.en:"enable link optimize. use '-s -w' to trim target size. default true"                help.zh:"开启链接优化. 将使用'-s -w'参数压缩目标大小. 默认开启"`
	EnableCompileOptimize bool     `name:"enable-compile-optimize" help.en:"enable compile optimize. use '-N -l' to avoid inline optimization. default false"   help.zh:"开启编译优化. 将使用'-N -l'参数关闭内联优化. 默认关闭"`
	UseStaticBuildMode    bool     `name:""`
	WithImageEntry        bool     `name:""`
	WithBuildMetaFlags    bool     `name:""`
	BuildMetaPath         string   `name:""`
	EnableCgo             bool     `name:""`
	CLibDirectories       []string `name:""`
	CLibNames             []string `name:""`
	TargetName            string   `name:""`
	OutputDIR             string   `name:""`
}
