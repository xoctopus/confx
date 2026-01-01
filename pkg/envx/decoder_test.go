package envx_test

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/xoctopus/x/misc/must"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/envx"
	"github.com/xoctopus/confx/pkg/types"
)

type DefaultSetter struct {
	Int int
}

func (v *DefaultSetter) SetDefault() {
	v.Int = 101
}

type MustFailedArshaler struct {
	V any
}

func (MustFailedArshaler) MarshalText() ([]byte, error) {
	return nil, errors.New("MUST FAILED MARSHALER")
}

func (*MustFailedArshaler) UnmarshalText([]byte) error {
	return errors.New("MUST FAILED UNMARSHALER")
}

func TestDecoder_Decode(t *testing.T) {
	t.Run("Invalid", func(t *testing.T) {
		grp := envx.NewGroup("TEST")
		Expect(t, grp.Name(), Equal("TEST"))
		Expect(t, grp.Values(), HaveLen[map[string]*envx.Var](0))

		dec := envx.NewDecoder(grp)

		t.Run("InvalidValue", func(t *testing.T) {
			err := dec.Decode(nil)
			Expect(t, err, IsCodeError(envx.E_DEC__INVALID_VALUE))
		})
		t.Run("CannotSet", func(t *testing.T) {
			err := dec.Decode(struct{}{})
			Expect(t, err, IsCodeError(envx.E_DEC__INVALID_VALUE_CANNOT_SET))
		})
	})

	t.Run("DefaultSetter", func(t *testing.T) {
		grp := envx.NewGroup("TEST")
		dec := envx.NewDecoder(grp)

		// no var for decoding, set as default
		val := &DefaultSetter{}
		Expect(t, dec.Decode(val), Succeed())
		Expect(t, val.Int, Equal(101))

		// overwritten by env var
		grp.Add(envx.NewVar("Int", "200"))
		Expect(t, dec.Decode(val), Succeed())
		Expect(t, val.Int, Equal(200))
	})

	t.Run("SkippedTypes", func(t *testing.T) {
		grp := envx.NewGroup("TEST")
		dec := envx.NewDecoder(grp)

		err := dec.Decode(&struct {
			Chan chan struct{}
			Func func()
			Any  any
		}{
			Chan: make(chan struct{}),
			Func: func() {},
			Any:  1,
		})
		Expect(t, err, Succeed())
	})

	t.Run("Map", func(t *testing.T) {
		t.Run("InvalidMapKeyType", func(t *testing.T) {
			grp := envx.NewGroup("TEST")
			dec := envx.NewDecoder(grp)
			pos := envx.NewPathWalker()

			type Key struct{}

			pos.Enter("Map")
			err := dec.Decode(&struct {
				Map map[Key]any
			}{})
			Expect(t, err, IsCodeError(envx.E_DEC__INVALID_MAP_KEY_TYPE))
		})
		t.Run("FailedToUnmarshalMapKey", func(t *testing.T) {
			grp := envx.NewGroup("TEST")
			dec := envx.NewDecoder(grp)
			pos := envx.NewPathWalker()

			grp.Add(envx.NewVar("Map_INVALIDINT", "any"))
			grp.Add(envx.NewVar("OTHER_VAR_KEY", "any"))
			pos.Enter("Map")

			err := dec.Decode(&struct {
				Map map[int]string
			}{})
			Expect(t, err, IsCodeError(envx.E_DEC__FAILED_UNMARSHAL))
		})
		t.Run("FailedToDecodeMapValue", func(t *testing.T) {
			grp := envx.NewGroup("TEST")
			dec := envx.NewDecoder(grp)
			pos := envx.NewPathWalker()

			pos.Enter("Map")
			pos.Enter("KEY1")
			grp.Add(envx.NewVar("Map_KEY1", "invalid integer string"))

			err := dec.Decode(&struct {
				Map map[string]int
			}{})
			Expect(t, err, IsCodeError(envx.E_DEC__FAILED_UNMARSHAL))
		})
		t.Run("Success", func(t *testing.T) {
			t.Run("RawMap", func(t *testing.T) {
				grp := envx.NewGroup("TEST")
				dec := envx.NewDecoder(grp)

				grp.Add(envx.NewVar("1", "1"))
				grp.Add(envx.NewVar("2", "2"))

				target := map[int]string{}
				expect := map[int]string{1: "1", 2: "2"}

				err := dec.Decode(&target)
				Expect(t, err, Succeed())
				Expect(t, expect, Equal(target))
			})
			t.Run("PrefixedMap", func(t *testing.T) {
				grp := envx.NewGroup("TEST")
				dec := envx.NewDecoder(grp)

				grp.Add(envx.NewVar("Map_1", "1"))
				grp.Add(envx.NewVar("Map_2", "2"))

				target := struct{ Map map[int]string }{}
				expect := struct{ Map map[int]string }{
					Map: map[int]string{1: "1", 2: "2"},
				}

				err := dec.Decode(&target)
				Expect(t, err, Succeed())
				Expect(t, expect, Equal(target))
			})
			t.Run("PrefixedAndSuffixedMap", func(t *testing.T) {
				grp := envx.NewGroup("TEST")
				dec := envx.NewDecoder(grp)

				grp.Add(envx.NewVar("Map_1_1", "1"))
				grp.Add(envx.NewVar("Map_2_2", "2"))

				target := struct{ Map map[int]map[string]int }{}
				expect := struct{ Map map[int]map[string]int }{
					Map: map[int]map[string]int{1: {"1": 1}, 2: {"2": 2}},
				}

				err := dec.Decode(&target)
				Expect(t, err, Succeed())
				Expect(t, expect, Equal(target))
			})
		})
	})

	t.Run("ArrayAndSlice", func(t *testing.T) {
		t.Run("FailedToDecodeElem", func(t *testing.T) {
			grp := envx.NewGroup("TEST")
			dec := envx.NewDecoder(grp)
			pos := envx.NewPathWalker()

			pos.Enter("0")
			grp.Add(envx.NewVar("0", "invalid integer string"))

			err := dec.Decode(&([]int{}))
			Expect(t, err, IsCodeError(envx.E_DEC__FAILED_UNMARSHAL))
		})
		t.Run("Success", func(t *testing.T) {
			t.Run("RawSlice", func(t *testing.T) {
				grp := envx.NewGroup("TEST")
				dec := envx.NewDecoder(grp)

				grp.Add(envx.NewVar("2", "2"))
				grp.Add(envx.NewVar("4", "4"))
				grp.Add(envx.NewVar("INVALID_INDEX", "any")) // skipped

				target := make([]int, 0)
				expect := []int{0, 0, 2, 0, 4}
				Expect(t, dec.Decode(&target), Succeed())
				Expect(t, expect, Equal(target))
			})
			t.Run("PrefixedArray", func(t *testing.T) {
				grp := envx.NewGroup("TEST")
				dec := envx.NewDecoder(grp)

				grp.Add(envx.NewVar("Array_2", "2"))
				grp.Add(envx.NewVar("Array_4", "4"))               // skipped: out of capacity
				grp.Add(envx.NewVar("Array_INVALID_INDEX", "any")) // skipped

				target := struct{ Array [3]int }{}
				expect := struct{ Array [3]int }{Array: [3]int{0, 0, 2}}
				Expect(t, dec.Decode(&target), Succeed())
				Expect(t, expect, Equal(target))
			})
			t.Run("PrefixedAndSuffixedArray", func(t *testing.T) {
				grp := envx.NewGroup("TEST")
				dec := envx.NewDecoder(grp)

				grp.Add(envx.NewVar("Array_1", "http://localhost:9999"))
				grp.Add(envx.NewVar("Array_4", "4"))               // skipped: out of capacity
				grp.Add(envx.NewVar("Array_INVALID_INDEX", "any")) // skipped

				target := struct{ Array [3]*types.Endpoint }{}
				expect := struct{ Array [3]*types.Endpoint }{
					Array: [3]*types.Endpoint{
						nil,
						{Scheme: "http", Host: "localhost", Port: 9999, Param: url.Values{}},
						nil,
					},
				}
				Expect(t, dec.Decode(&target), Succeed())
				Expect(t, expect, Equal(target))
			})
		})
	})

	t.Run("Struct", func(t *testing.T) {
		t.Run("FailedToDecodeField", func(t *testing.T) {
			grp := envx.NewGroup("TEST")
			dec := envx.NewDecoder(grp)
			pos := envx.NewPathWalker()

			pos.Enter("MustFailed")
			grp.Add(envx.NewVar("MustFailed", "any"))

			err := dec.Decode(&struct {
				Endpoint   *types.Endpoint
				MustFailed MustFailedArshaler
			}{
				MustFailed: MustFailedArshaler{},
			})
			Expect(t, err, IsCodeError(envx.E_DEC__FAILED_UNMARSHAL))
		})
	})
}

type TestData struct {
	// expect skipped because field is unexported
	unexported any
	// expect skipped because field is a nil pointer
	NilPointer *int
	// expect is optional and name overwritten to `address`
	HasTag *types.Endpoint `env:"address,optional"`
	// expect skipped
	SkipTag types.Endpoint `env:"-"`
	// expect marshaled as a single string and password field will be masked
	Endpoint *types.Endpoint
	// expect masked
	types.Password
	RedisInstances map[string]*types.Endpoint `env:"redis"`
	MysqlInstances [2]*types.Endpoint         `env:"database"`
}

var x = &TestData{
	unexported: "any",
	HasTag:     &types.Endpoint{Scheme: "http", Host: "localhost"},
	SkipTag:    types.Endpoint{Scheme: "http", Host: "localhost"},
	Endpoint: &types.Endpoint{
		Scheme:   "http",
		Host:     "localhost",
		Port:     8888,
		Base:     "root",
		Username: "username",
		Password: "password",
		Param:    url.Values{"key": []string{"value1", "value2"}},
	},
	Password: "password",
	RedisInstances: map[string]*types.Endpoint{
		"instance1": must.NoErrorV(types.ParseEndpoint("redis://u1:p1@host1:3306/1")),
		"instance2": must.NoErrorV(types.ParseEndpoint("redis://u2:p2@host2:3306/2")),
	},
	MysqlInstances: [2]*types.Endpoint{
		must.NoErrorV(types.ParseEndpoint("mysql://u:p@host1:3306/db1?ssl=off")),
		must.NoErrorV(types.ParseEndpoint("mysql://u:p@host2:3306/db2?ssl=off")),
	},
}

func Example_env() {
	envs := [][2]string{
		{"EXAMPLE__ADDRESS", "asset://group/filename.png"},
		{"EXAMPLE__DATABASE_0", "mysql://u:p@host1:3306/db1?ssl=off"},
		{"EXAMPLE__DATABASE_1", "mysql://u:p@host2:3306/db2?ssl=off"},
		{"EXAMPLE__Endpoint", "http://username:password@localhost:8888/root?key=value1&key=value2"},
		{"EXAMPLE__Password", "password"},
		{"EXAMPLE__REDIS_instance1", "redis://u1:p1@host1:3306/1"},
		{"EXAMPLE__REDIS_instance2", "redis://u2:p2@host2:3306/2"},
	}
	for _, v := range envs {
		_ = os.Setenv(v[0], v[1])
	}

	defer func() {
		for _, v := range envs {
			_ = os.Unsetenv(v[0])
		}
	}()

	target := TestData{}

	// read configurations from env vars and decode to target
	dec := envx.NewDecoder(envx.ParseGroupFromEnv("EXAMPLE"))
	must.NoError(dec.Decode(&target))

	// encode target
	grp := envx.NewGroup("EXAMPLE")
	enc := envx.NewEncoder(grp)
	must.NoError(enc.Encode(target))

	// output dotenv raw and masked
	fmt.Println(string(grp.Bytes()))
	fmt.Println(string(grp.MaskBytes()))

	// Output:
	// EXAMPLE__ADDRESS=asset://group/filename.png
	// EXAMPLE__DATABASE_0=mysql://u:p@host1:3306/db1?ssl=off
	// EXAMPLE__DATABASE_1=mysql://u:p@host2:3306/db2?ssl=off
	// EXAMPLE__Endpoint=http://username:password@localhost:8888/root?key=value1&key=value2
	// EXAMPLE__NilPointer=0
	// EXAMPLE__Password=password
	// EXAMPLE__REDIS_instance1=redis://u1:p1@host1:3306/1
	// EXAMPLE__REDIS_instance2=redis://u2:p2@host2:3306/2
	//
	// EXAMPLE__ADDRESS=asset://group/filename.png
	// EXAMPLE__DATABASE_0=mysql://u:--------@host1:3306/db1?ssl=off
	// EXAMPLE__DATABASE_1=mysql://u:--------@host2:3306/db2?ssl=off
	// EXAMPLE__Endpoint=http://username:--------@localhost:8888/root?key=value1&key=value2
	// EXAMPLE__NilPointer=0
	// EXAMPLE__Password=--------
	// EXAMPLE__REDIS_instance1=redis://u1:--------@host1:3306/1
	// EXAMPLE__REDIS_instance2=redis://u2:--------@host2:3306/2
}
