package ssmconfig

import (
	"encoding"
	"errors"
	"path"
	"reflect"
	"strconv"
)

var (
	ErrUnknownType  = errors.New("ssmconfig: unknown type")
	ErrMissingValue = errors.New("ssmconfig: missing data for key")
)

var (
	textUnmarshalerType = reflect.TypeOf(new(encoding.TextUnmarshaler)).Elem()
)

type Values map[string]string

type valueSetter func(v reflect.Value, p string, data Values) error

func newValueSetter(t reflect.Type) valueSetter {
	if t.Implements(textUnmarshalerType) {
		return textUnmarshalerSetter(t, false)
	}
	if reflect.PtrTo(t).Implements(textUnmarshalerType) {
		return textUnmarshalerSetter(t, true)
	}

	switch t.Kind() {
	case reflect.Ptr:
		return ptrSetter(t)
	case reflect.Interface:
		return interfaceSetter
	case reflect.String:
		return stringSetter
	case reflect.Struct:
		return structSetter(t)
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return intSetter
	case reflect.Float32:
		return floatSetter(32)
	case reflect.Float64:
		return floatSetter(64)
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		return uintSetter
	case reflect.Bool:
		return boolSetter
	}
	return unknownSetter

}

func structSetter(t reflect.Type) valueSetter {
	spec := cachedStructSpec(t)
	return func(v reflect.Value, p string, data Values) error {
		for i, fieldSpec := range spec {
			if !fieldSpec.ok {
				continue
			}

			vs := newValueSetter(fieldSpec.t)
			if fieldSpec.required {
				vs = requiredSetter(vs)
			}
			if len(fieldSpec.defaultValue) > 0 {
				vs = defaultSetter(fieldSpec.defaultValue, vs)
			}

			if err := vs(v.Field(i), path.Join(p, fieldSpec.name), data); err != nil {
				return err
			}
		}
		return nil
	}
}

func unknownSetter(_ reflect.Value, _ string, _ Values) error {
	return ErrUnknownType
}

func nilSetter(v reflect.Value, _ string, _ Values) error {
	if v.IsNil() {
		return nil
	}

	if v.CanSet() {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}

	v = reflect.Indirect(v)
	if v.CanSet() {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}

	return nil
}

func requiredSetter(next valueSetter) valueSetter {
	return func(v reflect.Value, p string, data Values) error {
		if _, ok := data[p]; !ok {
			// TODO: add better error type
			return ErrMissingValue
		}

		return next(v, p, data)
	}
}

func defaultSetter(def string, next valueSetter) valueSetter {
	return func(v reflect.Value, p string, data Values) error {
		if _, ok := data[p]; !ok {
			return next(v, p, map[string]string{p: def})
		}

		return next(v, p, data)
	}
}

func textUnmarshalerSetter(t reflect.Type, shouldAddr bool) valueSetter {
	return func(v reflect.Value, p string, data Values) error {
		if shouldAddr {
			v = v.Addr()
		}

		// set to zero value if this is nil
		if t.Kind() == reflect.Ptr && v.IsNil() {
			v.Set(reflect.New(t.Elem()))
		}

		s := data[p]
		return v.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(s))
	}
}

func interfaceSetter(v reflect.Value, p string, data Values) error {
	elm := v.Elem()
	return newValueSetter(elm.Type())(elm, p, data)
}

func ptrSetter(t reflect.Type) valueSetter {
	innerSetter := newValueSetter(t.Elem())
	return func(v reflect.Value, p string, data Values) error {
		_, ok := data[p]
		if !ok {
			return nilSetter(v, p, data)
		}

		if v.IsNil() {
			v.Set(reflect.New(t.Elem()))
		}
		return innerSetter(v, p, data)
	}
}

func stringSetter(v reflect.Value, p string, data Values) error {
	v.SetString(data[p])
	return nil
}

func intSetter(v reflect.Value, p string, data Values) error {
	s := data[p]

	if len(s) < 1 {
		v.SetInt(0)
		return nil
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	v.SetInt(int64(i))
	return nil
}

func floatSetter(bitSize int) valueSetter {
	return func(v reflect.Value, p string, data Values) error {
		s := data[p]
		if len(s) < 1 {
			v.SetFloat(0)
			return nil
		}
		f, err := strconv.ParseFloat(s, bitSize)
		if err != nil {
			return err
		}
		v.SetFloat(f)
		return nil
	}
}

func uintSetter(v reflect.Value, p string, data Values) error {
	s := data[p]
	if len(s) < 1 {
		v.SetUint(0)
		return nil
	}
	i, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	v.SetUint(uint64(i))
	return nil
}

func boolSetter(v reflect.Value, p string, data Values) error {
	s := data[p]
	if len(s) < 1 {
		v.SetBool(false)
		return nil
	}

	val, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}

	v.SetBool(val)
	return nil
}
