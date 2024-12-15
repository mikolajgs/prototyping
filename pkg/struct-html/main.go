package structhtml

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	validator "github.com/mikolajgs/struct-validator"
)

const TypeInt64 = 64
const TypeInt = 128
const TypeString = 256

func GetFields(u interface{}, values map[string]string, withFieldValues bool, tagName string) string {
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

		htm += fmt.Sprintf("<p><label>%s</label>%s</p>", field.Name, fieldHTMLs[field.Name])

		// TODO: very hacky
		if gotDoubleEntry {
			htmAgain := fmt.Sprintf("<p><label>%s (repeat)</label>%s</p>", field.Name, fieldHTMLs[field.Name])
			htm += reName.ReplaceAllString(htmAgain, fmt.Sprintf("name=\"%s___repeat\"", field.Name))
		}
	}

	return htm
}
