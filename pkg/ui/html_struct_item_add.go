package ui

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"

	"github.com/mikolajgs/crud/pkg/struct2html"
	"github.com/mikolajgs/crud/pkg/struct2sql"
)	

type StructItemAddTplObj struct {
	Name string
	URI string
	FieldsHTML string
	MsgHTML string
	OnlyMsg bool
}

func (c *Controller) getStructItemAddTplObj(uri string, objFunc func() interface{}, postValues map[string]string, msgType int, msg string) (*StructItemAddTplObj, error) {
	o := objFunc()

	onlyMsg := false
	if msgType == MsgSuccess {
		onlyMsg = true
	}

	a := &StructItemAddTplObj{
		URI: uri,
		Name: struct2sql.GetStructName(o),
		FieldsHTML: struct2html.GetFields(o, postValues, false),
		MsgHTML: c.getMsgHTML(msgType, msg),
		OnlyMsg: onlyMsg,
	}

	return a, nil
} 

func (c *Controller) getStructItemAddHTML(uri string, objFunc func() interface{}, postValues map[string]string, msgType int, msg string) (string, error) {
	structItemAddTpl, err := embed.FS.ReadFile(htmlDir, "html/struct_item_add.html")
	if err != nil {
		return "", fmt.Errorf("error reading struct item add template from embed: %w", err)
	}

	tplObj, err := c.getStructItemAddTplObj(uri, objFunc, postValues, msgType, msg)
	if err != nil {
		return "", fmt.Errorf("error getting struct item add for html: %w", err)
	}

	buf := &bytes.Buffer{}
	t := template.Must(template.New("structItems").Parse(string(structItemAddTpl)))
	err = t.Execute(buf, &tplObj)
	if err != nil {
		return "", fmt.Errorf("error processing struct item add template: %w", err)
	}
	return buf.String(), nil
}
