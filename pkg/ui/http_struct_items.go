package ui

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)


func (c *Controller) tryStructItems(w http.ResponseWriter, r *http.Request, uri string) bool {
	structName, id := c.getStructAndIDFromURI("x/struct_items/", c.getRealURI(uri, r.RequestURI))

	if structName == "" || id != "" {
		return false
	}

	// Check if struct exists
	newObjFunc, ok := c.uriStructNameFunc[uri][structName]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return true
	}

	if r.Method != http.MethodPut && r.Method != http.MethodPost && r.Method != http.MethodDelete {
		return false
	}

	_ = c.uriStructNameFunc[uri][structName]()

	// Handle delete here
	if r.Method == http.MethodDelete {
		// Read ids from the param 'ids'
		ids := r.URL.Query().Get("ids")
		match, err := regexp.MatchString(`^[0-9]+(,[0-9]+)*$`, ids)
		if err != nil || !match {
			w.WriteHeader(http.StatusBadRequest)
			return true
		}

		// Run delete for multiple rows
		// TODO: Replace it with a single delete happening on many IDs
		idsList := strings.Split(ids, ",")
		errInLoop := false
		for _, id := range idsList {
			obj := newObjFunc()

			val := reflect.ValueOf(obj).Elem()
			valField := val.FieldByName("ID")
			if !valField.CanSet() {
				errInLoop = true
				continue
			}
			i, _ := strconv.ParseInt(id, 10, 64)
			valField.SetInt(i)

			err2 := c.struct2db.Delete(obj)
			if err2 != nil {
				errInLoop = true
				continue
			}
		}

		if errInLoop {
			c.renderMsg(w, r, MsgFailure, fmt.Sprintf("Problem with removing %s items.", structName))
			return true
		}

		c.renderMsg(w, r, MsgSuccess, fmt.Sprintf("%s items have been successfully deleted.", structName))
		return true
	}

	// Any other request is invalid
	w.WriteHeader(http.StatusBadRequest)
	return true
}
