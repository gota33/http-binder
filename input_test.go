package binder

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBind(t *testing.T) {
	b := NewInput(InputConfig{})

	t.Run("query", func(t *testing.T) {
		req, err := http.NewRequest(
			"GET",
			"https://domain.com/pp?q0=0&q1=1",
			nil,
		)
		require.NoError(t, err)

		var r struct {
			Q0 string `query:"q0"`
			Q1 string `query:"q1"`
		}
		err = b.BindInput(req, &r)
		require.NoError(t, err)

		require.Equal(t, "0", r.Q0)
		require.Equal(t, "1", r.Q1)
	})

	t.Run("form", func(t *testing.T) {
		const ContentType = "application/x-www-form-urlencoded"

		form := url.Values{}
		form.Set("f0", "f0")
		form.Set("f1", "f1")

		req, err := http.NewRequest(
			"POST",
			"https://domain.com/pp",
			strings.NewReader(form.Encode()),
		)
		require.NoError(t, err)

		req.Header.Set("Content-Type", ContentType)

		var r struct {
			F0          string `form:"f0"`
			F1          string `form:"f1"`
			ContentType string `header:"Content-Type"`
		}
		err = b.BindInput(req, &r)
		require.NoError(t, err)

		require.Equal(t, ContentType, r.ContentType)
		require.Equal(t, "f0", r.F0)
		require.Equal(t, "f1", r.F1)
	})

	t.Run("json", func(t *testing.T) {
		const ContentType = "application/json"

		req, err := http.NewRequest(
			"POST",
			"https://domain.com/pp",
			strings.NewReader(`{"f0": "v0", "f1": 1}`),
		)
		require.NoError(t, err)

		req.Header.Set("Content-Type", ContentType)

		var r struct {
			ContentType string `header:"Content-Type"`
			F0          string `json:"f0"`
			F1          int    `json:"f1"`
		}
		err = b.BindInput(req, &r)
		require.NoError(t, err)

		require.Equal(t, ContentType, r.ContentType)
		require.Equal(t, "v0", r.F0)
		require.Equal(t, 1, r.F1)
	})

	t.Run("uri", func(t *testing.T) {
		params := map[string]string{"id": "123"}
		b := NewInput(InputConfig{
			UriParamGetter: func(req *http.Request, key string) string { return params[key] },
		})

		req, err := http.NewRequest("GET", "https://domain.com/123", nil)
		require.NoError(t, err)

		var r struct {
			ID string `uri:"id"`
		}
		err = b.BindInput(req, &r)
		require.NoError(t, err)

		require.Equal(t, "123", r.ID)
	})
}
