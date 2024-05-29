package commands

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/xoctopus/confx/confcmd"
	"github.com/xoctopus/confx/envconf"
)

func NewGoCmdGenDefaultConfigOptions(defaults []*envconf.Group, output string) *GoCmdGenDefaultConfigOptions {
	return &GoCmdGenDefaultConfigOptions{
		Defaults:        &defaults,
		Output:          output,
		MultiLangHelper: confcmd.NewDefaultMultiLangHelper(),
		FlagSet:         confcmd.NewFlagSet(),
	}
}

type GoCmdGenDefaultConfigOptions struct {
	Defaults *[]*envconf.Group `cmd:"-"`
	Output   string            `cmd:"output" help:"config file output dir"`

	*confcmd.MultiLangHelper
	*confcmd.FlagSet
}

func (c *GoCmdGenDefaultConfigOptions) Use() string {
	return "defaults"
}

func (c *GoCmdGenDefaultConfigOptions) Short() string {
	return "generate go cmd default config template"
}

func (c *GoCmdGenDefaultConfigOptions) Exec(cmd *cobra.Command) error {
	if err := os.MkdirAll(c.Output, os.ModePerm); err != nil {
		return errors.Wrapf(err, "failed to create output dir")
	}

	m := make(map[string]string)
	for _, g := range *c.Defaults {
		for _, v := range g.Values {
			if !v.Optional {
				m[v.GroupName(g.Name)] = v.Value
			}
		}
	}

	content, err := yaml.Marshal(m)
	if err != nil {
		return errors.Wrap(err, "failed to marshal default vars")
	}

	filename := filepath.Join(c.Output, "default.yml")
	err = os.WriteFile(filename, content, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "failed to write file")
	}

	cmd.Printf("default config generated to: %s\n", filename)
	return nil
}
