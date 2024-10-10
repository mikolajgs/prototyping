package structhtml

import (
	"fmt"
	"reflect"

	validator "github.com/mikolajgs/struct-validator"
)

const TypeInt64 = 64
const TypeInt = 128
const TypeString = 256

func GetFields(u interface{}, values map[string]string, withFieldValues bool) string {
	fieldHTMLs := validator.GenerateHTML(u, &validator.HTMLOptions{
		OverwriteTagName: "ui",
		ExcludeFields: map[string]bool{
			"ID": true,
		},
		OverwriteValues: values,
		FieldValues:     true,
	})

	htm := ""

	v := reflect.ValueOf(u)
	i := reflect.Indirect(v)
	s := i.Type()
	for j := 0; j < s.NumField(); j++ {
		field := s.Field(j)

		if field.Name == "ID" {
			continue
		}

		htm += fmt.Sprintf("<p><label>%s</label>%s</p>", field.Name, fieldHTMLs[field.Name])
	}

	return htm
}
