package restapi

import (
	"net/http"
)

type HandlerOptions struct {
	CreateConstructor func() interface{}
	ReadConstructor   func() interface{}
	UpdateConstructor func() interface{}
	ListConstructor   func() interface{}
	Operations        int
	ForceName         string
}

// Values for CRUD operations
const OpAll = 0
const OpRead = 2
const OpUpdate = 4
const OpCreate = 8
const OpDelete = 16
const OpList = 32

// Handler returns a REST API HTTP handler that can be attached to HTTP server. It creates a CRUD endpoint
// for creating, reading, updating, deleting and listing objects.
// Each of the func() argument should be funcs that create new object (instance of a struct). For each of the
// operation (create, read etc.), a different struct with different fields can be used. It's important to pass
// "uri" argument same as the one that the handler is attached to.
func (c Controller) Handler(uri string, constructor func() interface{}, options HandlerOptions) http.Handler {
	c.initHelpers(constructor, options)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, b := c.getIDFromURI(r.RequestURI[len(uri):], w)
		if !b {
			return
		}
		if r.Method == http.MethodPut && id == "" && (options.Operations == OpAll || options.Operations&OpCreate > 0) {
			if options.CreateConstructor != nil {
				c.handleHTTPPut(w, r, options.CreateConstructor, id)
			} else {
				c.handleHTTPPut(w, r, constructor, id)
			}
			return
		}
		if r.Method == http.MethodPut && id != "" && (options.Operations == OpAll || options.Operations&OpUpdate > 0) {
			if options.UpdateConstructor != nil {
				c.handleHTTPPut(w, r, options.UpdateConstructor, id)
			} else {
				c.handleHTTPPut(w, r, constructor, id)
			}
			return
		}

		if r.Method == http.MethodGet && id != "" && (options.Operations == OpAll || options.Operations&OpRead > 0) {
			if options.ReadConstructor != nil {
				c.handleHTTPGet(w, r, options.ReadConstructor, id)
			} else {
				c.handleHTTPGet(w, r, constructor, id)
			}
			return
		}
		if r.Method == http.MethodGet && id == "" && (options.Operations == OpAll || options.Operations&OpList > 0) {
			if options.ListConstructor != nil {
				c.handleHTTPGet(w, r, options.ListConstructor, id)
			} else {
				c.handleHTTPGet(w, r, constructor, id)
			}
			return
		}
		if r.Method == http.MethodDelete && id != "" && (options.Operations == OpAll || options.Operations&OpDelete > 0) {
			c.handleHTTPDelete(w, r, constructor, id)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
	})
}
