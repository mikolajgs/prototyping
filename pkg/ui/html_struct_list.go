package ui

import (
	"bytes"
	"embed"
	"fmt"
	"reflect"
	"text/template"
)

type structListTplObj struct {
	Structs []*structListTplObjItem
	URI     string
}

type structListTplObjItem struct {
	Name string
}

func (c *Controller) getStructListTplObj(uri string, objFuncs ...func() interface{}) (*structListTplObj, error) {
	l := &structListTplObj{
		URI:     uri,
		Structs: []*structListTplObjItem{},
	}

	for _, objFunc := range objFuncs {
		o := objFunc()
		v := reflect.ValueOf(o)
		i := reflect.Indirect(v)
		s := i.Type()
		l.Structs = append(l.Structs, &structListTplObjItem{
			Name: s.Name(),
		})
	}

	return l, nil
}

func (c *Controller) getStructListHTML(uri string, objFuncs ...func() interface{}) (string, error) {
	structListTpl, err := embed.FS.ReadFile(htmlDir, "html/struct_list.html")
	if err != nil {
		return "", fmt.Errorf("error reading struct list template from embed: %w", err)
	}

	tplObj, err := c.getStructListTplObj(uri, objFuncs...)
	if err != nil {
		return "", fmt.Errorf("error getting struct list for html: %w", err)
	}

	buf := &bytes.Buffer{}
	t := template.Must(template.New("structList").Parse(string(structListTpl)))
	err = t.Execute(buf, &tplObj)
	if err != nil {
		return "", fmt.Errorf("error processing struct list template: %w", err)
	}

	return buf.String(), nil
}
