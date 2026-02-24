package liveness_test

import (
	"context"
	"encoding/json"
	"log"

	"github.com/xoctopus/x/misc/must"

	"github.com/xoctopus/confx/pkg/types"
	"github.com/xoctopus/confx/pkg/types/liveness"
)

func ExampleCheckLiveness() {
	endpoints := []liveness.HasSchemeEndpoint{
		&types.EndpointNoOption{Address: "redis://example:6379/1"},
		&types.EndpointNoOption{Address: "https://www.google.com:443"},
		&types.EndpointNoOption{Address: "mysql://example:3306/mysql"},
	}

	for _, ep := range endpoints {
		if x, ok := ep.(interface{ Init() error }); ok {
			if err := x.Init(); err != nil {
				return
			}
		}
	}

	//	{
	//		"https": {
	//			"https://www.google.com:443": {
	//				"reachable": true,
	//				"rtt(ms)": 0,
	//				"msg": "success"
	//			}
	//		},
	//		"mysql": {
	//			"mysql://example:3306/mysql": {
	//				"reachable": false,
	//				"rtt(ms)": 0,
	//				"msg": "dial tcp: lookup example: no such host"
	//			}
	//		},
	//		"redis": {
	//			"redis://example:6379/1": {
	//				"reachable": false,
	//				"rtt(ms)": 0,
	//				"msg": "dial tcp: lookup example: no such host"
	//			}
	//		}
	//	}

	m := liveness.CheckLiveness(context.Background(), endpoints...)
	log.Println("\n" + string(must.NoErrorV(json.MarshalIndent(m, "", "\t"))))

	// Output:
	//
}
