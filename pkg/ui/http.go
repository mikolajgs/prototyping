package ui

import (
	"net/http"
	"regexp"
	"strings"
)

// GetHTTPHandler returns an HTTP handler that can be attached to HTTP server. It runs a simple UI that allows
// managing the data.
// Each of the func() argument should be funcs that create objects that are meant to be managed in the UI.
func (c *Controller) GetHTTPHandler(uri string, objFuncs ...func() interface{}) http.Handler {
	c.setStructNameFunc(uri, objFuncs...)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			realURI := c.getRealURI(uri, r.RequestURI)
			if realURI == "x/struct_list/" {
				c.renderStructList(w, r, uri, objFuncs...)
				return
			}

			if strings.HasPrefix(realURI, "x/struct_items/") {
				structName := strings.Replace(realURI, "x/struct_items/", "", 1)
				matched, err := regexp.MatchString(`^[a-zA-Z0-9_]+/$`, structName)
				if err != nil || !matched {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				structName = strings.Replace(structName, "/", "", 1)

				_, ok := c.uriStructNameFunc[uri][structName]
				if !ok {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				
				c.renderStructItems(w, r, uri, c.uriStructNameFunc[uri][structName])
				return
			}

			if realURI == "" {
				c.renderMain(w, r, uri, objFuncs...)
				return
			}
		}

		w.WriteHeader(http.StatusBadRequest)
	})
}
