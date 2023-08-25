package envi_test

import (
	"errors"
	"os"
	"strconv"
	"testing"

	"github.com/bounoable/envi"
	"github.com/google/go-cmp/cmp"
)

// TestParse is a testing function that validates the operation of the Parse
// function in the envi package. It tests parsing a variety of environment
// variable types including basic types such as string, int, bool, as well as
// complex types like arrays, slices, maps and structs. It also checks for
// expected error conditions. The function uses table-driven testing where each
// table entry defines a name for the test case, the set of environment
// variables to be parsed, the expected result after parsing and an optional
// expected error.
func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		environment map[string]string
		want        env
		wantError   error
	}{
		{
			name:        "string",
			environment: map[string]string{"MY_STRING": "foo"},
			want:        env{String: "foo"},
		},
		{
			name:        "int",
			environment: map[string]string{"MY_INT": "3"},
			want:        env{Int: 3},
		},
		{
			name:        "int (negative)",
			environment: map[string]string{"MY_INT": "-3"},
			want:        env{Int: -3},
		},
		{
			name:        "int8",
			environment: map[string]string{"MY_INT8": "5"},
			want:        env{Int8: 5},
		},
		{
			name:        "int16",
			environment: map[string]string{"MY_INT16": "12345"},
			want:        env{Int16: 12345},
		},
		{
			name:        "int32",
			environment: map[string]string{"MY_INT32": "12345"},
			want:        env{Int32: 12345},
		},
		{
			name:        "int64",
			environment: map[string]string{"MY_INT64": "8888888888"},
			want:        env{Int64: 8888888888},
		},
		{
			name:        "uint",
			environment: map[string]string{"MY_UINT": "123456789"},
			want:        env{UInt: 123456789},
		},
		{
			name:        "uint8",
			environment: map[string]string{"MY_UINT8": "8"},
			want:        env{UInt8: 8},
		},
		{
			name:        "uint16",
			environment: map[string]string{"MY_UINT16": "12345"},
			want:        env{UInt16: 12345},
		},
		{
			name:        "uint32",
			environment: map[string]string{"MY_UINT32": "12345"},
			want:        env{UInt32: 12345},
		},
		{
			name:        "uint64",
			environment: map[string]string{"MY_UINT64": "8888888888"},
			want:        env{UInt64: 8888888888},
		},
		{
			name:        "uint (negative)",
			environment: map[string]string{"MY_UINT": "-1000"},
			wantError:   strconv.ErrSyntax,
		},
		{
			name:        "complex64",
			environment: map[string]string{"MY_COMPLEX64": "3+6i"},
			want:        env{Complex64: complex(3, 6)},
		},
		{
			name:        "complex128",
			environment: map[string]string{"MY_COMPLEX128": "3+6i"},
			want:        env{Complex128: complex(3, 6)},
		},
		{
			name:        "float64",
			environment: map[string]string{"MY_FLOAT64": "9.87654321"},
			want:        env{Float64: 9.87654321},
		},
		{
			name:        "float32",
			environment: map[string]string{"MY_FLOAT32": "9.87654321"},
			want:        env{Float32: 9.87654321},
		},
		{
			name:        "bool (true)",
			environment: map[string]string{"MY_BOOL": "true"},
			want:        env{Bool: true},
		},
		{
			name:        "bool (false)",
			environment: map[string]string{"MY_BOOL": "false"},
			want:        env{Bool: false},
		},
		{
			name:        "bool (0)",
			environment: map[string]string{"MY_BOOL": "0"},
			want:        env{Bool: false},
		},
		{
			name:        "bool (1)",
			environment: map[string]string{"MY_BOOL": "1"},
			want:        env{Bool: true},
		},
		{
			name:        "bool (int)",
			environment: map[string]string{"MY_BOOL": "5"},
			want:        env{Bool: true},
		},
		{
			name:        "bool (string)",
			environment: map[string]string{"MY_BOOL": "foo"},
			want:        env{Bool: true},
		},
		{
			name:        "bool (empty string)",
			environment: map[string]string{"MY_BOOL": ""},
			want:        env{Bool: false},
		},
		{
			name: "struct",
			environment: map[string]string{
				"MY_STRUCT_FOO": "foo",
				"MY_STRUCT_BAR": "123",
				"MY_STRUCT_BAZ": "true",
				"MY_NESTED_FOO": "8,4,2",
			},
			want: env{Struct: myStruct{
				Foo: "foo",
				Bar: 123,
				Baz: true,
				Nested: nestedStruct{
					Foo: [...]uint8{8, 4, 2},
				},
			}},
		},
		{
			name:        "string array",
			environment: map[string]string{"MY_STRING_ARRAY": "foo,bar,baz"},
			want:        env{StringArray: [...]string{"foo", "bar", "baz"}},
		},
		{
			name:        "bool array",
			environment: map[string]string{"MY_BOOL_ARRAY": "1,true,0,false,,foo,-1"},
			want:        env{BoolArray: [...]bool{true, true, false, false, false, true, true}},
		},
		{
			name:        "string array (overflow)",
			environment: map[string]string{"MY_STRING_ARRAY": "foo,bar,baz,foobar"},
			want:        env{StringArray: [...]string{"foo", "bar", "baz"}},
		},
		{
			name:        "string array (incomplete)",
			environment: map[string]string{"MY_STRING_ARRAY": "foo,bar"},
			want:        env{StringArray: [...]string{"foo", "bar", ""}},
		},
		{
			name:        "string slice",
			environment: map[string]string{"MY_STRING_SLICE": "foo,bar,baz,,,foobar"},
			want:        env{StringSlice: []string{"foo", "bar", "baz", "", "", "foobar"}},
		},
		{
			name:        "float64 slice",
			environment: map[string]string{"MY_FLOAT64_SLICE": "0,-1,2.4,-3.6"},
			want:        env{Float64Slice: []float64{0, -1, 2.4, -3.6}},
		},
		{
			name: "string map",
			environment: map[string]string{
				"MY_STRING_MAP_fOo": "baR",
				"MY_STRING_MAP_BaR": "Baz",
			},
			want: env{StringMap: map[string]string{
				"fOo": "baR",
				"BaR": "Baz",
			}},
		},
		{
			name: "string-int map",
			environment: map[string]string{
				"MY_INT_STRING_MAP_8":   "foo",
				"MY_INT_STRING_MAP_-18": "bar",
			},
			want: env{IntStringMap: map[int]string{
				8:   "foo",
				-18: "bar",
			}},
		},
		{
			name: "bool-int map",
			environment: map[string]string{
				"MY_BOOL_INT_MAP_true":  "1",
				"MY_BOOL_INT_MAP_false": "-2",
			},
			want: env{BoolIntMap: map[bool]int{
				true:  1,
				false: -2,
			}},
		},
		{
			name: "bool-int map (implicit from string)",
			environment: map[string]string{
				"MY_BOOL_INT_MAP_":    "1",
				"MY_BOOL_INT_MAP_abc": "-2",
			},
			want: env{BoolIntMap: map[bool]int{
				false: 1,
				true:  -2,
			}},
		},
		{
			name: "bool-int map (implicit from int)",
			environment: map[string]string{
				"MY_BOOL_INT_MAP_-8": "-1",
				"MY_BOOL_INT_MAP_0":  "2",
			},
			want: env{BoolIntMap: map[bool]int{
				true:  -1,
				false: 2,
			}},
		},
		{
			name: "bool-int map (implicit from int)",
			environment: map[string]string{
				"MY_BOOL_INT_MAP_-8": "-1",
				"MY_BOOL_INT_MAP_0":  "2",
			},
			want: env{BoolIntMap: map[bool]int{
				true:  -1,
				false: 2,
			}},
		},
		{
			name: "complex64-uint16 map",
			environment: map[string]string{
				"MY_COMPLEX64_UINT16_MAP_4+7i":    "500",
				"MY_COMPLEX64_UINT16_MAP_-3+100i": "1500",
				"MY_COMPLEX64_UINT16_MAP_200":     "2000",
			},
			want: env{Complex64UInt16Map: map[complex64]uint16{
				complex(4, 7):    500,
				complex(-3, 100): 1500,
				complex(200, 0):  2000,
			}},
		},
		{
			name: "float64-complex128 map",
			environment: map[string]string{
				"MY_FLOAT64_COMPLEX128_MAP_4.5": "5+7i",
				"MY_FLOAT64_COMPLEX128_MAP_-3":  "-3+100i",
				"MY_FLOAT64_COMPLEX128_MAP_200": "200",
			},
			want: env{Float64Complex128Map: map[float64]complex128{
				4.5: complex(5, 7),
				-3:  complex(-3, 100),
				200: complex(200, 0),
			}},
		},
		{
			name: "pointers",
			environment: map[string]string{
				"MY_STRING_PTR":       "foo",
				"MY_INT_PTR":          "42",
				"MY_BOOL_PTR":         "true",
				"MY_STRING_ARRAY_PTR": "foo,bar,baz",
				"MY_INT_ARRAY_PTR":    "1,2,3",
				"MY_BOOL_ARRAY_PTR":   "true,false,true",
				"MY_STRING_SLICE_PTR": "foo,bar,baz",
				"MY_INT_SLICE_PTR":    "1,2,3",
			},
			want: env{
				StringPtr:      ptr("foo"),
				IntPtr:         ptr(42),
				BoolPtr:        ptr(true),
				StringArrayPtr: &[3]string{"foo", "bar", "baz"},
				IntArrayPtr:    &[3]int{1, 2, 3},
				BoolArrayPtr:   &[3]bool{true, false, true},
				StringSlicePtr: &[]string{"foo", "bar", "baz"},
				IntSlicePtr:    &[]int{1, 2, 3},
			},
		},
		{
			name: "struct pointer",
			environment: map[string]string{
				"MY_PTR_STRUCT_FOO": "foo",
				"MY_PTR_STRUCT_BAR": "42",
				"MY_PTR_STRUCT_BAZ": "true",
				"MY_PTR_NESTED_FOO": "1,2,3",
			},
			want: env{
				StructPtr: &myPtrStruct{
					Foo:    "foo",
					Bar:    42,
					Baz:    true,
					Nested: &nestedPtrStruct{Foo: [...]uint8{1, 2, 3}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tt.environment {
				if err := os.Setenv(k, v); err != nil {
					t.Fatalf("os.Setenv(%q, %q): %v", k, v, err)
				}
			}

			var e env

			err := envi.Parse(&e)
			if tt.wantError != nil {
				if !errors.Is(err, tt.wantError) {
					t.Fatalf("Parse() should fail with %q; got %q", tt.wantError, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Parse() failed: %v", err)
			}

			if !cmp.Equal(tt.want, e) {
				t.Fatalf("env = %v, want = %v\n\n%s", e, tt.want, cmp.Diff(tt.want, e))
			}
		})
	}
}

type env struct {
	Struct               myStruct
	StructPtr            *myPtrStruct
	String               string                 `env:"MY_STRING"`
	Int                  int                    `env:"MY_INT"`
	Int8                 int                    `env:"MY_INT8"`
	Int16                int                    `env:"MY_INT16"`
	Int32                int                    `env:"MY_INT32"`
	Int64                int                    `env:"MY_INT64"`
	UInt                 uint                   `env:"MY_UINT"`
	UInt8                uint8                  `env:"MY_UINT8"`
	UInt16               uint16                 `env:"MY_UINT16"`
	UInt32               uint32                 `env:"MY_UINT32"`
	UInt64               uint64                 `env:"MY_UINT64"`
	Complex64            complex64              `env:"MY_COMPLEX64"`
	Complex128           complex128             `env:"MY_COMPLEX128"`
	Float64              float64                `env:"MY_FLOAT64"`
	Float32              float32                `env:"MY_FLOAT32"`
	Bool                 bool                   `env:"MY_BOOL"`
	StringArray          [3]string              `env:"MY_STRING_ARRAY"`
	BoolArray            [7]bool                `env:"MY_BOOL_ARRAY"`
	StringSlice          []string               `env:"MY_STRING_SLICE"`
	BoolSlice            []bool                 `env:"MY_BOOL_SLICE"`
	Float64Slice         []float64              `env:"MY_FLOAT64_SLICE"`
	StringMap            map[string]string      `env:"MY_STRING_MAP"`
	IntStringMap         map[int]string         `env:"MY_INT_STRING_MAP"`
	BoolIntMap           map[bool]int           `env:"MY_BOOL_INT_MAP"`
	Complex64UInt16Map   map[complex64]uint16   `env:"MY_COMPLEX64_UINT16_MAP"`
	Float64Complex128Map map[float64]complex128 `env:"MY_FLOAT64_COMPLEX128_MAP"`
	StringPtr            *string                `env:"MY_STRING_PTR"`
	IntPtr               *int                   `env:"MY_INT_PTR"`
	BoolPtr              *bool                  `env:"MY_BOOL_PTR"`
	StringArrayPtr       *[3]string             `env:"MY_STRING_ARRAY_PTR"`
	IntArrayPtr          *[3]int                `env:"MY_INT_ARRAY_PTR"`
	BoolArrayPtr         *[3]bool               `env:"MY_BOOL_ARRAY_PTR"`
	StringSlicePtr       *[]string              `env:"MY_STRING_SLICE_PTR"`
	IntSlicePtr          *[]int                 `env:"MY_INT_SLICE_PTR"`
}

type myStruct struct {
	Foo    string `env:"MY_STRUCT_FOO"`
	Bar    int    `env:"MY_STRUCT_BAR"`
	Baz    bool   `env:"MY_STRUCT_BAZ"`
	Nested nestedStruct
}

type nestedStruct struct {
	Foo [3]uint8 `env:"MY_NESTED_FOO"`
}

type myPtrStruct struct {
	Foo    string `env:"MY_PTR_STRUCT_FOO"`
	Bar    int    `env:"MY_PTR_STRUCT_BAR"`
	Baz    bool   `env:"MY_PTR_STRUCT_BAZ"`
	Nested *nestedPtrStruct
}

type nestedPtrStruct struct {
	Foo [3]uint8 `env:"MY_PTR_NESTED_FOO"`
}

func ptr[V any](v V) *V {
	return &v
}
