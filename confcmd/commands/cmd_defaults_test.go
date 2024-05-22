package commands_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/sincospro/conf/confcmd/commands"
	"github.com/sincospro/conf/envconf"
)

func TestGoCmdGenDefaultConfigOptions(t *testing.T) {
	cmd := &commands.GoCmdGenDefaultConfigOptions{
		Output: os.TempDir(),
		Defaults: &[]*envconf.Group{
			{
				Name: "PREFIX",
				Values: map[string]*envconf.Var{
					"SomeVarName": {
						Name:  "SomeVarName",
						Value: "SomeVarValue",
					},
				},
			},
		},
	}
	t.Log(cmd.Use())
	t.Log(cmd.Short())

	err := cmd.Exec(&cobra.Command{})
	NewWithT(t).Expect(err).To(BeNil())

	filename := filepath.Join(cmd.Output, "default.yml")
	content, err := os.ReadFile(filename)
	NewWithT(t).Expect(err).To(BeNil())
	NewWithT(t).Expect(content).To(Equal([]byte("PREFIX__SomeVarName: SomeVarValue\n")))
}
