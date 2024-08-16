package ui

import (
	"reflect"
	"strings"
)

func (c *Controller) getRealURI(handlerURI string, requestURI string) string {
	uri := requestURI[len(handlerURI):]
	xs := strings.SplitN(uri, "?", 2)
	return xs[0]
}

func (c *Controller) setStructNameFunc(uri string, objFuncs ...func() interface{}) {
	// Loop through objFuncs and get their struct names
	if c.uriStructNameFunc == nil {
		c.uriStructNameFunc = make(map[string]map[string]func() interface{})
	}
	c.uriStructNameFunc[uri] = map[string]func() interface{}{}
	for _, obj := range objFuncs {
		o := obj()
		v := reflect.ValueOf(o)
		i := reflect.Indirect(v)
		s := i.Type()
		c.uriStructNameFunc[uri][s.Name()] = obj
	}
}
