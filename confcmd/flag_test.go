package confcmd_test

import (
	"reflect"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	. "github.com/xoctopus/confx/confcmd"
)

func TestNewFlagByStructField(t *testing.T) {
	some := &struct {
		Field1 int     `cmd:",required"         help.en:"field1 help" help.zh:"帮助1"`
		Field2 string  `cmd:"field"     env:"-" help.en:"field2 help" help.zh:"帮助2"`
		Field3 float64 `cmd:"-"`
		Field4 int64   `                env:"overwrite env"`
	}{
		Field2: "default value",
	}

	{
		sf := reflect.TypeOf(some).Elem().Field(0)
		fv := reflect.ValueOf(some).Elem().Field(0)

		flag := NewFlagByStructInfo("prev", sf, fv)
		NewWithT(t).Expect(flag.Name()).To(Equal("prev-field-1"))
		NewWithT(t).Expect(flag.EnvKey("some")).To(Equal("SOME__PREV_FIELD_1"))
		NewWithT(t).Expect(flag.Help(LangZH)).To(Equal("帮助1"))
		NewWithT(t).Expect(flag.Help(LangEN)).To(Equal("field1 help"))
		NewWithT(t).Expect(flag.IsRequired()).To(BeTrue())
		_, ok := flag.Value().(int)
		NewWithT(t).Expect(ok).To(BeTrue())
		_, ok = flag.ValueVarP().(*int)
		NewWithT(t).Expect(ok).To(BeTrue())
		NewWithT(t).Expect(flag.DefaultValue()).To(Equal(some.Field1))
	}

	{
		sf := reflect.TypeOf(some).Elem().Field(1)
		fv := reflect.ValueOf(some).Elem().Field(1)

		flag := NewFlagByStructInfo("", sf, fv)
		NewWithT(t).Expect(flag.Name()).To(Equal("field"))
		NewWithT(t).Expect(flag.EnvKey("some")).To(Equal(""))
		NewWithT(t).Expect(flag.Help(LangZH)).To(Equal("帮助2"))
		NewWithT(t).Expect(flag.Help(LangEN)).To(Equal("field2 help"))
		NewWithT(t).Expect(flag.IsRequired()).To(BeFalse())
		_, ok := flag.Value().(string)
		NewWithT(t).Expect(ok).To(BeTrue())
		_, ok = flag.ValueVarP().(*string)
		NewWithT(t).Expect(ok).To(BeTrue())
		NewWithT(t).Expect(flag.DefaultValue()).To(Equal(some.Field2))
	}

	{
		sf := reflect.TypeOf(some).Elem().Field(2)
		fv := reflect.ValueOf(some).Elem().Field(2)

		flag := NewFlagByStructInfo("any", sf, fv)
		NewWithT(t).Expect(flag).To(BeNil())
	}

	{
		sf := reflect.TypeOf(some).Elem().Field(3)
		fv := reflect.ValueOf(some).Elem().Field(3)

		flag := NewFlagByStructInfo("any", sf, fv)
		NewWithT(t).Expect(flag.Help(LangZH)).To(Equal(""))
		NewWithT(t).Expect(flag.Help(LangEN)).To(Equal(""))
		NewWithT(t).Expect(flag.EnvKey("")).To(Equal("ANY_OVERWRITE_ENV"))
	}
}

func TestFlag_ParseEnv(t *testing.T) {
	some := &struct {
		Bool          bool
		String        string
		Int           int
		Int8          int8
		Int16         int16
		Int32         int32
		Int64         int64
		Uint          uint
		Uint8         uint8
		Uint16        uint16
		Uint32        uint32
		Uint64        uint64
		Float32       float32
		Float64       float64
		IntSlice      []int
		Int32Slice    []int32
		Int64Slice    []int64
		UintSlice     []uint
		Float32Slice  []float32
		Float64Slice  []float64
		StringSlice   []string
		BoolSlice     []bool
		Bytes         []byte
		WrongIntSlice []int
	}{}

	cases := []*struct {
		envVal string
		expect any
		err    error
	}{
		{"true", true, nil},                                                   // 0
		{"any thing", "any thing", nil},                                       // 1
		{"100", int(100), nil},                                                // 2
		{"101", int8(101), nil},                                               // 3
		{"102", int16(102), nil},                                              // 4
		{"103", int32(103), nil},                                              // 5
		{"104", int64(104), nil},                                              // 6
		{"105", uint(105), nil},                                               // 7
		{"106", uint8(106), nil},                                              // 8
		{"107", uint16(107), nil},                                             // 9
		{"108", uint32(108), nil},                                             // 10
		{"109", uint64(109), nil},                                             // 11
		{"110", float32(110), nil},                                            // 12
		{"111", float64(111), nil},                                            // 13
		{"1 2 3", []int{1, 2, 3}, nil},                                        // 14
		{"1 2 3", []int32{1, 2, 3}, nil},                                      // 15
		{"1 2 3", []int64{1, 2, 3}, nil},                                      // 16
		{"1 2 3", []uint{1, 2, 3}, nil},                                       // 17
		{"1 2 3", []float32{1, 2, 3}, nil},                                    // 18
		{"1 2 3", []float64{1, 2, 3}, nil},                                    // 19
		{"1 2 3", []string{"1", "2", "3"}, nil},                               // 20
		{"1 1 0", []bool{true, true, false}, nil},                             // 21
		{"any", nil, errors.New("unsupported flag value type")},               // 22
		{"1 2 a", nil, errors.New("failed unmarshal from `a` to type `int`")}, // 23
	}

	rv := reflect.ValueOf(some).Elem()
	for i := 0; i < rv.NumField(); i++ {
		t.Log(i)
		sf := rv.Type().Field(i)
		f := NewFlagByStructInfo("", sf, rv.Field(i))
		err := f.ParseEnv(cases[i].envVal)
		if err == nil {
			NewWithT(t).Expect(err).To(BeNil())
			NewWithT(t).Expect(f.Value()).To(Equal(cases[i].expect))
			NewWithT(t).Expect(f.DefaultValue()).To(Equal(cases[i].expect))
		} else {
			NewWithT(t).Expect(err.Error()).To(ContainSubstring(cases[i].err.Error()))
		}
	}
}
