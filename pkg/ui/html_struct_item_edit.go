package ui

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"

	"github.com/mikolajgs/crud/pkg/struct2html"
	"github.com/mikolajgs/crud/pkg/struct2sql"
)	

type StructItemEditTplObj struct {
	Name string
	URI string
	ID string
	FieldsHTML string
	MsgHTML string
}

func (c *Controller) getStructItemEditTplObj(uri string, objFunc func() interface{}, id string, postValues map[string]string, msgType int, msg string) (*StructItemEditTplObj, error) {
	o := objFunc()

	err := c.struct2db.Load(o, id)
	if err != nil {
		return nil, err
	}

	e := &StructItemEditTplObj{
		URI: uri,
		Name: struct2sql.GetStructName(o),
		ID: id,
		FieldsHTML: struct2html.GetFields(o, postValues, true),
		MsgHTML: c.getMsgHTML(msgType, msg),
	}

	return e, nil
} 

func (c *Controller) getStructItemEditHTML(uri string, objFunc func() interface{}, id string, postValues map[string]string, msgType int, msg string) (string, error) {
	structItemEditTpl, err := embed.FS.ReadFile(htmlDir, "html/struct_item_edit.html")
	if err != nil {
		return "", fmt.Errorf("error reading struct item edit template from embed: %w", err)
	}

	tplObj, err := c.getStructItemEditTplObj(uri, objFunc, id, postValues, msgType, msg)
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
