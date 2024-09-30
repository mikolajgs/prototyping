package ui

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	struct2db "github.com/mikolajgs/prototyping/pkg/struct-db-postgres"
	validator "github.com/mikolajgs/struct-validator"
)

func (c *Controller) tryStructItem(w http.ResponseWriter, r *http.Request, uri string) bool {
	structName, id := c.getStructAndIDFromURI("x/struct_item/", c.getRealURI(uri, r.RequestURI))

	if structName == "" {
		return false
	}

	// Check if struct exists
	_, ok := c.uriStructNameFunc[uri][structName]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return true
	}

	if r.Method != http.MethodPut && r.Method != http.MethodPost && r.Method != http.MethodDelete {
		return false
	}

	if r.Method == http.MethodDelete && id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return true
	}

	obj := c.uriStructNameFunc[uri][structName]()
	// Set ID if present in the URI
	if id != "" {
		val := reflect.ValueOf(obj).Elem()
		valField := val.FieldByName("ID")
		if !valField.CanSet() {
			w.WriteHeader(http.StatusInternalServerError)
			return true
		}
		i, _ := strconv.ParseInt(id, 10, 64)
		valField.SetInt(i)
	}

	// Handle delete here
	if r.Method == http.MethodDelete {
		err2 := c.struct2db.Delete(obj, struct2db.DeleteOptions{})
		if err2 != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return true
		}

		c.renderMsg(w, r, MsgSuccess, fmt.Sprintf("%s item has been successfully deleted.", structName))
		return true
	}

	// Get form data
	r.ParseForm()

	// Create object, set value and validate it
	invalidFormFields := map[string]bool{}

	// Value for each form key is actually an array of strings. We're taking the first one here only
	// TODO: Tweak it
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
		c.renderStructItem(w, r, uri, c.uriStructNameFunc[uri][structName], id, postValues, MsgFailure, fmt.Sprintf("The following fields have invalid values: %s", strings.Join(invVals, ",")))
		return true
	}

	err2 := c.struct2db.Save(obj, struct2db.SaveOptions{})
	if err2 != nil {
		c.renderStructItem(w, r, uri, c.uriStructNameFunc[uri][structName], id, postValues, MsgFailure, fmt.Sprintf("Problem with saving: %s", err2.Unwrap().Error()))
		return true
	}

	// Update
	if id != "" {
		c.renderStructItem(w, r, uri, c.uriStructNameFunc[uri][structName], id, postValues, MsgSuccess, fmt.Sprintf("%s item has been successfully updated.", structName))
		return true
	}

	// Create
	c.renderMsg(w, r, MsgSuccess, fmt.Sprintf("%s item has been successfully added.", structName))
	return true
}
