package ssmconfig

import (
	"reflect"
	"sync"
)

type structSpec []fieldSpec

type fieldSpec struct {
	ok bool

	name         string
	t            reflect.Type
	defaultValue string
	required     bool
}

func buildStructSpec(t reflect.Type) (spec structSpec) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		name := f.Tag.Get("ssm")
		if len(name) < 0 {
			spec = append(spec, fieldSpec{ok: false})
			continue
		}

		spec = append(spec, fieldSpec{
			ok:           true,
			name:         name,
			t:            f.Type,
			defaultValue: t.Field(i).Tag.Get("default"),
			required:     t.Field(i).Tag.Get("required") == "true",
		})
	}
	return spec
}

var structSpecCache sync.Map // map[reflect.Type]structSpec

// cachedStructSpec is like buildStructSpec but cached to prevent duplicate work.
func cachedStructSpec(t reflect.Type) structSpec {
	if f, ok := structSpecCache.Load(t); ok {
		return f.(structSpec)
	}
	f, _ := structSpecCache.LoadOrStore(t, buildStructSpec(t))
	return f.(structSpec)
}
