package binder

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	. "github.com/gota33/http-binder/internal"
)

type StringSliceUnmarshaler = IStringSliceUnmarshaler

type UriParamGetter func(req *http.Request, key string) string

type Input interface {
	BindInput(req *http.Request, v any) error
}

type input struct {
	InputConfig
}

type InputConfig struct {
	UriParamGetter UriParamGetter
}

func NewInput(c InputConfig) Input {
	return input{InputConfig: c}
}

func (b input) BindInput(req *http.Request, v any) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("input binder: %w", err)
		}
	}()

	if req.Body != nil {
		defer func() { _ = req.Body.Close() }()
	}

	switch GetContentType(req.Header.Get("Content-Type")) {
	case ContentTypeForm:
		err = req.ParseForm()
	case ContentTypeJSON:
		err = json.NewDecoder(req.Body).Decode(v)
	case ContentTypeXML:
		err = xml.NewDecoder(req.Body).Decode(v)
	}
	if err != nil {
		return
	}

	acc, err := NewAccessor(v)
	if err != nil {
		return
	}

	if err = b.bindValues(acc, TagQuery, req.URL.Query()); err != nil {
		return
	}
	if err = b.bindValues(acc, TagForm, req.PostForm); err != nil {
		return
	}
	if err = b.bindHeader(acc, req.Header); err != nil {
		return
	}
	if err = b.bindUriParam(acc, req); err != nil {
		return
	}

	return
}

func (b input) bindValues(acc Accessor, tagType TagType, values url.Values) error {
	var errs []error
	for field, arr := range values {
		if err := acc.Set(tagType, field, arr...); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (b input) bindHeader(acc Accessor, header http.Header) error {
	var errs []error
	for name, arr := range header {
		name = http.CanonicalHeaderKey(name)
		if err := acc.Set(TagHeader, name, arr...); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (b input) bindUriParam(acc Accessor, req *http.Request) error {
	fields := acc.GetFields(TagUri)
	errs := make([]error, len(fields))

	for i, field := range fields {
		value := b.UriParamGetter(req, field)
		errs[i] = acc.Set(TagUri, field, value)
	}
	return errors.Join(errs...)
}
