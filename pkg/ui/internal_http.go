package ui

import (
	"bytes"
	"embed"
	"log"
	"net/http"
	"reflect"
	"strings"
	"text/template"
)

func (c *Controller) renderMain(w http.ResponseWriter, r *http.Request, uri string, objFuncs... func() interface{}) {
	indexTpl, _ := embed.FS.ReadFile(htmlDir, "html/index.html")
	
	structListTpl, err := c.getStructListHTML(uri, objFuncs...)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	
	contentHomeTpl, _ := embed.FS.ReadFile(htmlDir, "html/content_home.html")

	tplObj := struct{
		URI string
		StructList string
		Content string
	}{
		URI: uri,
		StructList: structListTpl,
		Content: string(contentHomeTpl),
	}
	buf := &bytes.Buffer{}
	t := template.Must(template.New("index").Parse(string(indexTpl)))
	err = t.Execute(buf, &tplObj)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(buf.Bytes())
}

func (c *Controller) renderStructList(w http.ResponseWriter, r *http.Request, uri string, objFuncs ...func() interface{}) {
	structListTpl, err := c.getStructListHTML(uri, objFuncs...)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
		return
	}
	w.Write([]byte(structListTpl))
}

func (c *Controller) renderStructItems(w http.ResponseWriter, r *http.Request, uri string, objFunc func() interface{}) {
	structListTpl, err := c.getStructItemsHTML(uri, objFunc)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
		return
	}
	w.Write([]byte(structListTpl))
}

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
