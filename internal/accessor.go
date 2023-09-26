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

type IStringSliceUnmarshaler interface {
	UnmarshalStringSlice(values []string) error
}

var RTStringSliceUnmarshaler = reflect.TypeOf((*IStringSliceUnmarshaler)(nil)).Elem()

type Setters map[TagType]map[string][]IStringSliceUnmarshaler

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
		setters[tag] = make(map[string][]IStringSliceUnmarshaler)
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
				var setter IStringSliceUnmarshaler
				if setter, err = getSetter(rv.Field(i)); err != nil {
					err = fmt.Errorf("field %q: %w", field.Name, err)
					return
				}
				mapper[v] = append(mapper[v], setter)
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
		if err = setter.UnmarshalStringSlice(values); err != nil {
			return
		}
	}
	return
}

func getSetter(field reflect.Value) (setter IStringSliceUnmarshaler, err error) {
	// Convert to value type if it's a pointer
	rt := field.Type()
	rv := field
	if kind := rt.Kind(); kind == reflect.Pointer {
		rt = field.Type().Elem()
		rv = field.Elem()
	}

	// Return if it implements StringSliceUnmarshaler
	if reflect.PointerTo(rt).Implements(RTStringSliceUnmarshaler) {
		setter = rv.Addr().Interface().(IStringSliceUnmarshaler)
		return
	}

	// Otherwise only accept string or []string
	if kind := rt.Kind(); kind != reflect.String && kind != reflect.Slice {
		err = fmt.Errorf("field must be string or []string: %w", ErrInvalidTarget)
		return
	}

	setter = FieldBinder{rv: field}
	return
}

type FieldBinder struct {
	rv reflect.Value
}

func (b FieldBinder) UnmarshalStringSlice(values []string) (err error) {
	if len(values) == 0 {
		return
	}

	var target any
	switch b.rv.Type().Kind() {
	case reflect.Pointer:
		target = &values[0]
	case reflect.String:
		target = values[0]
	case reflect.Slice:
		target = values
	}
	if target != nil {
		b.rv.Set(reflect.ValueOf(target))
	}
	return
}
