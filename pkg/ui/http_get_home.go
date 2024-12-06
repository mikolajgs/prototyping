package ui

import (
	"bytes"
	"embed"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/mikolajgs/prototyping/pkg/umbrella"
)

func (c *Controller) tryGetHome(w http.ResponseWriter, r *http.Request, uri string, objFuncs ...func() interface{}) bool {
	realURI := c.getRealURI(uri, r.RequestURI)
	if realURI == "" {
		c.renderMain(w, r, uri, objFuncs...)
		return true
	}
	return false
}

func (c *Controller) renderMain(w http.ResponseWriter, r *http.Request, uri string, objFuncs ...func() interface{}) {
	configCss, _ := embed.FS.ReadFile(htmlDir, "html/config.css")
	stylesCss, _ := embed.FS.ReadFile(htmlDir, "html/styles.css")

	indexTpl, _ := embed.FS.ReadFile(htmlDir, "html/index.html")

	structListTpl, err := c.getStructListHTML(uri, objFuncs...)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	contentHomeTpl, _ := embed.FS.ReadFile(htmlDir, "html/content_home.html")

	userId := umbrella.GetUserIDFromRequest(r)
	userName := fmt.Sprintf("%d", userId)
	if userId != 0 {
		userName = r.Context().Value("LoggedUserName").(string)
	}

	tplObj := struct {
		URI        string
		StructList string
		Content    string
		ConfigCss  string
		StylesCss  string
		Username   string
	}{
		URI:        uri,
		StructList: structListTpl,
		Content:    string(contentHomeTpl),
		ConfigCss:  string(configCss),
		StylesCss:  string(stylesCss),
		Username:   userName,
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
