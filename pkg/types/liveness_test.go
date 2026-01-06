package types_test

import (
	"context"
	"encoding/json"
	"log"

	"github.com/xoctopus/x/misc/must"

	"github.com/xoctopus/confx/pkg/types"
)

func ExampleCheckLiveness() {
	endpoints := []types.SchemedEndpoint{
		&types.EndpointNoOption{Address: "redis://localhost:6379/1"},
		&types.EndpointNoOption{Address: "https://www.google.com:443"},
		&types.EndpointNoOption{Address: "http://www.google.com"},
	}

	for _, ep := range endpoints {
		if x, ok := ep.(interface{ Init() error }); ok {
			if err := x.Init(); err != nil {
				return
			}
		}
	}

	m := types.CheckLiveness(context.Background(), endpoints...)
	log.Println("\n" + string(must.NoErrorV(json.MarshalIndent(m, "", "    "))))

	//Output:
}
