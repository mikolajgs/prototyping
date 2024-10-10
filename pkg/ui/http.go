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
			if c.tryGetHome(w, r, uri, objFuncs...) {
				return
			}
			if c.tryGetStructList(w, r, uri, objFuncs...) {
				return
			}
			if c.tryGetStructItems(w, r, uri) {
				return
			}
			if c.tryGetStructItem(w, r, uri) {
				return
			}
		}

		if c.tryStructItem(w, r, uri) {
			return
		}
		if c.tryStructItems(w, r, uri) {
			return
		}

		w.WriteHeader(http.StatusBadRequest)
	})
}

func (c *Controller) getStructAndIDFromURI(prefix string, uri string) (string, string) {
	if prefix != "" {
		uri = strings.Replace(uri, prefix, "", 1)
	}

	structName := ""
	id := ""

	re := regexp.MustCompile(`^([a-zA-Z0-9_]+)/([0-9]+)$`)
	re2 := regexp.MustCompile(`^([a-zA-Z0-9_]+)/$`)

	// Try matching to a URI with ID (edit item)
	matched := re.FindStringSubmatch(uri)
	if len(matched) == 3 {
		structName = matched[1]
		id = matched[2]
	}

	// Try matching to a URI without ID (add item)
	if structName == "" {
		matched2 := re2.FindStringSubmatch(uri)
		if len(matched2) == 2 {
			structName = matched2[1]
		}
	}

	return structName, id
}
