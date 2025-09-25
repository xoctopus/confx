package envconf_test

import (
	"fmt"
	"net/url"
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/xoctopus/datatypex"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/textx/testdata"

	"github.com/xoctopus/confx/envconf"
)

type DefaultSetter struct {
	Int int
}

func (v *DefaultSetter) SetDefault() {
	v.Int = 101
}

func TestDecoder_Decode(t *testing.T) {

	t.Run("Invalid", func(t *testing.T) {
		grp := envconf.NewGroup("TEST")
		dec := envconf.NewDecoder(grp)

		t.Run("InvalidValue", func(t *testing.T) {
			expect := envconf.NewError(nil, envconf.E_DEC__INVALID_VALUE)
			target := dec.Decode(nil)
			NewWithT(t).Expect(errors.Is(expect, target)).To(BeTrue())
		})
		t.Run("CannotSet", func(t *testing.T) {
			expect := envconf.NewError(nil, envconf.E_DEC__INVALID_VALUE_CANNOT_SET)
			target := dec.Decode(struct{}{})
			NewWithT(t).Expect(errors.Is(expect, target)).To(BeTrue())
		})
	})

	t.Run("DefaultSetter", func(t *testing.T) {
		grp := envconf.NewGroup("TEST")
		dec := envconf.NewDecoder(grp)

		// no var for decoding, set as default
		val := &DefaultSetter{}
		NewWithT(t).Expect(dec.Decode(val)).To(BeNil())
		NewWithT(t).Expect(val.Int).To(Equal(101))

		// overwritten by env var
		grp.Add(envconf.NewVar("Int", "200"))
		NewWithT(t).Expect(dec.Decode(val)).To(BeNil())
		NewWithT(t).Expect(val.Int).To(Equal(200))
	})

	t.Run("SkippedTypes", func(t *testing.T) {
		grp := envconf.NewGroup("TEST")
		dec := envconf.NewDecoder(grp)

		err := dec.Decode(&struct {
			Chan chan struct{}
			Func func()
			Any  any
		}{
			Chan: make(chan struct{}),
			Func: func() {},
			Any:  1,
		})
		NewWithT(t).Expect(err).To(BeNil())
	})

	t.Run("Map", func(t *testing.T) {
		t.Run("InvalidMapKeyType", func(t *testing.T) {
			grp := envconf.NewGroup("TEST")
			dec := envconf.NewDecoder(grp)
			pos := envconf.NewPathWalker()

			type Key struct{}

			pos.Enter("Map")
			expect := envconf.NewError(pos, envconf.E_DEC__INVALID_MAP_KEY_TYPE)
			target := dec.Decode(&struct {
				Map map[Key]any
			}{})
			NewWithT(t).Expect(errors.Is(expect, target)).To(BeTrue())
		})
		t.Run("FailedToUnmarshalMapKey", func(t *testing.T) {
			grp := envconf.NewGroup("TEST")
			dec := envconf.NewDecoder(grp)
			pos := envconf.NewPathWalker()

			grp.Add(envconf.NewVar("Map_INVALIDINT", "any"))
			grp.Add(envconf.NewVar("OTHER_VAR_KEY", "any"))
			pos.Enter("Map")

			expect := envconf.NewError(pos, envconf.E_DEC__FAILED_UNMARSHAL)
			target := dec.Decode(&struct {
				Map map[int]string
			}{})
			NewWithT(t).Expect(errors.Is(expect, target)).To(BeTrue())
		})
		t.Run("FailedToDecodeMapValue", func(t *testing.T) {
			grp := envconf.NewGroup("TEST")
			dec := envconf.NewDecoder(grp)
			pos := envconf.NewPathWalker()

			pos.Enter("Map")
			pos.Enter("KEY1")
			grp.Add(envconf.NewVar("Map_KEY1", "invalid integer string"))

			expect := envconf.NewError(pos, envconf.E_DEC__FAILED_UNMARSHAL)
			target := dec.Decode(&struct {
				Map map[string]int
			}{})

			NewWithT(t).Expect(errors.Is(expect, target)).To(BeTrue())
		})
		t.Run("Success", func(t *testing.T) {
			t.Run("RawMap", func(t *testing.T) {
				grp := envconf.NewGroup("TEST")
				dec := envconf.NewDecoder(grp)

				grp.Add(envconf.NewVar("1", "1"))
				grp.Add(envconf.NewVar("2", "2"))

				target := map[int]string{}
				expect := map[int]string{1: "1", 2: "2"}

				err := dec.Decode(&target)
				NewWithT(t).Expect(err).To(BeNil())
				NewWithT(t).Expect(expect).To(Equal(target))
			})
			t.Run("PrefixedMap", func(t *testing.T) {
				grp := envconf.NewGroup("TEST")
				dec := envconf.NewDecoder(grp)

				grp.Add(envconf.NewVar("Map_1", "1"))
				grp.Add(envconf.NewVar("Map_2", "2"))

				target := struct{ Map map[int]string }{}
				expect := struct{ Map map[int]string }{
					Map: map[int]string{1: "1", 2: "2"},
				}

				err := dec.Decode(&target)
				NewWithT(t).Expect(err).To(BeNil())
				NewWithT(t).Expect(expect).To(Equal(target))
			})
			t.Run("PrefixedAndSuffixedMap", func(t *testing.T) {
				grp := envconf.NewGroup("TEST")
				dec := envconf.NewDecoder(grp)

				grp.Add(envconf.NewVar("Map_1_1", "1"))
				grp.Add(envconf.NewVar("Map_2_2", "2"))

				target := struct{ Map map[int]map[string]int }{}
				expect := struct{ Map map[int]map[string]int }{
					Map: map[int]map[string]int{1: {"1": 1}, 2: {"2": 2}},
				}

				err := dec.Decode(&target)
				NewWithT(t).Expect(err).To(BeNil())
				NewWithT(t).Expect(expect).To(Equal(target))
			})
		})
	})

	t.Run("ArrayAndSlice", func(t *testing.T) {
		t.Run("FailedToDecodeElem", func(t *testing.T) {
			grp := envconf.NewGroup("TEST")
			dec := envconf.NewDecoder(grp)
			pos := envconf.NewPathWalker()

			pos.Enter("0")
			grp.Add(envconf.NewVar("0", "invalid integer string"))

			expect := envconf.NewError(pos, envconf.E_DEC__FAILED_UNMARSHAL)
			target := dec.Decode(&([]int{}))
			NewWithT(t).Expect(errors.Is(expect, target)).To(BeTrue())
		})
		t.Run("Success", func(t *testing.T) {
			t.Run("RawSlice", func(t *testing.T) {
				grp := envconf.NewGroup("TEST")
				dec := envconf.NewDecoder(grp)

				grp.Add(envconf.NewVar("2", "2"))
				grp.Add(envconf.NewVar("4", "4"))
				grp.Add(envconf.NewVar("INVALID_INDEX", "any")) // skipped

				target := make([]int, 0)
				expect := []int{0, 0, 2, 0, 4}
				NewWithT(t).Expect(dec.Decode(&target)).To(BeNil())
				NewWithT(t).Expect(expect).To(Equal(target))
			})
			t.Run("PrefixedArray", func(t *testing.T) {
				grp := envconf.NewGroup("TEST")
				dec := envconf.NewDecoder(grp)

				grp.Add(envconf.NewVar("Array_2", "2"))
				grp.Add(envconf.NewVar("Array_4", "4"))               // skipped: out of capacity
				grp.Add(envconf.NewVar("Array_INVALID_INDEX", "any")) // skipped

				target := struct{ Array [3]int }{}
				expect := struct{ Array [3]int }{Array: [3]int{0, 0, 2}}
				NewWithT(t).Expect(dec.Decode(&target)).To(BeNil())
				NewWithT(t).Expect(expect).To(Equal(target))
			})
			t.Run("PrefixedAndSuffixedArray", func(t *testing.T) {
				grp := envconf.NewGroup("TEST")
				dec := envconf.NewDecoder(grp)

				grp.Add(envconf.NewVar("Array_1", "http://localhost:9999"))
				grp.Add(envconf.NewVar("Array_4", "4"))               // skipped: out of capacity
				grp.Add(envconf.NewVar("Array_INVALID_INDEX", "any")) // skipped

				target := struct{ Array [3]*datatypex.Endpoint }{}
				expect := struct{ Array [3]*datatypex.Endpoint }{
					Array: [3]*datatypex.Endpoint{
						nil,
						{Scheme: "http", Host: "localhost", Port: 9999, Param: url.Values{}},
						nil,
					},
				}
				NewWithT(t).Expect(dec.Decode(&target)).To(BeNil())
				NewWithT(t).Expect(expect).To(Equal(target))
			})
		})
	})

	t.Run("Struct", func(t *testing.T) {
		t.Run("FailedToDecodeField", func(t *testing.T) {
			grp := envconf.NewGroup("TEST")
			dec := envconf.NewDecoder(grp)
			pos := envconf.NewPathWalker()

			pos.Enter("MustFailed")
			grp.Add(envconf.NewVar("MustFailed", "any"))

			expect := envconf.NewError(pos, envconf.E_DEC__FAILED_UNMARSHAL)
			target := dec.Decode(&struct {
				Endpoint   *datatypex.Endpoint
				MustFailed testdata.MustFailedArshaler
			}{
				MustFailed: testdata.MustFailedArshaler{},
			})
			NewWithT(t).Expect(errors.Is(expect, target)).To(BeTrue())
		})
	})
}

type TestData struct {
	// expect skipped because field is unexported
	unexported any
	// expect skipped because field is a nil pointer
	NilPointer *int
	// expect is optional and name overwritten to `address`
	HasTag *datatypex.Address `env:"address,optional"`
	// expect skipped
	SkipTag datatypex.Endpoint `env:"-"`
	// expect marshaled as a single string and password field will be masked
	Endpoint *datatypex.Endpoint
	// expect masked
	datatypex.Password
	RedisInstances map[string]*datatypex.Endpoint `env:"redis"`
	MysqlInstances [2]*datatypex.Endpoint         `env:"database"`
}

var x = &TestData{
	unexported: "any",
	HasTag:     datatypex.NewAddress("group", "filename.png"),
	SkipTag:    datatypex.Endpoint{Scheme: "http", Host: "localhost"},
	Endpoint: &datatypex.Endpoint{
		Scheme:   "http",
		Host:     "localhost",
		Port:     8888,
		Base:     "root",
		Username: "username",
		Password: "password",
		Param:    url.Values{"key": []string{"value1", "value2"}},
	},
	Password: "password",
	RedisInstances: map[string]*datatypex.Endpoint{
		"instance1": must.NoErrorV(datatypex.ParseEndpoint("redis://u1:p1@host1:3306/1")),
		"instance2": must.NoErrorV(datatypex.ParseEndpoint("redis://u2:p2@host2:3306/2")),
	},
	MysqlInstances: [2]*datatypex.Endpoint{
		must.NoErrorV(datatypex.ParseEndpoint("mysql://u:p@host1:3306/db1?ssl=off")),
		must.NoErrorV(datatypex.ParseEndpoint("mysql://u:p@host2:3306/db2?ssl=off")),
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
	dec := envconf.NewDecoder(envconf.ParseGroupFromEnv("EXAMPLE"))
	must.NoError(dec.Decode(&target))

	// encode target
	grp := envconf.NewGroup("EXAMPLE")
	enc := envconf.NewEncoder(grp)
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
