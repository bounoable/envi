package envi

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// New creates an instance of the provided Env type by parsing environment
// variables according to the struct tags. It returns the parsed environment and
// an error if any occurred during parsing.
func New[Env any]() (Env, error) {
	var env Env
	err := Parse(&env)
	return env, err
}

// Must creates a new environment of type Env and parses the environment
// variables into it. If an error occurs during parsing, it panics.
func Must[Env any]() Env {
	env, err := New[Env]()
	if err != nil {
		panic(err)
	}
	return env
}

// MustParse parses the given environment variables into the provided env
// pointer, which must be a pointer to a struct. It panics if there is an error
// during parsing.
func MustParse[Env any](env *Env) {
	if err := Parse(env); err != nil {
		panic(err)
	}
}

// Parse populates the provided env pointer, which must be a pointer to a
// struct, with the parsed values of environment variables specified in the
// struct tags. It returns an error if the parsing fails.
func Parse[Env any](env *Env) error {
	rv := reflect.ValueOf(env)
	p, err := parseStruct(rv)
	if err != nil {
		return err
	}
	rv.Elem().Set(p)
	return nil
}

func parseStruct(envValue reflect.Value) (reflect.Value, error) {
	envType := envValue.Type()
	staticType := envType.Elem()

	if envType.Kind() != reflect.Pointer || staticType.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("env must be a pointer to a struct, got %s", envType)
	}

	if envValue.IsNil() {
		return reflect.Value{}, fmt.Errorf("env must not be nil")
	}

	ptr := reflect.New(staticType)
	val := ptr.Elem()

	for n := 0; n < val.NumField(); n++ {
		field := staticType.Field(n)
		parsed, ok, err := parseField(field)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("parse %q field: %w", field.Name, err)
		}
		if !ok {
			continue
		}

		val.Field(n).Set(parsed)
	}

	return val, nil
}

func parseField(field reflect.StructField) (reflect.Value, bool, error) {
	fieldKind := field.Type.Kind()

	isStruct, isPointer := isStruct(field.Type)

	if isStruct {
		ft := field.Type
		if isPointer {
			ft = ft.Elem()
		}

		fv := reflect.New(ft)

		rv, err := parseStruct(fv)
		if err != nil {
			return reflect.Value{}, false, err
		}

		if rv.IsZero() {
			return reflect.Value{}, false, nil
		}

		if isPointer {
			rv = rv.Addr()
		}

		return rv, true, nil
	}

	if fieldKind == reflect.Map {
		v, err := parseMap(field)
		if err != nil {
			return reflect.Value{}, false, fmt.Errorf("parse %q field: %w", field.Name, err)
		}
		fv := reflect.New(field.Type).Elem()
		fv.Set(v)
		return v, true, nil
	}

	envKey, ok := field.Tag.Lookup("env")
	if !ok {
		return reflect.Value{}, false, nil
	}

	s := os.Getenv(envKey)
	return parseValue(s, field.Type)
}

func parseValue(value string, t reflect.Type) (reflect.Value, bool, error) {
	kind := t.Kind()

	if value == "" && valueRequired(kind) {
		return reflect.Value{}, false, nil
	}

	switch kind {
	case reflect.String:
		return reflect.ValueOf(value), true, nil
	case reflect.Int:
		n, err := strconv.ParseInt(value, 10, strconv.IntSize)
		return reflect.ValueOf(int(n)), err == nil, err
	case reflect.Int8:
		n, err := strconv.ParseInt(value, 10, 8)
		return reflect.ValueOf(int8(n)), err == nil, err
	case reflect.Int16:
		n, err := strconv.ParseInt(value, 10, 16)
		return reflect.ValueOf(int16(n)), err == nil, err
	case reflect.Int32:
		n, err := strconv.ParseInt(value, 10, 32)
		return reflect.ValueOf(int32(n)), err == nil, err
	case reflect.Int64:
		n, err := strconv.ParseInt(value, 10, 64)
		return reflect.ValueOf(n), err == nil, err
	case reflect.Uint:
		n, err := strconv.ParseUint(value, 10, strconv.IntSize)
		return reflect.ValueOf(uint(n)), err == nil, err
	case reflect.Uint8:
		n, err := strconv.ParseUint(value, 10, 8)
		return reflect.ValueOf(uint8(n)), err == nil, err
	case reflect.Uint16:
		n, err := strconv.ParseUint(value, 10, 16)
		return reflect.ValueOf(uint16(n)), err == nil, err
	case reflect.Uint32:
		n, err := strconv.ParseUint(value, 10, 32)
		return reflect.ValueOf(uint32(n)), err == nil, err
	case reflect.Uint64:
		n, err := strconv.ParseUint(value, 10, 64)
		return reflect.ValueOf(uint64(n)), err == nil, err
	case reflect.Complex64:
		c, err := strconv.ParseComplex(value, 64)
		return reflect.ValueOf(complex64(c)), err == nil, err
	case reflect.Complex128:
		c, err := strconv.ParseComplex(value, 128)
		return reflect.ValueOf(c), err == nil, err
	case reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		return reflect.ValueOf(f), err == nil, err
	case reflect.Float32:
		f, err := strconv.ParseFloat(value, 32)
		return reflect.ValueOf(float32(f)), err == nil, err
	case reflect.Bool:
		return reflect.ValueOf(parseBool(value)), true, nil
	case reflect.Array:
		vals := mapSlice(strings.Split(value, ","), strings.TrimSpace)
		return parseArray(vals, t)
	case reflect.Slice:
		vals := mapSlice(strings.Split(value, ","), strings.TrimSpace)
		return parseSlice(vals, t)
	case reflect.Pointer:
		v, ok, err := parseValue(value, t.Elem())
		if err != nil {
			return reflect.Value{}, false, err
		}
		if !ok {
			return reflect.Value{}, false, nil
		}
		p := reflect.New(v.Type())
		p.Elem().Set(v)
		return p, true, nil

	default:
		return reflect.Value{}, false, fmt.Errorf("unsupported Kind: %q", t.Kind())
	}
}

func parseArray(vals []string, t reflect.Type) (reflect.Value, bool, error) {
	out := reflect.New(t).Elem()

	len := out.Len()
	for i, val := range vals {
		if len <= i {
			break
		}

		el := out.Index(i)

		v, ok, err := parseValue(val, el.Type())
		if err != nil {
			return reflect.Value{}, false, fmt.Errorf("parse array value %q of kind %q: %w", val, el.Kind(), err)
		}

		if ok {
			el.Set(v)
		}
	}

	return out, true, nil
}

func parseSlice(vals []string, t reflect.Type) (reflect.Value, bool, error) {
	out := reflect.MakeSlice(t, len(vals), cap(vals))

	for i, val := range vals {
		el := out.Index(i)

		v, ok, err := parseValue(val, el.Type())
		if err != nil {
			return reflect.Value{}, false, fmt.Errorf("parse array value %q of kind %q: %w", val, el.Kind(), err)
		}

		if ok {
			el.Set(v)
		}
	}

	return out, true, nil
}

func parseMap(field reflect.StructField) (reflect.Value, error) {
	ft := field.Type
	ftk := ft.Key()
	vt := ft.Elem()

	mt := reflect.MapOf(ftk, vt)

	prefix := field.Tag.Get("env")
	if prefix != "" {
		prefix = prefix + "_"
	}

	out := reflect.MakeMap(mt)

	var found int
	for _, env := range os.Environ() {
		split := strings.Split(env, "=")
		if len(split) != 2 {
			continue
		}

		key := split[0]
		val := split[1]

		if !strings.HasPrefix(key, prefix) {
			continue
		}

		stripped := strings.TrimPrefix(key, prefix)

		kv, ok, err := parseValue(stripped, ftk)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("parse map key %q of kind %q: %w", key, ftk.Kind(), err)
		}
		if !ok {
			continue
		}

		vv, ok, err := parseValue(val, vt)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("parse map value %q of kind %q [key=%s]: %w", val, vt.Kind(), key, err)
		}
		if !ok {
			continue
		}

		out.SetMapIndex(kv, vv)
		found++
	}

	if found == 0 {
		return reflect.Zero(mt), nil
	}

	return out, nil
}

func parseBool(s string) bool {
	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}
	return s != ""
}

func mapSlice[In, Out any](s []In, fn func(In) Out) []Out {
	if s == nil {
		return nil
	}
	out := make([]Out, len(s))
	for i, v := range s {
		out[i] = fn(v)
	}
	return out
}

var optionalValues = map[reflect.Kind]bool{reflect.Bool: true}

func valueRequired(kind reflect.Kind) bool {
	return !optionalValues[kind]
}

func isStruct(v reflect.Type) (isStruct bool, isPointer bool) {
	kind := v.Kind()
	isPointer = kind == reflect.Pointer
	isStruct = kind == reflect.Struct || (isPointer && v.Elem().Kind() == reflect.Struct)
	return
}
