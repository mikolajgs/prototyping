package ui

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"

	"github.com/mikolajgs/crud/pkg/struct2sql"
)	

type StructItemEditTplObj struct {
	Name string
	URI string
	Fields []string
	Obj interface{}
}

func (c *Controller) getStructItemEditTplObj(uri string, objFunc func() interface{}, id string) (*StructItemEditTplObj, error) {
	o := objFunc()

	c.struct2db.Load(o, id)

	e := &StructItemEditTplObj{
		URI: uri,
		Name: struct2sql.GetStructName(o),
		Fields: struct2sql.GetStructFieldNames(o),
		Obj: o,
	}

	return e, nil
} 

func (c *Controller) getStructItemEditHTML(uri string, objFunc func() interface{}, id string) (string, error) {
	structItemEditTpl, err := embed.FS.ReadFile(htmlDir, "html/struct_item_add.html")
	if err != nil {
		return "", fmt.Errorf("error reading struct item edit template from embed: %w", err)
	}

	tplObj, err := c.getStructItemEditTplObj(uri, objFunc, id)
	if err != nil {
		return "", fmt.Errorf("error getting struct item edit for html: %w", err)
	}

	buf := &bytes.Buffer{}
	t := template.Must(template.New("structItems").Parse(string(structItemEditTpl)))
	err = t.Execute(buf, &tplObj)
	if err != nil {
		return "", fmt.Errorf("error processing struct item edit template: %w", err)
	}
	return buf.String(), nil
}
