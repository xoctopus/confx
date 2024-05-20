package cmd

import "github.com/spf13/cobra"

type GoCmdDefaultConfigGenerator struct {
	EnableCgo       bool     `cmd:""`
	CLibDirectories []string `cmd:""`
	CLibNames       []string `cmd:""`
	OutputDIR       string   `cmd:""`
}

func (v *GoCmdDefaultConfigGenerator) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "default",
		Short: "generate go cmd default config file",
	}
	return cmd
}

/*
func (app *AppCtx) GenerateDockerfileTemplate() error {
	_ = path.Join(app.root, "./Dockerfile.default")
	return nil
}

func (app *AppCtx) GenerateMakefileTemplate() error {
	_ = path.Join(app.root, "./Makefile.default")
	return nil
}

func (app *AppCtx) GenerateDefaultConfigYML(defaults ...*envconf.Group) error {
	m := map[string]string{RuntimeKey: app.Runtime.String()}
	for _, vars := range defaults {
		for _, v := range vars.Values {
			// todo if optional or required?
			m[v.GroupName(vars.Name)] = v.Value
		}
	}

	content, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	filename := path.Join(app.root, "./config/default.yml")
	if dir := path.Dir(filename); dir != "" {
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}
	return os.WriteFile(filename, content, os.ModePerm)
}
*/
