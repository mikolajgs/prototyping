package ui

import (
	"net/http"
	"reflect"
	"regexp"
	"strings"
)

// GetHTTPHandler returns an HTTP handler that can be attached to HTTP server. It runs a simple UI that allows
// managing the data.
// Each of the func() argument should be funcs that create objects that are meant to be managed in the UI.
func (c Controller) GetHTTPHandler(uri string, xobj ...func() interface{}) http.Handler {
	// Loop through xobj and get their struct names
	if c.uriStructNameFunc == nil {
		c.uriStructNameFunc = make(map[string]map[string]func() interface{})
		c.uriStructArgsName = make(map[string]map[int]string)
		c.structNames = make(map[string][]string)
	}
	c.uriStructNameFunc[uri] = map[string]func() interface{}{}
	c.uriStructArgsName[uri] = map[int]string{}
	structNames := []string{}
	for idx, objFunc := range xobj {
		o := objFunc()
		v := reflect.ValueOf(o)
		i := reflect.Indirect(v)
		s := i.Type()
		c.uriStructNameFunc[uri][s.Name()] = objFunc
		c.uriStructArgsName[uri][idx] = s.Name()
		structNames = append(structNames, s.Name())
	}
	c.structNames[uri] = structNames

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			realURI := c.getRealURI(uri, r.RequestURI)
			if realURI == "x/struct_list/" {
				c.renderStructList(w, r, uri, c.structNames[uri])
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

				objFunc, ok := c.uriStructNameFunc[uri][structName]
				if !ok {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				
				c.renderStructItems(w, r, uri, structName, objFunc)
				return
			}

			if realURI == "" {
				c.renderMain(w, r, uri, c.structNames[uri])
				return
			}
		}

		w.WriteHeader(http.StatusBadRequest)
	})
}
