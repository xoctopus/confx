package cmds

import (
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/xoctopus/confx/confcmd"
)

type RunServer struct {
	*Global
	Port uint16 `help:"server listen port"`

	*confcmd.FlagSet
	*confcmd.MultiLangHelper
}

var _ confcmd.Executor = (*RunServer)(nil)

func (s *RunServer) Use() string { return "run" }

func (s *RunServer) Short() string { return "run a http server" }

func (s *RunServer) Exec(cmd *cobra.Command, args ...string) error {
	http.HandleFunc("/echo", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		defer req.Body.Close()
		content, err := io.ReadAll(req.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = w.Write(content)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	for _, f := range s.Flags() {
		cmd.Printf("%-12s %v\n", f.Name()+":", f.Value())
	}
	addr := fmt.Sprintf(":%d", s.Port)
	cmd.Printf("server started, listening %s\n", addr)
	if s.SecurityEnabled() {
		return http.ListenAndServeTLS(addr, s.CertFile, s.KeyFile, nil)
	} else {
		return http.ListenAndServe(addr, nil)
	}
}

var ServerCmd *cobra.Command

func init() {
	ServerCmd = &cobra.Command{
		Use:   "server",
		Short: "server sub command",
	}
	ServerCmd.AddCommand(
		confcmd.NewCommand(&RunServer{
			Global:          DefaultGlobal,
			Port:            90,
			FlagSet:         confcmd.NewFlagSet(),
			MultiLangHelper: confcmd.NewDefaultMultiLangHelper(),
		}),
		&cobra.Command{
			Use:   "version",
			Short: "print server version",
			Run: func(cmd *cobra.Command, args []string) {
				cmd.Println("v0.0.1")
			},
		},
	)
}
