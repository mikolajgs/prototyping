package ui

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

type StructListTplObj struct {
	Structs []*StructListTplObjItem
	URI string
}

type StructListTplObjItem struct {
	Name string
}

func (c *Controller) getStructListTplObj(uri string, structNames []string) (*StructListTplObj, error) {
	l := &StructListTplObj{
		URI: uri,
		Structs: []*StructListTplObjItem{},
	}

	for _, n := range structNames {
		l.Structs = append(l.Structs, &StructListTplObjItem{
			Name: n,
		})
	}

	return l, nil
} 
	
func (c Controller) getStructListHTML(uri string, structNames []string) (string, error) {
	structListTpl, err := embed.FS.ReadFile(htmlDir, "html/struct_list.html")
	if err != nil {
		return "", fmt.Errorf("error reading struct list template from embed: %w", err)
	}

	tplObj, err := c.getStructListTplObj(uri, structNames)
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
