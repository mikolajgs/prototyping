package ui

import (
	"bytes"
	"embed"
	"fmt"
	"html"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"regexp"
	"math"

	struct2db "github.com/mikolajgs/prototyping/pkg/struct-db-postgres"
	stsql "github.com/mikolajgs/prototyping/pkg/struct-sql-postgres"
)

type structItemsTplObj struct {
	Name        string
	URI         string
	Fields      []string
	ItemsHTML   []interface{}
	ItemsCount  int64
	PageNumbers []string
	ParamPage   int
	ParamLimit  int
}

type structItemsParams struct {
	Page     int
	Limit    int
	RawQuery string
	Filters  map[string]string
}

func (c *Controller) tryGetStructItems(w http.ResponseWriter, r *http.Request, uri string) bool {
	realURI := c.getRealURI(uri, r.RequestURI)
	if strings.HasPrefix(realURI, "x/struct_items/") {
		structName := strings.Replace(realURI, "x/struct_items/", "", 1)
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

		c.renderStructItems(w, r, uri, c.uriStructNameFunc[uri][structName], structItemsParams{
			Page: 1,
			Limit: 25,
		})
		return true
	}
	return false
}

func (c *Controller) tryStructItems(w http.ResponseWriter, r *http.Request, uri string) bool {
	structName, id := c.getStructAndIDFromURI("x/struct_items/", c.getRealURI(uri, r.RequestURI))

	if structName == "" || id != "" {
		return false
	}

	// Check if struct exists
	newObjFunc, ok := c.uriStructNameFunc[uri][structName]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return true
	}

	if r.Method != http.MethodPut && r.Method != http.MethodPost && r.Method != http.MethodDelete {
		return false
	}

	_ = c.uriStructNameFunc[uri][structName]()

	// Handle delete here
	if r.Method == http.MethodDelete {
		// Read ids from the param 'ids'
		ids := r.URL.Query().Get("ids")
		match, err := regexp.MatchString(`^[0-9]+(,[0-9]+)*$`, ids)
		if err != nil || !match {
			w.WriteHeader(http.StatusBadRequest)
			return true
		}

		// Run delete for multiple rows
		idsList := strings.Split(ids, ",")
		idsInt := []int64{}
		for _, id := range idsList {
			idInt, _ := strconv.ParseInt(id, 10, 64)
			idsInt = append(idsInt, idInt)
		}

		err2 := c.struct2db.DeleteMultiple(newObjFunc, struct2db.DeleteMultipleOptions{
			Filters: map[string]interface{}{
				"_raw": []interface{}{
					".ID IN (?)",
					idsInt,
				},
			},
		})

		if err2 != nil {
			c.renderMsg(w, r, MsgFailure, fmt.Sprintf("Problem with removing %s items.", structName))
			return true
		}

		c.renderMsg(w, r, MsgSuccess, fmt.Sprintf("%s items have been successfully deleted.", structName))
		return true
	}

	// Any other request is invalid
	w.WriteHeader(http.StatusBadRequest)
	return true
}

func (c *Controller) renderStructItems(w http.ResponseWriter, r *http.Request, uri string, objFunc func() interface{}, params structItemsParams) {
	tpl, err := c.getStructItemsHTML(uri, objFunc, params)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
		return
	}
	w.Write([]byte(tpl))
}

func (c *Controller) getStructItemsHTML(uri string, objFunc func() interface{}, params structItemsParams) (string, error) {
	structItemsTpl, err := embed.FS.ReadFile(htmlDir, "html/struct_items.html")
	if err != nil {
		return "", fmt.Errorf("error reading struct items template from embed: %w", err)
	}

	tplObj, err := c.getStructItemsTplObj(uri, objFunc, params)
	if err != nil {
		return "", fmt.Errorf("error getting struct items for html: %w", err)
	}

	buf := &bytes.Buffer{}
	t := template.Must(template.New("structItems").Funcs(template.FuncMap{
		"SplitRow": func(s string) []string {
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

func (c *Controller) getStructItemsTplObj(uri string, objFunc func() interface{}, params structItemsParams) (*structItemsTplObj, error) {
	o := objFunc()

	getPage, getLimit, getOffset := c.getPageLimitOffset(params.Page, params.Limit)

	itemsHTML, err := c.struct2db.Get(objFunc, struct2db.GetOptions{
		Offset: getOffset,
		Limit:  getLimit,
		RowObjTransformFunc: func(obj interface{}) interface{} {
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

	itemsCount, err := c.struct2db.GetCount(objFunc, struct2db.GetCountOptions{})
	if err != nil {
		return nil, err
	}
	pageNumbers := c.getPageNumbers(itemsCount, getLimit, getPage)

	its := &structItemsTplObj{
		URI:         uri,
		Name:        stsql.GetStructName(o),
		Fields:      stsql.GetStructFieldNames(o),
		ItemsHTML:   itemsHTML,
		ItemsCount:  itemsCount,
		ParamPage:   getPage,
		ParamLimit:  getLimit,
		PageNumbers: pageNumbers,
	}

	return its, nil
}

func (c *Controller) getPageLimitOffset(page int, limit int) (int, int, int) {
	if limit < 1 {
		limit = 25
	}
	if page < 1 {
		page = 1
	}
	return page, limit, limit * (page - 1)
}

func (c *Controller) getPageNumbers(itemsCount int64, getLimit int, getPage int) []string {
	p := int(math.Ceil(float64(itemsCount) / float64(getLimit)))

	a := []string{}
	if p < 11 {
		for i:=1; i<=p; i++ {
			a = append(a, fmt.Sprintf("%d", i))
		}
		return a
	}
	a = append(a, "1")
	if getPage-2 > 1 {
		a = append(a, "")
	}
	if getPage-1 > 1 {
		a = append(a, fmt.Sprintf("%d", getPage-1))
	}
	if getPage > 1 && getPage < p {
		a = append(a, fmt.Sprintf("%d", getPage))
	}
	if getPage+1 < p {
		a = append(a, fmt.Sprintf("%d", getPage+1))
	}
	if getPage+2 < p {
		a = append(a, "")
	}
	a = append(a, fmt.Sprintf("%d", p))
	log.Printf("a: %v", a)
	return a
} 