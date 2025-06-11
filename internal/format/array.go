package format

import (
	"fmt"
	"reflect"

	"github.com/goccy/go-yaml"
)

type Array[E comparable] []E

func (a *Array[E]) UnmarshalYAML(b []byte) error {
	var singleValue E
	if err := yaml.Unmarshal(b, &singleValue); err == nil {
		*a = Array[E]{singleValue}
		return nil
	}

	if err := yaml.Unmarshal(b, (*[]E)(a)); err == nil {
		return nil
	}

	var raw any
	if err := yaml.Unmarshal(b, &raw); err != nil {
		return fmt.Errorf("cannot unmarshal array: %w", err)
	}

	switch reflect.TypeOf(raw).Kind() {
	case reflect.Slice, reflect.Array:
		return fmt.Errorf("cannot unmarshal %s into Array[%s]: expected element of type %T but got array element of type %T",
			string(b), reflect.TypeFor[E]().String(), *new(E), reflect.TypeOf(raw).Elem())
	default:
		return fmt.Errorf("cannot unmarshal %s into Array[%s]: expected array or single element of type %T but got %T",
			string(b), reflect.TypeFor[E]().String(), *new(E), reflect.TypeOf(raw))
	}
}
