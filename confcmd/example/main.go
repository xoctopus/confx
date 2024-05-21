package main

import (
	"github.com/spf13/cobra"

	"github.com/sincospro/conf/confcmd"
)

type Executor struct {
	ServerExpose   uint16 `name:"" help:"server expose port"`
	PrivateKey     string `name:"" help:"server private key"`
	ClientEndpoint string `name:"" help:"client endpoint"`
	Key            string `name:"key,required" help:"query command flag by key"`
}

var _ confcmd.Executor = (*Executor)(nil)

func (c *Executor) Use() string {
	return "example"
}

func (c *Executor) Short() string {
	return "a cmd factory example"
}

func (c *Executor) Exec(cmd *cobra.Command) error {
	flag := cmd.Flags().Lookup(c.Key)
	if flag == nil {
		cmd.Printf("key %s not found\n", c.Key)
		return nil
	}
	cmd.Printf("query: %s ==> %s\n", c.Key, flag.Value.String())
	return nil
}

var (
	cmd *cobra.Command
)

func init() {
	var (
		err      error
		executor = Executor{
			ServerExpose:   10000,
			ClientEndpoint: "localhost:10001",
		}
	)
	cmd, err = confcmd.NewCommand(confcmd.EN, &executor)
	if err != nil {
		panic(err)
	}
}

func main() {
	_ = cmd.Execute()
}
