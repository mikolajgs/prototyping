package ui

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	struct2db "github.com/mikolajgs/prototyping/pkg/struct-db-postgres"
)

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
