package cmds

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/xoctopus/confx/confcmd"
	"github.com/xoctopus/confx/confmws/conftls"
)

var DefaultServerConfig = &ServerConfig{
	Debug:    false,
	Security: false,
	Port:     8888,
}

type ServerConfig struct {
	Debug    bool   `cmd:",p,nop=1"    help:"debug mode"`
	LogLevel string `cmd:",p"          help:"set log level [trace debug info warn error]"`
	Security bool   `cmd:",p,nop=true" help:"enable https serve and request"`
	Port     uint16 `cmd:",p"          help:"server listen port"`
	Tls      conftls.X509KeyPair
}

var _ confcmd.Executor = (*ServerConfig)(nil)

func (s *ServerConfig) Use() string { return "serve" }

func (s *ServerConfig) Short() string { return "start http server" }

func (s *ServerConfig) Exec(cmd *cobra.Command, args ...string) error {
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
	})

	cmd.Println("Debug:   ", s.Debug)
	cmd.Println("LogLevel:", s.LogLevel)
	cmd.Println("Security:", s.Security)
	cmd.Println("Port:    ", s.Port)

	if !s.Debug {
		log.SetOutput(io.Discard)
	}

	addr := fmt.Sprintf(":%d", s.Port)
	cmd.Printf("server started, listening %s\n", addr)
	if s.Security && !s.Tls.IsZero() {
		if err := s.Tls.Init(); err != nil {
			return err
		}
		return http.ListenAndServeTLS(addr, s.Tls.Crt, s.Tls.Key, nil)
	}
	return http.ListenAndServe(addr, nil)
}

var ServerCmd *cobra.Command

func init() {
	ServerCmd = confcmd.NewCommand(DefaultServerConfig, nil)
}
