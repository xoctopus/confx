package helper

import (
	"os"
	"path/filepath"

	"cgtech.gitlab.com/saitox/x/stringsx"
)

func HostIdentifier(app string) string {
	hostname, _ := os.Hostname()
	if len(hostname) == 0 {
		hostname = stringsx.Random(8)
	}

	if len(app) == 0 {
		if len(os.Args) > 0 {
			app = filepath.Base(os.Args[0])
		} else {
			app = stringsx.Random(16)
		}
	}
	return app + "-" + hostname
}

func DefaultHostIdentifier() string {
	return HostIdentifier("")
}
