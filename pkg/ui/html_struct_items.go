package ui

import (
	"bytes"
	"embed"
	"fmt"
	"html"
	"reflect"
	"strings"
	"text/template"

	"github.com/mikolajgs/crud/pkg/struct2db"
	"github.com/mikolajgs/crud/pkg/struct2sql"
)	

type structItemsTplObj struct {
	Name string
	URI string
	Fields []string
	ItemsHTML []interface{}
}

func (c *Controller) getStructItemsTplObj(uri string, objFunc func() interface{}) (*structItemsTplObj, error) {
	o := objFunc()

	itemsHTML, err := c.struct2db.Get(objFunc, struct2db.GetOptions{
		RowObjTransformFunc: func(obj interface{}) interface{}{
			out := ""
			id := ""

			v := reflect.ValueOf(obj)
			elem := v.Elem()
			i := reflect.Indirect(v)
			s := i.Type()
			for j := 0; j < s.NumField(); j++ {
				out += "<td>"
				field := s.Field(j)
				fieldType := field.Type.Kind()
				if fieldType == reflect.String {
					out += html.EscapeString(elem.Field(j).String())
				}
				if fieldType == reflect.Bool {
					out += fmt.Sprintf("%v", elem.Field(j).Bool())
				}
				if fieldType == reflect.Int || fieldType == reflect.Int64 {
					out += fmt.Sprintf("%d", elem.Field(j).Int())
					if field.Name == "ID" {
						id = fmt.Sprintf("%d", elem.Field(j).Int())
					}
				}
				out += "</td>"
			}

			return fmt.Sprintf("%s:%s", id, out)
		},
	})
	if err != nil {
		return nil, err
	}

	its := &structItemsTplObj{
		URI: uri,
		Name: struct2sql.GetStructName(o),
		Fields: struct2sql.GetStructFieldNames(o),
		ItemsHTML: itemsHTML,
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
	t := template.Must(template.New("structItems").Funcs(template.FuncMap{
		"SplitRow": func(s string) ([]string) {
			sArr := strings.SplitN(s, ":", 2)
			return sArr
		},
	}).Parse(string(structItemsTpl)))
	err = t.Execute(buf, &tplObj)
	if err != nil {
		return "", fmt.Errorf("error processing struct items template: %w", err)
	}
	return buf.String(), nil
}
