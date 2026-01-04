package cmds

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/xoctopus/confx/pkg/cmdx"
)

// Server start an echo server
type Server struct {
	// Debug enable debug mode 1-enable
	Debug bool `cmd:",default=1"`
	// LogLevel set log level [debug info warn error]
	LogLevel string `cmd:",default=debug"`
	// Port server listen port
	Port uint16 `cmd:",default=80"`
}

var _ cmdx.Executor = (*Server)(nil)

func (s *Server) Exec(cmd *cobra.Command, args ...string) error {
	http.HandleFunc("/echo", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		defer req.Body.Close() //nolint:errcheck
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
	})

	cmd.Println("Debug:   ", s.Debug)
	cmd.Println("LogLevel:", s.LogLevel)
	cmd.Println("Port:    ", s.Port)

	if !s.Debug {
		log.SetOutput(io.Discard)
	}

	addr := fmt.Sprintf("0.0.0.0:%d", s.Port)
	cmd.Printf("server started, listening %s\n", addr)

	return http.ListenAndServe(addr, nil)
}

var CmdServer = cmdx.NewCommand("server", &Server{}).Cmd()
