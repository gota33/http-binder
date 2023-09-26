package internal

import (
	"encoding/json"
	"strconv"
	"testing"

	. "github.com/stretchr/testify/require"
)

func TestAccessor(t *testing.T) {
	type QueryInner struct {
		Q0 string  `query:"q0"`
		Q2 string  `query:"q2"`
		Q3 Integer `query:"q3"`
	}
	type FormInner struct {
		F0 string `form:"f0"`
		F1 string `form:"f1"`
	}
	type UriInner struct {
		U0 string `uri:"u0"`
		U1 string `uri:"u1"`
	}
	type Result struct {
		OQ0 *string  `query:"oq0"`
		OF0 *string  `query:"of0"`
		Arr []string `query:"arr"`
		QueryInner
		FormInner
		UriInner
	}

	var r Result
	_, err := NewAccessor(r)
	ErrorIs(t, err, ErrInvalidTarget)

	_, err = NewAccessor(&r.Q0)
	ErrorIs(t, err, ErrInvalidTarget)

	acc, err := NewAccessor(&r)
	NoError(t, err)

	t.Logf("%v", acc.setters)

	NoError(t, acc.Set(TagQuery, "oq0", "oq0"))
	NoError(t, acc.Set(TagQuery, "of0", "of0"))
	NoError(t, acc.Set(TagQuery, "arr", "arr0", "arr1"))
	NoError(t, acc.Set(TagQuery, "q0", "q0"))
	NoError(t, acc.Set(TagQuery, "q1", "q1"))
	NoError(t, acc.Set(TagQuery, "q2", "q2"))
	NoError(t, acc.Set(TagQuery, "q3", "1"))
	NoError(t, acc.Set(TagForm, "f0", "f0"))
	NoError(t, acc.Set(TagForm, "f1", "f1"))
	NoError(t, acc.Set(TagUri, "u0", "u0"))
	NoError(t, acc.Set(TagUri, "u1", "u1"))

	t.Logf("%+v", r)

	data, err := json.Marshal(r)
	NoError(t, err)

	t.Logf("%s", data)
}

type Integer int

func (n *Integer) UnmarshalStringSlice(values []string) (err error) {
	if len(values) == 0 {
		return
	}

	num, err := strconv.ParseInt(values[0], 10, 64)
	if err != nil {
		return
	}

	*n = Integer(num)
	return
}
