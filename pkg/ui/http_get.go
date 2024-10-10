package ui

import (
	"net/http"
	"regexp"
	"strings"
)

func (c *Controller) tryGetHome(w http.ResponseWriter, r *http.Request, uri string, objFuncs ...func() interface{}) bool {
	realURI := c.getRealURI(uri, r.RequestURI)
	if realURI == "" {
		c.renderMain(w, r, uri, objFuncs...)
		return true
	}
	return false
}

func (c *Controller) tryGetStructList(w http.ResponseWriter, r *http.Request, uri string, objFuncs ...func() interface{}) bool {
	realURI := c.getRealURI(uri, r.RequestURI)
	if realURI == "x/struct_list/" {
		c.renderStructList(w, r, uri, objFuncs...)
		return true
	}
	return false
}

func (c *Controller) tryGetStructItems(w http.ResponseWriter, r *http.Request, uri string) bool {
	realURI := c.getRealURI(uri, r.RequestURI)
	if strings.HasPrefix(realURI, "x/struct_items/") {
		structName := strings.Replace(realURI, "x/struct_items/", "", 1)
		matched, err := regexp.MatchString(`^[a-zA-Z0-9_]+/$`, structName)
		if err != nil || !matched {
			w.WriteHeader(http.StatusBadRequest)
			return true
		}
		structName = strings.Replace(structName, "/", "", 1)

		_, ok := c.uriStructNameFunc[uri][structName]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return true
		}

		c.renderStructItems(w, r, uri, c.uriStructNameFunc[uri][structName])
		return true
	}
	return false
}

func (c *Controller) tryGetStructItem(w http.ResponseWriter, r *http.Request, uri string) bool {
	structName, id := c.getStructAndIDFromURI("x/struct_item/", c.getRealURI(uri, r.RequestURI))

	if structName == "" {
		return false
	}

	// Check if struct exists
	_, ok := c.uriStructNameFunc[uri][structName]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return true
	}

	// Render the page
	c.renderStructItem(w, r, uri, c.uriStructNameFunc[uri][structName], id, map[string]string{}, 0, "")

	return true
}
