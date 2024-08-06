package ui

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"

	"github.com/mikolajgs/crud/pkg/struct2sql"
)	

type StructItemsTplObj struct {
	Name string
	URI string
	Fields []string
}


func (c *Controller) getStructItemsTplObj(uri string, objFunc func() interface{}) (*StructItemsTplObj, error) {
	o := objFunc()

	its := &StructItemsTplObj{
		URI: uri,
		Name: struct2sql.GetStructName(o),
		Fields: struct2sql.GetStructFieldNames(o),
	}

	return its, nil
} 

func (c *Controller) getStructItemsHTML(uri string, objFunc func() interface{}) (string, error) {
	structItemsTpl, err := embed.FS.ReadFile(htmlDir, "html/struct_items.html")
	if err != nil {
		return "", fmt.Errorf("error reading struct items template from embed: %w", err)
	}

	tplObj, err := c.getStructItemsTplObj(uri, objFunc)
	if err != nil {
		return "", fmt.Errorf("error getting struct items for html: %w", err)
	}

	buf := &bytes.Buffer{}
	t := template.Must(template.New("structItems").Parse(string(structItemsTpl)))
	err = t.Execute(buf, &tplObj)
	if err != nil {
		return "", fmt.Errorf("error processing struct items template: %w", err)
	}
	return buf.String(), nil
}
