package ui

import (
	"bytes"
	"embed"
	"net/http"
	"strings"
	"text/template"
)

func (c Controller) renderMain(w http.ResponseWriter, r *http.Request, uri string, structNames []string) {
	// Get templates
	indexTpl, _ := embed.FS.ReadFile(htmlDir, "html/index.html")
	
	structListTpl, err := c.getStructListHTML(uri, structNames)
	if err != nil {
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(buf.Bytes())
}

func (c Controller) renderStructList(w http.ResponseWriter, r *http.Request, uri string, structNames []string) {
	structListTpl, err := c.getStructListHTML(uri, structNames)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
		return
	}
	w.Write([]byte(structListTpl))
}

func (c Controller) renderStructItems(w http.ResponseWriter, r *http.Request, uri string, structName string, objFunc func() interface{}) {
	structListTpl, err := c.getStructItemsHTML(uri, structName, objFunc)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
		return
	}
	w.Write([]byte(structListTpl))
}

func (c Controller) getRealURI(handlerURI string, requestURI string) string {
	uri := requestURI[len(handlerURI):]
	xs := strings.SplitN(uri, "?", 2)
	return xs[0]
}
