package ui

import (
	"net/http"
)

// GetHTTPHandler returns an HTTP handler that can be attached to HTTP server. It runs a simple UI that allows
// managing the data.
// Each of the func() argument should be funcs that create objects that are meant to be managed in the UI.
func (c *Controller) GetHTTPHandler(uri string, objFuncs ...func() interface{}) http.Handler {
	c.setStructNameFunc(uri, objFuncs...)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			if c.tryGetHome(w, r, uri, objFuncs...) { return }
			if c.tryGetStructList(w, r, uri, objFuncs...) { return }
			if c.tryGetStructItems(w, r, uri) { return }
			if c.tryGetStructItemAdd(w, r, uri) { return }
			if c.tryGetStructItemEdit(w, r, uri) { return }
		}

		if r.Method == http.MethodPost {
			if c.tryPostStructItemAdd(w, r, uri) { return }
		}

		w.WriteHeader(http.StatusBadRequest)
	})
}
