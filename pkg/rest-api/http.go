package restapi

import "net/http"

// GetHTTPHandler returns a REST API HTTP handler that can be attached to HTTP server. It creates a CRUD endpoint
// for creating, reading, updating, deleting and listing objects.
// Each of the func() argument should be funcs that create new object (instance of a struct). For each of the
// operation (create, read etc.), a different struct with different fields can be used. It's important to pass
// "uri" argument same as the one that the handler is attached to.
func (c Controller) GetHTTPHandler(uri string, newObjFunc func() interface{}, newObjCreateFunc func() interface{}, newObjReadFunc func() interface{}, newObjUpdateFunc func() interface{}, newObjDeleteFunc func() interface{}, newObjListFunc func() interface{}) http.Handler {
	c.initHelpersForHTTPHandler(newObjFunc, newObjCreateFunc, newObjReadFunc, newObjUpdateFunc, newObjDeleteFunc, newObjListFunc)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, b := c.getIDFromURI(r.RequestURI[len(uri):], w)
		if !b {
			return
		}
		if r.Method == http.MethodPut && id == "" {
			c.handleHTTPPut(w, r, newObjCreateFunc, id)
			return
		}
		if r.Method == http.MethodPut && id != "" {
			c.handleHTTPPut(w, r, newObjUpdateFunc, id)
			return
		}
		if r.Method == http.MethodGet && id != "" {
			c.handleHTTPGet(w, r, newObjReadFunc, id)
			return
		}
		if r.Method == http.MethodGet && id == "" {
			c.handleHTTPGet(w, r, newObjListFunc, id)
			return
		}
		if r.Method == http.MethodDelete && id != "" {
			c.handleHTTPDelete(w, r, newObjFunc, id)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
	})
}
