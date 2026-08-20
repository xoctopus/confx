package confws_test

import (
	"context"
	"net/http"
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/confws"
)

func TestValidateHooks(t *testing.T) {
	t.Run("empty fails", func(t *testing.T) {
		var o confws.Option
		Expect(t, o.ValidateHooks(), BeFalse())
	})

	t.Run("message only ok", func(t *testing.T) {
		var o confws.Option
		o.SetMessageHandler(func(context.Context, confws.Client, int, []byte) {})
		Expect(t, o.ValidateHooks(), BeTrue())
	})

	t.Run("establish only ok", func(t *testing.T) {
		var o confws.Option
		o.SetEstablishHandler(func(ctx context.Context, _ confws.Client) (context.Context, error) {
			return ctx, nil
		})
		Expect(t, o.ValidateHooks(), BeTrue())
	})

	t.Run("connected alone fails", func(t *testing.T) {
		var o confws.Option
		o.SetConnectionHandler(func(ctx context.Context, _ *http.Request) (context.Context, []confws.ClientOptionApplier, error) {
			return ctx, nil, nil
		})
		Expect(t, o.ValidateHooks(), BeFalse())
	})
}

func TestSetDefault(t *testing.T) {
	var o confws.Option
	o.SetDefault()
	Expect(t, o.Path, Equal("/"))
	Expect(t, o.MaxConnection, Equal(65536))
	Expect(t, o.MaxMessageSize, Equal(32768))
	Expect(t, o.CheckOriginAllowAll, BeTrue())
}
