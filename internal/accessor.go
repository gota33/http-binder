package internal

import (
	"errors"
	"fmt"
	"reflect"
)

type TagType string

var (
	ErrInvalidTarget = errors.New("invalid target")

	TagQuery  TagType = "query"
	TagForm   TagType = "form"
	TagHeader TagType = "header"
	TagUri    TagType = "uri"

	AllSupportTags = []TagType{
		TagQuery,
		TagForm,
		TagHeader,
		TagUri,
	}
)

type Setters map[TagType]map[string][]reflect.Value

type Accessor struct {
	setters Setters
}

func NewAccessor(target any, tags ...TagType) (out Accessor, err error) {
	size := len(tags)
	if size == 0 {
		tags = AllSupportTags
	}

	setters := make(Setters, size)
	for _, tag := range tags {
		setters[tag] = make(map[string][]reflect.Value)
	}

	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Pointer {
		err = fmt.Errorf("target must be a pointer: %w", ErrInvalidTarget)
		return
	}

	rv = rv.Elem()

	if rv.Kind() != reflect.Struct {
		err = fmt.Errorf("target must be a struct: %w", ErrInvalidTarget)
		return
	}

	if err = travel(setters, rv); err != nil {
		return
	}

	out = Accessor{setters}
	return
}

func travel(setters Setters, rv reflect.Value) (err error) {
	rt := rv.Type()
	num := rt.NumField()

	for i := 0; i < num; i++ {
		// Process struct field
		if rv.Field(i).Kind() == reflect.Struct {
			if err = travel(setters, rv.Field(i)); err != nil {
				return
			}
			continue
		}

		field := rt.Field(i)

		// Skip unexported field
		if !field.IsExported() {
			continue
		}

		// Process exported string or []string field
		for tagType, mapper := range setters {
			if v, ok := field.Tag.Lookup(string(tagType)); ok {
				if err = ensureFieldType(field); err != nil {
					return
				}
				mapper[v] = append(mapper[v], rv.Field(i))
				continue
			}
		}
	}
	return
}

func (a Accessor) GetFields(tag TagType) (arr []string) {
	for name := range a.setters[tag] {
		arr = append(arr, name)
	}
	return
}

func (a Accessor) Set(tag TagType, field string, values ...string) (err error) {
	if len(values) == 0 {
		return
	}

	arr, ok := a.setters[tag][field]
	if !ok {
		return
	}

	for _, setter := range arr {
		var rv reflect.Value
		switch setter.Kind() {
		case reflect.Pointer:
			rv = reflect.ValueOf(&values[0])
		case reflect.String:
			rv = reflect.ValueOf(values[0])
		case reflect.Slice:
			rv = reflect.ValueOf(values)
		}
		setter.Set(rv)
	}
	return
}

func ensureFieldType(field reflect.StructField) (err error) {
	kind := field.Type.Kind()
	if kind == reflect.Pointer {
		kind = field.Type.Elem().Kind()
	}
	if kind != reflect.String && kind != reflect.Slice {
		return fmt.Errorf("field %q must be string or []string: %w", field.Name, ErrInvalidTarget)
	}
	return
}
