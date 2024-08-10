package struct2sql

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

func (h *Struct2sql) setDefaultTags(src *Struct2sql) {
	if src != nil {
		h.defaultFieldsTags = make(map[string]map[string]string)
		h.defaultFieldsTags = src.getFieldsTags()
	}
}

// TODO Is it used?
func (h *Struct2sql) getColsCommaSeparated(fields []string) (string, int, string, string) {
	cols := ""
	colCnt := 0
	vals := ""
	colVals := ""
	for _, k := range fields {
		if h.dbFieldCols[k] == "" {
			continue
		}
		cols = h.addWithComma(cols, h.dbFieldCols[k])
		colCnt++
		vals = h.addWithComma(vals, "$"+strconv.Itoa(colCnt))
		colVals = h.addWithComma(colVals, h.dbFieldCols[k]+"=$"+strconv.Itoa(colCnt))
	}
	return cols, colCnt, vals, colVals
}

func (h *Struct2sql) getFieldsTags() map[string]map[string]string {
	return h.fieldsTags
}

func (h *Struct2sql) reflectStruct(u interface{}, dbTablePrefix string, forceName string) {
	h.reflectStructForValidation(u)
	h.reflectStructForDBQueries(u, dbTablePrefix, forceName)
}

func (h *Struct2sql) reflectStructForDBQueries(u interface{}, dbTablePrefix string, forceName string) {
	v := reflect.ValueOf(u)
	i := reflect.Indirect(v)
	s := i.Type()

	usName := h.getUnderscoredName(s.Name())
	if forceName != "" {
		usName = h.getUnderscoredName(forceName)
	}
	usPluName := h.getPluralName(usName)
	h.dbTbl = dbTablePrefix + usPluName
	h.dbColPrefix = usName
	h.url = usPluName

	h.dbFieldCols = make(map[string]string)
	h.dbCols = make(map[string]string)

	colsWithTypes := ""
	cols := ""
	valsWithoutID := ""
	colsWithoutID := ""
	colVals := ""
	idCol := h.dbColPrefix + "_id"

	valCnt := 1
	for j := 0; j < s.NumField(); j++ {
		field := s.Field(j)
		if field.Type.Kind() != reflect.Int64 && field.Type.Kind() != reflect.String && field.Type.Kind() != reflect.Int {
			continue
		}

		if field.Type.Kind() == reflect.Int64 {
			h.fieldsFlags[field.Name] += TypeInt64
		}
		if field.Type.Kind() == reflect.Int {
			h.fieldsFlags[field.Name] += TypeInt
		}
		if field.Type.Kind() == reflect.String {
			h.fieldsFlags[field.Name] += TypeString
		}

		dbCol := h.getDBCol(field.Name)
		h.dbFieldCols[field.Name] = dbCol
		h.dbCols[dbCol] = field.Name
		uniq := false
		if h.fieldsUniq[field.Name] {
			uniq = true
		}
		dbColParams := h.getDBColParams(field.Name, field.Type.String(), uniq)

		colsWithTypes = h.addWithComma(colsWithTypes, dbCol+" "+dbColParams)
		cols = h.addWithComma(cols, dbCol)

		if field.Name != "ID" {
			colsWithoutID = h.addWithComma(colsWithoutID, dbCol)
			valsWithoutID = h.addWithComma(valsWithoutID, "$"+strconv.Itoa(valCnt))
			colVals = h.addWithComma(colVals, dbCol+"=$"+strconv.Itoa(valCnt))
			valCnt++
		}

		h.fields = append(h.fields, field.Name)
	}

	h.queryDropTable = fmt.Sprintf("DROP TABLE IF EXISTS %s", h.dbTbl)
	h.queryCreateTable = fmt.Sprintf("CREATE TABLE %s (%s)", h.dbTbl, colsWithTypes)
	h.queryDeleteById = fmt.Sprintf("DELETE FROM %s WHERE %s = $1", h.dbTbl, idCol)
	h.querySelectById = fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", cols, h.dbTbl, idCol)
	h.queryInsert = fmt.Sprintf("INSERT INTO %s(%s) VALUES (%s) RETURNING %s", h.dbTbl, colsWithoutID, valsWithoutID, idCol)
	h.queryUpdateById = fmt.Sprintf("UPDATE %s SET %s WHERE %s = $%d", h.dbTbl, colVals, idCol, valCnt)
	h.querySelectPrefix = fmt.Sprintf("SELECT %s FROM %s", cols, h.dbTbl)
	h.querySelectCountPrefix = fmt.Sprintf("SELECT COUNT(*) AS cnt FROM %s", h.dbTbl)
	h.queryDeletePrefix = fmt.Sprintf("DELETE FROM %s", h.dbTbl)
}

func (h *Struct2sql) reflectStructForValidation(u interface{}) {
	v := reflect.ValueOf(u)
	i := reflect.Indirect(v)
	s := i.Type()

	h.fieldsFlags = make(map[string]int)
	h.fieldsDefaultValue = make(map[string]string)
	h.fieldsUniq = make(map[string]bool)
	h.fieldsTags = make(map[string]map[string]string)

	for j := 0; j < s.NumField(); j++ {
		field := s.Field(j)
		if field.Type.Kind() != reflect.Int64 && field.Type.Kind() != reflect.String && field.Type.Kind() != reflect.Int {
			continue
		}

		crudTag := field.Tag.Get("crud")
		crudValTag := field.Tag.Get("crud_val")
		if h.defaultFieldsTags != nil {
			if crudTag == "" && h.defaultFieldsTags[field.Name]["crud"] != "" {
				crudTag = h.defaultFieldsTags[field.Name]["crud"]
			}
			if crudValTag == "" && h.defaultFieldsTags[field.Name]["crud_val"] != "" {
				crudValTag = h.defaultFieldsTags[field.Name]["crud_val"]
			}
		}

		h.setFieldFromTag(crudTag, j, field.Name)
		if h.err != nil {
			return
		}

		if crudValTag != "" {
			h.fieldsDefaultValue[field.Name] = crudValTag
		}

		h.fieldsTags[field.Name] = make(map[string]string)
		h.fieldsTags[field.Name]["crud"] = field.Tag.Get("crud")
		h.fieldsTags[field.Name]["crud_val"] = field.Tag.Get("crud_val")
	}
}

func (h *Struct2sql) setFieldFromTag(tag string, fieldIdx int, fieldName string) {
	var ErrStruct2sql *ErrStruct2sql
	opts := strings.SplitN(tag, " ", -1)
	for _, opt := range opts {
		if ErrStruct2sql != nil {
			break
		}
		h.setFieldFromTagOptWithoutVal(opt, fieldIdx, fieldName)
	}
}

func (h *Struct2sql) setFieldFromTagOptWithoutVal(opt string, fieldIdx int, fieldName string) {
	if opt == "uniq" {
		h.fieldsUniq[fieldName] = true
	}
}

func (h *Struct2sql) getDBCol(n string) string {
	dbCol := ""
	if n == "ID" {
		dbCol = h.dbColPrefix + "_id"
	} else if n == "Flags" {
		dbCol = h.dbColPrefix + "_flags"
	} else {
		dbCol = h.getUnderscoredName(n)
	}
	return dbCol
}

func (h *Struct2sql) getDBColParams(n string, t string, uniq bool) string {
	dbColParams := ""
	if n == "ID" {
		dbColParams = "SERIAL PRIMARY KEY"
	} else if n == "Flags" {
		dbColParams = "BIGINT DEFAULT 0"
	} else {
		switch t {
		case "string":
			dbColParams = "VARCHAR(255) DEFAULT ''"
		case "int64":
			dbColParams = "BIGINT DEFAULT 0"
		case "int":
			dbColParams = "BIGINT DEFAULT 0"
		default:
			dbColParams = "VARCHAR(255) DEFAULT ''"
		}
	}
	if uniq {
		dbColParams += " UNIQUE"
	}
	return dbColParams
}

func (h *Struct2sql) addWithComma(s string, v string) string {
	if s != "" {
		s += ","
	}
	s += v
	return s
}

func (h *Struct2sql) addWithAnd(s string, v string) string {
	if s != "" {
		s += " AND "
	}
	s += v
	return s
}

func (h *Struct2sql) getUnderscoredName(s string) string {
	o := ""

	var prev rune
	for i, ch := range s {
		if i == 0 {
			o += strings.ToLower(string(ch))
		} else {
			if unicode.IsUpper(ch) {
				if prev == 'I' && ch == 'D' {
					o += strings.ToLower(string(ch))
				} else {
					o += "_" + strings.ToLower(string(ch))
				}
			} else {
				o += string(ch)
			}
		}
		prev = ch
	}
	return o
}

func (h *Struct2sql) getPluralName(s string) string {
	re := regexp.MustCompile(`y$`)
	if re.MatchString(s) {
		return string(re.ReplaceAll([]byte(s), []byte(`ies`)))
	}
	re = regexp.MustCompile(`s$`)
	if re.MatchString(s) {
		return s + "es"
	}
	return s + "s"
}

func (h *Struct2sql) getQueryOrder(order []string, orderFieldsToInclude map[string]bool) string {
	qOrder := ""
	if len(order) > 0 {
		for i := 0; i < len(order); i = i + 2 {
			k := order[i]
			v := order[i+1]

			if len(orderFieldsToInclude) > 0 && !orderFieldsToInclude[k] && !orderFieldsToInclude[h.dbCols[k]] {
				continue
			}

			if h.dbFieldCols[k] == "" && h.dbCols[k] == "" {
				continue
			}

			d := "ASC"
			if v == strings.ToLower("desc") {
				d = "DESC"
			}
			if h.dbFieldCols[k] != "" {
				qOrder = h.addWithComma(qOrder, h.dbFieldCols[k]+" "+d)
			} else {
				qOrder = h.addWithComma(qOrder, k+" "+d)
			}
		}
	}
	return qOrder
}

func (h *Struct2sql) getQueryLimitOffset(limit int, offset int) string {
	qLimitOffset := ""
	if limit > 0 {
		if offset > 0 {
			qLimitOffset = fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
		} else {
			qLimitOffset = fmt.Sprintf("LIMIT %d", limit)
		}
	}
	return qLimitOffset
}

func (h *Struct2sql) getQueryFilters(filters map[string]interface{}, filterFieldsToInclude map[string]bool) string {
	qWhere := ""
	i := 1
	if len(filters) > 0 {
		sorted := []string{}
		for k := range filters {
			if h.dbFieldCols[k] == "" {
				continue
			}
			if len(filterFieldsToInclude) > 0 && !filterFieldsToInclude[k] {
				continue
			}
			sorted = append(sorted, h.dbFieldCols[k])
		}
		sort.Strings(sorted)
		for _, col := range sorted {
			qWhere = h.addWithAnd(qWhere, fmt.Sprintf(col+"=$%d", i))
			i++
		}
	}

	return qWhere
}
