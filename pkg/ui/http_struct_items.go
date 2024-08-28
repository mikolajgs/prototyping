package ui

import (
	"fmt"
	"net/http"
)


func (c *Controller) tryStructItems(w http.ResponseWriter, r *http.Request, uri string) bool {
	structName, id := c.getStructAndIDFromURI("x/struct_items/", c.getRealURI(uri, r.RequestURI))

	if structName == "" {
		return false
	}

	// Check if struct exists
	_, ok := c.uriStructNameFunc[uri][structName]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return true
	}

	if r.Method != http.MethodPut && r.Method != http.MethodPost && r.Method != http.MethodDelete {
		return false
	}

	if r.Method == http.MethodDelete && id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return true
	}

	_ = c.uriStructNameFunc[uri][structName]()

	// Handle delete here
	if r.Method == http.MethodDelete {
		// Read ids from the param 'ids'
		// Run delete for multiple rows
		c.renderMsg(w, r, MsgSuccess, fmt.Sprintf("%s items have been successfully deleted.", structName))
		return true
	}

	// Any other request is invalid
	w.WriteHeader(http.StatusBadRequest)
	return true
}
