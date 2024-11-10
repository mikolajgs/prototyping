package ui

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
	"log"

	stdb "github.com/mikolajgs/prototyping/pkg/struct-db-postgres"
	sthtml "github.com/mikolajgs/prototyping/pkg/struct-html"
	stsql "github.com/mikolajgs/prototyping/pkg/struct-sql-postgres"
	validator "github.com/mikolajgs/struct-validator"

	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type structItemTplObj struct {
	Name       string
	URI        string
	FieldsHTML string
	MsgHTML    string
	OnlyMsg    bool
	ID         string
}

func (c *Controller) tryGetStructItem(w http.ResponseWriter, r *http.Request, uri string) bool {
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

	// Render the page
	c.renderStructItem(w, r, uri, c.uriStructNameFunc[uri][structName], id, map[string]string{}, 0, "")

	return true
}

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
		err2 := c.struct2db.Delete(obj, stdb.DeleteOptions{})
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
		postValues[fk] = fv[0]

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
		for k := range invalidFormFields {
			failedFields[k] = failedFields[k] | validator.FailRegexp
		}
	}

	if !valid || len(failedFields) > 0 {
		invVals := []string{}
		for k := range failedFields {
			invVals = append(invVals, k)
		}
		c.renderStructItem(w, r, uri, c.uriStructNameFunc[uri][structName], id, postValues, MsgFailure, fmt.Sprintf("The following fields have invalid values: %s", strings.Join(invVals, ",")))
		return true
	}

	err2 := c.struct2db.Save(obj, stdb.SaveOptions{})
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

func (c *Controller) renderStructItem(w http.ResponseWriter, r *http.Request, uri string, objFunc func() interface{}, id string, postValues map[string]string, msgType int, msg string) {
	tpl, err := c.getStructItemHTML(uri, objFunc, id, postValues, msgType, msg)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
		return
	}
	w.Write([]byte(tpl))
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

func (c *Controller) getStructItemTplObj(uri string, objFunc func() interface{}, id string, postValues map[string]string, msgType int, msg string) (*structItemTplObj, error) {
	o := objFunc()

	if id != "" {
		err := c.struct2db.Load(o, id, stdb.LoadOptions{})
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
		URI:        uri,
		Name:       stsql.GetStructName(o),
		FieldsHTML: sthtml.GetFields(o, postValues, useFieldValues),
		MsgHTML:    c.getMsgHTML(msgType, msg),
		OnlyMsg:    onlyMsg,
		ID:         id,
	}

	return a, nil
}
