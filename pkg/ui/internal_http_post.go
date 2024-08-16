package ui

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	validator "github.com/mikolajgs/struct-validator"
)

func (c *Controller) tryPostStructItemAdd(w http.ResponseWriter, r *http.Request, uri string) bool {
	realURI := c.getRealURI(uri, r.RequestURI)
	if strings.HasPrefix(realURI, "x/struct_item_add/") {
		structName := strings.Replace(realURI, "x/struct_item_add/", "", 1)
		matched, err := regexp.MatchString(`^[a-zA-Z0-9_]+/$`, structName)
		if err != nil || !matched {
			w.WriteHeader(http.StatusBadRequest)
			return true
		}
		structName = strings.Replace(structName, "/", "", 1)

		_, ok := c.uriStructNameFunc[uri][structName]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return true
		}
		
		// Get form data
		r.ParseForm()

		// Create object, set value and validate it
		invalidFormFields := map[string]bool{}

		// Value for each form key is actually an array of strings. We're taking the first one here only
		// TODO: Tweak it
		obj := c.uriStructNameFunc[uri][structName]()
		s := reflect.ValueOf(obj).Elem()

		postValues := map[string]string{}
		for fk, fv := range r.Form {
			postValues[fk] = fv[0];

			if fv[0] == "" {
				continue
			}

			f := s.FieldByName(fk)
			if f.IsValid() && f.CanSet() {
				if f.Kind() == reflect.String {
					f.SetString(fv[0])
				}

				if f.Kind() == reflect.Int || f.Kind() == reflect.Int64 {
					i, err := strconv.ParseInt(fv[0], 10, 64)
					if err != nil {
						invalidFormFields[fk] = true
						continue
					}

					f.SetInt(i)
				}
			}
		}

		valid, failedFields := validator.Validate(obj, &validator.ValidationOptions{
			OverwriteTagName: "ui",
		})

		if len(invalidFormFields) > 0 {
			for k, _ := range invalidFormFields {
				failedFields[k] = failedFields[k] | validator.FailRegexp
			}
		}

		if !valid || len(failedFields) > 0 {
			invVals := []string{}
			for k, _ := range failedFields {
				invVals = append(invVals, k)
			}
			c.renderStructItemAdd(w, r, uri, c.uriStructNameFunc[uri][structName], postValues, MsgFailure, fmt.Sprintf("The following fields have invalid values: %s", strings.Join(invVals, ",")))
			return true
		}

		err2 := c.struct2db.Save(obj)
		if err2 != nil {
			c.renderStructItemAdd(w, r, uri, c.uriStructNameFunc[uri][structName], postValues, MsgFailure, fmt.Sprintf("Problem with saving: %s", err2.Unwrap().Error()))
			return true
		}

		c.renderMsg(w, r, MsgSuccess, fmt.Sprintf("%s item has been successfully added.", structName))
	
		return true
	}
	return false
}

