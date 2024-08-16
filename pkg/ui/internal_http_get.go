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

func (c *Controller) tryGetStructItemAdd(w http.ResponseWriter, r *http.Request, uri string) bool {
	realURI := c.getRealURI(uri, r.RequestURI)
	if strings.HasPrefix(realURI, "x/struct_item_add/") {
		structName := strings.Replace(realURI, "x/struct_item_add/", "", 1)
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
		
		c.renderStructItemAdd(w, r, uri, c.uriStructNameFunc[uri][structName], map[string]string{}, 0, "")
		return true
	}
	return false
}

func (c *Controller) tryGetStructItemEdit(w http.ResponseWriter, r *http.Request, uri string) bool {
	realURI := c.getRealURI(uri, r.RequestURI)
	if strings.HasPrefix(realURI, "x/struct_item_edit/") {
		structNameAndID := strings.Replace(realURI, "x/struct_item_edit/", "", 1)
		matched, err := regexp.MatchString(`^[a-zA-Z0-9_]+/[0-9]+$`, structNameAndID)
		if err != nil || !matched {
			w.WriteHeader(http.StatusBadRequest)
			return true
		}
		structNameIDArr := strings.Split(structNameAndID, "/")

		_, ok := c.uriStructNameFunc[uri][structNameIDArr[0]]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return true
		}
		
		c.renderStructItemEdit(w, r, uri, c.uriStructNameFunc[uri][structNameIDArr[0]], structNameIDArr[1])
		return true
	}
	return false
}
