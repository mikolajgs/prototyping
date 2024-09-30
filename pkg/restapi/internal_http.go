package restapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/mikolajgs/crud/pkg/struct2db"
)

func (c Controller) handleHTTPPut(w http.ResponseWriter, r *http.Request, newObjFunc func() interface{}, id string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		c.writeErrText(w, http.StatusInternalServerError, "cannot_read_request_body")
		return
	}

	objClone := newObjFunc()

	if id != "" {
		err2 := c.struct2db.Load(objClone, id)
		if err2 != nil {
			c.writeErrText(w, http.StatusInternalServerError, "cannot_get_from_db")
			return
		}
		if c.struct2db.GetObjIDValue(objClone) == 0 {
			c.writeErrText(w, http.StatusNotFound, "not_found_in_db")
			return
		}
	} else {
		c.struct2db.ResetFields(objClone)
	}

	err = json.Unmarshal(body, objClone)
	if err != nil {
		c.writeErrText(w, http.StatusBadRequest, "invalid_json")
		return
	}

	b, _, err := c.Validate(objClone, nil)
	if !b || err != nil {
		c.writeErrText(w, http.StatusBadRequest, "validation_failed")
		return
	}

	err2 := c.struct2db.Save(objClone, struct2db.SaveOptions{})
	if err2 != nil {
		c.writeErrText(w, http.StatusInternalServerError, "cannot_save_to_db")
		return
	}

	if id != "" {
		c.writeOK(w, http.StatusOK, map[string]interface{}{
			"id": c.struct2db.GetObjIDValue(objClone),
		})
	} else {
		c.writeOK(w, http.StatusCreated, map[string]interface{}{
			"id": c.struct2db.GetObjIDValue(objClone),
		})
	}
}

func (c Controller) handleHTTPGet(w http.ResponseWriter, r *http.Request, newObjFunc func() interface{}, id string) {
	if id == "" {
		obj := newObjFunc()
		params := c.getParamsFromURI(r.RequestURI)
		limit, _ := strconv.Atoi(params["limit"])
		offset, _ := strconv.Atoi(params["offset"])
		if limit < 1 {
			limit = 10
		}
		if offset < 0 {
			offset = 0
		}

		order := []string{}
		if params["order"] != "" {
			order = append(order, params["order"])
			order = append(order, params["order_direction"])
		}

		filters := make(map[string]interface{})
		for k, v := range params {
			if strings.HasPrefix(k, "filter_") {
				k = k[7:]
				fieldName, fieldValue, errF := c.uriFilterToFilter(obj, k, v)
				if errF != nil {
					if errF.Op == "GetHelper" {
						c.writeErrText(w, http.StatusInternalServerError, "get_helper")
						return
					} else {
						c.writeErrText(w, http.StatusBadRequest, "invalid_filter")
						return
					}
				}
				if fieldName != "" {
					filters[fieldName] = fieldValue
				}
			}
		}

		xobj, err1 := c.struct2db.Get(newObjFunc, struct2db.GetOptions{
			Order: order,
			Limit: limit,
			Offset: offset,
			Filters: filters,
		})
		if err1 != nil {
			if err1.Op == "ValidateFilters" {
				c.writeErrText(w, http.StatusBadRequest, "invalid_filter_value")
				return
			} else {
				c.writeErrText(w, http.StatusInternalServerError, "cannot_get_from_db")
				return
			}
		}

		c.writeOK(w, http.StatusOK, map[string]interface{}{
			"items": xobj,
		})

		return
	}

	objClone := newObjFunc()

	err := c.struct2db.Load(objClone, id)
	if err != nil {
		c.writeErrText(w, http.StatusInternalServerError, "cannot_get_from_db")
		return
	}

	if c.struct2db.GetObjIDValue(objClone) == 0 {
		c.writeErrText(w, http.StatusNotFound, "not_found_in_db")
		return
	}

	c.writeOK(w, http.StatusOK, map[string]interface{}{
		"item": objClone,
	})
}

func (c Controller) handleHTTPDelete(w http.ResponseWriter, r *http.Request, newObjFunc func() interface{}, id string) {
	if id == "" {
		c.writeErrText(w, http.StatusBadRequest, "invalid_id")
		return
	}

	objClone := newObjFunc()

	err := c.struct2db.Load(objClone, id)
	if err != nil {
		c.writeErrText(w, http.StatusInternalServerError, "cannot_get_from_db")
		return
	}
	if c.struct2db.GetObjIDValue(objClone) == 0 {
		c.writeErrText(w, http.StatusNotFound, "not_found_in_db")
		return
	}

	err = c.struct2db.Delete(objClone, struct2db.DeleteOptions{})
	if err != nil {
		c.writeErrText(w, http.StatusInternalServerError, "cannot_delete_from_db")
		return
	}

	c.writeOK(w, http.StatusOK, map[string]interface{}{
		"id": id,
	})
}

func (c Controller) getIDFromURI(uri string, w http.ResponseWriter) (string, bool) {
	xs := strings.SplitN(uri, "?", 2)
	if xs[0] == "" {
		return "", true
	}
	matched, err := regexp.Match(`^[0-9]+$`, []byte(xs[0]))
	if err != nil || !matched {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(c.jsonError("invalid id"))
		return "", false
	}
	return xs[0], true
}

func (c Controller) getParamsFromURI(uri string) map[string]string {
	o := make(map[string]string)
	xs := strings.SplitN(uri, "?", 2)
	if len(xs) < 2 || xs[1] == "" {
		return o
	}
	xp := strings.SplitN(xs[1], "&", -1)
	for _, p := range xp {
		pv := strings.SplitN(p, "=", 2)
		matched, err := regexp.Match(`^[0-9a-zA-Z_]+$`, []byte(pv[0]))
		if len(pv) == 1 || err != nil || !matched {
			continue
		}
		unesc, err := url.QueryUnescape(pv[1])
		if err != nil {
			continue
		}
		o[pv[0]] = unesc
	}
	return o
}

func (c Controller) jsonError(e string) []byte {
	return []byte(fmt.Sprintf("{\"err\":\"%s\"}", e))
}

func (c Controller) jsonID(id int64) []byte {
	return []byte(fmt.Sprintf("{\"id\":\"%d\"}", id))
}

func (c Controller) uriFilterToFilter(obj interface{}, filterName string, filterValue string) (string, interface{}, *ErrController) {
	fieldName, cErr := c.struct2db.GetFieldNameFromDBCol(obj, filterName)
	if cErr != nil {
		return "", nil, &ErrController{
			Op:  "GetDBCol",
			Err: fmt.Errorf("Error getting field name from filter: %w", cErr.Unwrap()),
		}
	}

	if fieldName == "" {
		return "", nil, nil
	}

	val := reflect.ValueOf(obj).Elem()
	valueField := val.FieldByName(fieldName)
	if valueField.Type().Name() == "int" {
		filterInt, err := strconv.Atoi(filterValue)
		if err != nil {
			return "", nil, &ErrController{
				Op:  "InvalidValue",
				Err: fmt.Errorf("Error converting string to int: %w", err),
			}
		}
		return fieldName, filterInt, nil
	}
	if valueField.Type().Name() == "int64" {
		filterInt64, err := strconv.ParseInt(filterValue, 10, 64)
		if err != nil {
			return "", nil, &ErrController{
				Op:  "InvalidValue",
				Err: fmt.Errorf("Error converting string to int64: %w", err),
			}
		}
		return fieldName, filterInt64, nil
	}
	if valueField.Type().Name() == "string" {
		return fieldName, filterValue, nil
	}

	return "", nil, nil
}

func (c Controller) writeErrText(w http.ResponseWriter, status int, errText string) {
	r := NewHTTPResponse(0, errText)
	j, err := json.Marshal(r)
	w.WriteHeader(status)
	if err == nil {
		w.Write(j)
	}
}

func (c Controller) writeOK(w http.ResponseWriter, status int, data map[string]interface{}) {
	r := NewHTTPResponse(1, "")
	r.Data = data
	j, err := json.Marshal(r)
	w.WriteHeader(status)
	if err == nil {
		w.Write(j)
	}
}
