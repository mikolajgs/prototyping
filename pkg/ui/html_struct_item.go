package ui

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"

	sthtml "github.com/mikolajgs/prototyping/pkg/struct-html"
	stsql "github.com/mikolajgs/prototyping/pkg/struct-sql-postgres"
)	

type structItemTplObj struct {
	Name string
	URI string
	FieldsHTML string
	MsgHTML string
	OnlyMsg bool
	ID string
}

func (c *Controller) getStructItemTplObj(uri string, objFunc func() interface{}, id string, postValues map[string]string, msgType int, msg string) (*structItemTplObj, error) {
	o := objFunc()

	if id != "" {
		err := c.struct2db.Load(o, id)
		if err != nil {
			return nil, err
		}
	}

	onlyMsg := false
	if msgType == MsgSuccess && id == "" {
		onlyMsg = true
	}

	useFieldValues := false
	if id != "" {
		useFieldValues = true
	}

	a := &structItemTplObj{
		URI: uri,
		Name: stsql.GetStructName(o),
		FieldsHTML: sthtml.GetFields(o, postValues, useFieldValues),
		MsgHTML: c.getMsgHTML(msgType, msg),
		OnlyMsg: onlyMsg,
		ID: id,
	}

	return a, nil
} 


func (c *Controller) getStructItemHTML(uri string, objFunc func() interface{}, id string, postValues map[string]string, msgType int, msg string) (string, error) {
	structItemTpl, err := embed.FS.ReadFile(htmlDir, "html/struct_item.html")
	if err != nil {
		return "", fmt.Errorf("error reading struct item template from embed: %w", err)
	}

	tplObj, err := c.getStructItemTplObj(uri, objFunc, id, postValues, msgType, msg)
	if err != nil {
		return "", fmt.Errorf("error getting struct item for html: %w", err)
	}

	buf := &bytes.Buffer{}
	t := template.Must(template.New("structItem").Parse(string(structItemTpl)))
	err = t.Execute(buf, &tplObj)
	if err != nil {
		return "", fmt.Errorf("error processing struct item template: %w", err)
	}

	return buf.String(), nil
}
