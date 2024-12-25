package ui

import (
	"fmt"
	"html"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	validator "github.com/go-phings/struct-validator"
)

func (c *Controller) getStructItemFieldsHTML(u interface{}, values map[string]string, withFieldValues bool, tagName string, intFieldValues map[string]IntFieldValues) string {
	fieldHTMLs := validator.GenerateHTML(u, &validator.HTMLOptions{
		OverwriteTagName: "ui",
		ExcludeFields: map[string]bool{
			"ID": true,
		},
		OverwriteValues: values,
		FieldValues:     true,
	})

	htm := ""

	reName := regexp.MustCompile(`name="[A-Za-z0-9\-_]+"`)

	v := reflect.ValueOf(u)
	i := reflect.Indirect(v)
	s := i.Type()
	structName := s.Name()

	for j := 0; j < s.NumField(); j++ {
		field := s.Field(j)

		if field.Name == "ID" {
			continue
		}

		gotDoubleEntry := false
		fieldTag := field.Tag.Get(tagName)
		if fieldTag != "" {
			fieldTags := strings.Split(fieldTag, " ")
			for _, ft := range fieldTags {
				if ft == "dblentry" {
					gotDoubleEntry = true
					break
				}
			}
		}

		var fieldHTML string
		if c.intFieldValues != nil {
			fv, ok := c.intFieldValues[structName+"_"+field.Name]
			if ok {
				if fv.Type == ValuesMultipleBitChoice {
					for ok, ov := range fv.Values {
						checked := ""
						if values[field.Name] != "" {
							i64, err := strconv.ParseInt(values[field.Name], 10, 64)
							if err == nil {
								if i64&int64(ok) > 0 {
									checked = " checked"
								}
							}
						}
						fieldHTML += fmt.Sprintf(`<input%s type="checkbox" name="%s" value="%d"/> %s`, checked, field.Name, ok, html.EscapeString(ov))
					}
				}
				if fv.Type == ValuesSingleChoice {
					fieldHTML = fmt.Sprintf(`<select name="%s">`, field.Name)
					for ok, ov := range fv.Values {
						selected := ""
						if values[field.Name] == fmt.Sprintf("%d", ok) {
							selected = " selected"
						}
						fieldHTML += fmt.Sprintf(`<option%s value="%d">%s</option>`, selected, ok, html.EscapeString(ov))
					}
					fieldHTML += "</select>"
				}
			}
		}

		if c.stringFieldValues != nil {
			fv, ok := c.stringFieldValues[structName+"_"+field.Name]
			if ok {
				if fv.Type == ValuesSingleChoice {
					fieldHTML = fmt.Sprintf(`<select name="%s">`, field.Name)
					for ok, ov := range fv.Values {
						selected := ""
						if values[field.Name] == ok {
							selected = " selected"
						}
						fieldHTML += fmt.Sprintf(`<option%s value="%s">%s</option>`, selected, html.EscapeString(ok), html.EscapeString(ov))
					}
					fieldHTML += "</select>"
				}
			}
		}

		if fieldHTML == "" {
			fieldHTML = fieldHTMLs[field.Name]
		}

		htm += fmt.Sprintf("<p><label>%s</label>%s</p>", field.Name, fieldHTML)

		// TODO: very hacky
		if gotDoubleEntry {
			htmAgain := fmt.Sprintf("<p><label>%s (repeat)</label>%s</p>", field.Name, fieldHTMLs[field.Name])
			htm += reName.ReplaceAllString(htmAgain, fmt.Sprintf("name=\"%s___repeat\"", field.Name))
		}
	}

	return htm
}
