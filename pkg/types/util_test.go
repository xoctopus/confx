package types_test

import (
	"testing"
	"time"

	"cgtech.gitlab.com/saitox/x/testx"

	"cgtech.gitlab.com/saitox/confx/pkg/types"
)

func TestCost(t *testing.T) {
	du := time.Millisecond * 100
	span := types.Span()
	time.Sleep(du)
	testx.Expect(t, span() >= du, testx.BeTrue())
}
