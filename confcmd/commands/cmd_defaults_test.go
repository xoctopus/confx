package commands_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/xhd2015/xgo/runtime/mock"
	"gopkg.in/yaml.v3"

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

	// xgo is not patch std function?
	// t.Run("FailedToMakeOutputDir", func(t *testing.T) {
	// 	mock.Patch(os.MkdirAll, func(string, os.FileMode) error {
	// 		return errors.New(t.Name())
	// 	})
	// 	err = cmd.Exec(&cobra.Command{})
	// 	t.Log(err)
	// 	NewWithT(t).Expect(err.Error()).To(ContainSubstring(t.Name()))
	// })

	t.Run("FailedToMarshalYML", func(t *testing.T) {
		mock.Patch(yaml.Marshal, func(any) ([]byte, error) {
			return nil, errors.New(t.Name())
		})
		err = cmd.Exec(&cobra.Command{})
		t.Log(err)
		NewWithT(t).Expect(err.Error()).To(ContainSubstring(t.Name()))
	})

	t.Run("FailedToWriteFile", func(t *testing.T) {
		mock.Patch(os.WriteFile, func(string, []byte, os.FileMode) error {
			return errors.New(t.Name())
		})
		err = cmd.Exec(&cobra.Command{})
		t.Log(err)
		NewWithT(t).Expect(err.Error()).To(ContainSubstring(t.Name()))
	})
}
