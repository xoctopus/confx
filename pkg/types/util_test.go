package types_test

import (
	"testing"
	"time"

	"github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/types"
)

func TestCost(t *testing.T) {
	du := time.Millisecond * 100
	span := types.Cost()
	time.Sleep(du)
	testx.Expect(t, span() >= du, testx.BeTrue())
}
