## Example

```go
package main

import (
	"net/http"
	"github.com/gota33/http-binder"
	"github.com/go-chi/chi/v5"
)

var b = binder.NewInput(binder.InputConfig{
	UriParamGetter: chi.URLParam,
})

func handle(w http.ResponseWriter, r *http.Request) {
	var request struct {
		H string `header:"h"`
		Q string `query:"q"`
		F string `form:"f"`
		U string `uri:"u"`
		J string `json:"j"`
		X string `xml:"x"`
	}

	if err := b.BindInput(r, &request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
```