package ui

import (
	"bytes"
	"embed"
	"fmt"
	"reflect"
	"text/template"
)

type StructListTplObj struct {
	Structs []*StructListTplObjItem
	URI string
}

type StructListTplObjItem struct {
	Name string
}

func (c *Controller) getStructListTplObj(uri string, objFuncs... func() interface{}) (*StructListTplObj, error) {
	l := &StructListTplObj{
		URI: uri,
		Structs: []*StructListTplObjItem{},
	}

	for _, objFunc := range objFuncs {
		o := objFunc()
		v := reflect.ValueOf(o)
		i := reflect.Indirect(v)
		s := i.Type()
		l.Structs = append(l.Structs, &StructListTplObjItem{
			Name: s.Name(),
		})
	}

	return l, nil
} 
	
func (c *Controller) getStructListHTML(uri string, objFuncs... func() interface{}) (string, error) {
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
