package ui

import (
	"bytes"
	"embed"
	"log"
	"net/http"
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
	tpl, err := c.getStructListHTML(uri, objFuncs...)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
		return
	}
	w.Write([]byte(tpl))
}

func (c *Controller) renderStructItems(w http.ResponseWriter, r *http.Request, uri string, objFunc func() interface{}) {
	tpl, err := c.getStructItemsHTML(uri, objFunc)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
		return
	}
	w.Write([]byte(tpl))
}

func (c *Controller) renderStructItemAdd(w http.ResponseWriter, r *http.Request, uri string, objFunc func() interface{}, postValues map[string]string, msgType int, msg string) {
	tpl, err := c.getStructItemAddHTML(uri, objFunc, postValues, msgType, msg)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
		return
	}
	w.Write([]byte(tpl))
}

func (c *Controller) renderStructItemEdit(w http.ResponseWriter, r *http.Request, uri string, objFunc func() interface{}, id string) {
	tpl, err := c.getStructItemEditHTML(uri, objFunc, id)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
		return
	}
	w.Write([]byte(tpl))
}

func (c *Controller) renderMsg(w http.ResponseWriter, r *http.Request, msgType int, msg string) {
	w.Write([]byte(c.getMsgHTML(msgType, msg)))
}
