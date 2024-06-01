package confcmd_test

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/ptrx"
	"github.com/xoctopus/x/reflectx"

	. "github.com/xoctopus/confx/confcmd"
)

func TestNewFlagValue(t *testing.T) {
	data := &struct {
		IntPtr                   *int       // 0
		Int                      int        // 1
		Int8                     int8       // 2
		Int16                    int16      // 3
		Int32                    int32      // 4
		Int64                    int64      // 5
		UintPtr                  *uint      // 6
		Uint                     uint       // 7
		Uint8                    uint8      // 8
		Uint16                   uint16     // 9
		Uint32                   uint32     // 10
		Uint64                   uint64     // 11
		Float32                  float32    // 12
		Float64                  float64    // 13
		String                   string     // 14
		Boolean                  bool       // 15
		IntegerSlice             []int      // 16
		IntegerPtrSlice          []*int     // 17
		UnsignedIntegerSlice     []uint     // 18
		UnsignedIntegerPtrSlice  []*uint    // 19
		Float64Slice             []float64  // 20
		Float64PtrSlice          []*float64 // 21
		StringSlice              []string   // 22
		StringPtrSlice           []*string  // 23
		BooleanSlice             []bool     // 24
		BooleanPtrSlice          []*bool    // 25
		CanBeTextEncoded         big.Int    // 26
		CanBeTextEncodedSlice    []big.Int  // 27
		CanBeTextEncodedPtr      *big.Int   // 28
		CanBeTextEncodedPtrSlice []*big.Int // 29
		MultiDimensionalSlice    [][]int    // 30
		MultiDimensionalSlice2   []*[]int   // 31
	}{
		UintPtr:          ptrx.Ptr(uint(100)),
		String:           "abc",
		CanBeTextEncoded: *must.BeTrueV(new(big.Int).SetString("100", 10)),
		CanBeTextEncodedSlice: []big.Int{
			*must.BeTrueV(new(big.Int).SetString("100", 10)),
			*must.BeTrueV(new(big.Int).SetString("101", 10)),
			*must.BeTrueV(new(big.Int).SetString("102", 10)),
		},
	}

	cases := []*struct {
		defaultValue string   // assert equal flagValue.String() before set
		typeName     string   // assert flagValue.Type()
		setValues    []string // call flagValue.Set(string)
		setValuesErr string   // assert call flagValue.Set(string) result
		value        string   // assert equal flagValue.String() after set
		value2       any      // assert equal flagValue's value after set
	}{
		// IntPtr *int
		{"0", "int", []string{"1"}, "", "1", ptrx.Ptr(int(1))},
		// Int int
		{"0", "int", []string{"0xFF"}, "", "255", int(255)},
		// Int8  int8
		{"0", "int8", []string{"077"}, "", "63", int8(63)},
		// Int16 int16
		{"0", "int16", []string{"0b1111"}, "", "15", int16(15)},
		// Int32 int32
		{"0", "int32", []string{"defg"}, "failed to parse int32", "", int32(0)},
		// Int64 int64
		{"0", "int64", []string{"1"}, "", "1", int64(1)},
		// UintPtr *uint
		{"100", "uint", []string{"101"}, "", "101", ptrx.Ptr(uint(101))},
		// Uint uint
		{"0", "uint", []string{"1"}, "", "1", uint(1)},
		// Uint8 uint8
		{"0", "uint8", []string{"1"}, "", "1", uint8(1)},
		// Uint16 uint16
		{"0", "uint16", []string{"1"}, "", "1", uint16(1)},
		// Uint32 uint32
		{"0", "uint32", []string{"1"}, "", "1", uint32(1)},
		// Uint64 uint64
		{"0", "uint64", []string{"1"}, "", "1", uint64(1)},
		// Float32 float32
		{"0", "float32", []string{"100.1"}, "", "100.1", float32(100.1)},
		// Float64 float64
		{"0", "float64", []string{"100.2"}, "", "100.2", float64(100.2)},
		// String string
		{"abc", "string", []string{"def"}, "", "def", "def"},
		// Boolean bool
		{"false", "bool", []string{"1", "0", "true"}, "", "true", true},
		// IntegerSlice []int
		{"[]", "[]int", []string{"1", "2\v3", "4 5", "6\t7"}, "", "[1 2 3 4 5 6 7]", []int{1, 2, 3, 4, 5, 6, 7}},
		// IntegerPtrSlice []*int
		{"[]", "[]*int", []string{"8\r9"}, "", "[8 9]", []*int{ptrx.Ptr(8), ptrx.Ptr(9)}},
		// UnsignedIntegerSlice []uint
		{"[]", "[]uint", []string{"99\n98"}, "", "[99 98]", []uint{99, 98}},
		// UnsignedIntegerPtrSlice []*uint
		// input string "101,102" is scanned as just one field
		// textx.UnmarshalText scans string to numeric use fmt.Sscan to support multi numeric bases
		// so should split field use strings.asciiSpace. it supports '\t', '\n', '\v' , '\f', '\r' and ' ' to split a string to fields
		{"[]", "[]*uint", []string{"101,102", "103\f104"}, "", "[101 103 104]", []*uint{ptrx.Ptr(uint(101)), ptrx.Ptr(uint(103)), ptrx.Ptr(uint(104))}},
		// Float64Slice []float64
		{"[]", "[]float64", []string{}, "", "[]", []float64(nil)},
		// Float64PtrSlice []*float64
		{"[]", "[]*float64", []string{"1.00001"}, "", "[1.00001]", []*float64{ptrx.Ptr(float64(1.00001))}},
		// StringSlice []string
		{"[]", "[]string", []string{"a", "b", "c d", "e\nf"}, "", "[a b c d e f]", []string{"a", "b", "c", "d", "e", "f"}},
		// StringPtrSlice []*string
		{"[]", "[]*string", []string{"ab cd ef"}, "", "[ab cd ef]", []*string{ptrx.Ptr("ab"), ptrx.Ptr("cd"), ptrx.Ptr("ef")}},
		// BooleanSlice []bool
		{"[]", "[]bool", []string{"1 true false 0"}, "", "[true true false false]", []bool{true, true, false, false}},
		// BooleanPtrSlice []*bool
		{"[]", "[]*bool", []string{"1 0 -1"}, "failed to parse []*bool at index 2", "true false", []*bool{ptrx.Ptr(true), ptrx.Ptr(false)}},
		// CanBeTextEncoded big.Int
		{"100", "big.Int", []string{"", "101"}, "", "101", *(new(big.Int).SetInt64(101))},
		// CanBeTextEncodedSlice []big.Int
		{
			defaultValue: "[100 101 102]",
			typeName:     "[]big.Int",
			setValues:    []string{"", "103", "104"},
			setValuesErr: "",
			value:        "[100 101 102 103 104]",
			value2: []big.Int{
				*(new(big.Int).SetInt64(100)),
				*(new(big.Int).SetInt64(101)),
				*(new(big.Int).SetInt64(102)),
				*(new(big.Int).SetInt64(103)),
				*(new(big.Int).SetInt64(104)),
			},
		},
		// CanBeTextEncodedPtr *big.Int
		{
			defaultValue: "0",
			typeName:     "big.Int",
			setValues:    []string{"999"},
			setValuesErr: "",
			value:        "999",
			value2:       new(big.Int).SetInt64(999),
		},
		// CanBeTextEncodedPtrSlice []*big.Int // 29
		{
			defaultValue: "[]",
			typeName:     "[]*big.Int",
			setValues:    []string{"998 997 996"},
			setValuesErr: "",
			value:        "[998 997 996]",
			value2: []*big.Int{
				new(big.Int).SetInt64(998),
				new(big.Int).SetInt64(997),
				new(big.Int).SetInt64(996),
			},
		},
		// MultiDimensionalSlice [][]int
		{"[]", "[][]int", []string{"1 2 3"}, "unsupported type", "", nil},
		// MultiDimensionalSlice2 []*[]int
		{"[]", "[]*[]int", []string{"1 2 3"}, "unsupported type", "", nil},
	}

	datarv := reflect.ValueOf(data)
	for i := 0; i < len(cases); i++ {
		c := cases[i]
		rv := reflectx.IndirectNew(datarv.Elem().Field(i))
		sf := datarv.Elem().Type().Field(i)
		name := fmt.Sprintf("Index%02d_%s", i, sf.Name)

		t.Run(name, func(t *testing.T) {
			fv := NewFlagValue(rv)
			NewWithT(t).Expect(fv.String()).To(Equal(c.defaultValue))
			NewWithT(t).Expect(fv.Type()).To(Equal(c.typeName))

			var err error
			for _, val := range c.setValues {
				if err = fv.Set(val); err != nil {
					break
				}
			}

			if err != nil {
				NewWithT(t).Expect(c.setValuesErr).NotTo(BeEmpty())
				NewWithT(t).Expect(err.Error()).To(ContainSubstring(c.setValuesErr))
			} else {
				NewWithT(t).Expect(c.setValuesErr).To(Equal(""))
				NewWithT(t).Expect(fv.String()).To(Equal(c.value))
				NewWithT(t).Expect(datarv.Elem().Field(i).Interface()).To(Equal(c.value2))
			}
		})
	}
}
