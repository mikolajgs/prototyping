package struct2sql

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

func (h *Struct2sql) setDefaultTags(src *Struct2sql) {
	if src != nil {
		h.defaultFieldsTags = make(map[string]map[string]string)
		h.defaultFieldsTags = src.getFieldsTags()
	}
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

	// Get table name
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
	vals := ""
	valsWithoutID := ""
	colsWithoutID := ""
	colVals := ""
	colValsAgain := ""
	idCol := h.dbColPrefix + "_id"

	valCnt := 0
	valWithoutIDCnt := 0
	for j := 0; j < s.NumField(); j++ {
		f := s.Field(j)
		k := f.Type.Kind()

		// Only basic golang types are included as columns for the database table.
		// Check the function below for the details.
		if !IsFieldKindSupported(k) {
			continue
		}

		dbCol := h.getDBCol(f.Name)
		h.dbFieldCols[f.Name] = dbCol
		h.dbCols[dbCol] = f.Name
		uniq := false
		if h.fieldsUniq[f.Name] {
			uniq = true
		}
		dbColParams := h.getDBColParams(f.Name, f.Type.String(), uniq)

		colsWithTypes = h.addWithComma(colsWithTypes, dbCol+" "+dbColParams)
		cols = h.addWithComma(cols, dbCol)

		// Assuming that primary field is named ID
		if f.Name != "ID" {
			colsWithoutID = h.addWithComma(colsWithoutID, dbCol)
			colVals = h.addWithComma(colVals, dbCol+"=?")
			valWithoutIDCnt++
		}

		valCnt++

		h.fields = append(h.fields, f.Name)
	}

	colValsAgain = colVals

	if valCnt > 0 {
		vals = "?"
		if valCnt > 1 {
			vals = vals + strings.Repeat(",?", valCnt-1)
		}
	}
	if valWithoutIDCnt > 0 {
		valsWithoutID = "?"
		if valWithoutIDCnt > 1 {
			valsWithoutID = valsWithoutID + strings.Repeat(",?", valWithoutIDCnt-1)
		}
	}
	if valCnt > 0 {
		for i:=1; i<=valCnt; i++ {
			vals = strings.Replace(vals, "?", fmt.Sprintf("$%d", i), 1)
			valsWithoutID = strings.Replace(valsWithoutID, "?", fmt.Sprintf("$%d", i), 1)
			colVals = strings.Replace(colVals, "?", fmt.Sprintf("$%d", i), 1)
		}
		dollarCnt := strings.Count(vals, "$")
		for i:=dollarCnt+1; i<=dollarCnt+valCnt; i++ {
			colValsAgain = strings.Replace(colValsAgain, "?", fmt.Sprintf("$%d", i), 1)
		}
	}

	// Full SQL queries or their prefixes. Query parts such as columns and values in UPDATE or conditions after WHERE etc. must be generated on the fly and cannot be cached.
	h.queryDropTable = fmt.Sprintf("DROP TABLE IF EXISTS %s", h.dbTbl)
	h.queryCreateTable = fmt.Sprintf("CREATE TABLE %s (%s)", h.dbTbl, colsWithTypes)
	h.queryDeleteById = fmt.Sprintf("DELETE FROM %s WHERE %s = $1", h.dbTbl, idCol)
	h.querySelectById = fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", cols, h.dbTbl, idCol)
	h.queryInsert = fmt.Sprintf("INSERT INTO %s(%s) VALUES (%s) RETURNING %s", h.dbTbl, colsWithoutID, valsWithoutID, idCol)
	h.queryUpdateById = fmt.Sprintf("UPDATE %s SET %s WHERE %s = $%d", h.dbTbl, colVals, idCol, valCnt)
	h.queryInsertOnConflictUpdate = fmt.Sprintf("INSERT INTO %s(%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s RETURNING %s", h.dbTbl, cols, vals, idCol, colValsAgain, idCol)
	h.querySelectPrefix = fmt.Sprintf("SELECT %s FROM %s", cols, h.dbTbl)
	h.querySelectCountPrefix = fmt.Sprintf("SELECT COUNT(*) AS cnt FROM %s", h.dbTbl)
	h.queryDeletePrefix = fmt.Sprintf("DELETE FROM %s", h.dbTbl)
	h.queryUpdatePrefix = fmt.Sprintf("UPDATE %s SET", h.dbTbl)
}

func (h *Struct2sql) reflectStructForValidation(u interface{}) {
	v := reflect.ValueOf(u)
	i := reflect.Indirect(v)
	s := i.Type()

	h.fieldsDefaultValue = make(map[string]string)
	h.fieldsUniq = make(map[string]bool)
	h.fieldsTags = make(map[string]map[string]string)
	h.fieldsOverwriteType = make(map[string]string)

	for j := 0; j < s.NumField(); j++ {
		f := s.Field(j)
		k := f.Type.Kind()

		// Only basic golang types are included as columns for the database table.
		// Check the function below for the details.
		if !IsFieldKindSupported(k) {
			continue
		}

		crudTag := f.Tag.Get(h.tagName)
		crudValTag := f.Tag.Get(h.tagName+"_val")
		if h.defaultFieldsTags != nil {
			if crudTag == "" && h.defaultFieldsTags[f.Name][h.tagName] != "" {
				crudTag = h.defaultFieldsTags[f.Name][h.tagName]
			}
			if crudValTag == "" && h.defaultFieldsTags[f.Name][h.tagName+"_val"] != "" {
				crudValTag = h.defaultFieldsTags[f.Name][h.tagName+"_val"]
			}
		}

		h.setFieldFromTag(crudTag, f.Name)
		if h.err != nil {
			return
		}

		if crudValTag != "" {
			h.fieldsDefaultValue[f.Name] = crudValTag
		}

		h.fieldsTags[f.Name] = make(map[string]string)
		h.fieldsTags[f.Name][h.tagName] = f.Tag.Get(h.tagName)
		h.fieldsTags[f.Name][h.tagName+"_val"] = f.Tag.Get(h.tagName+"_val")
	}
}

func (h *Struct2sql) setFieldFromTag(tag string, fieldName string) {
	opts := strings.SplitN(tag, " ", -1)
	for _, opt := range opts {
		h.setFieldFromTagOptWithoutVal(opt, fieldName)
	}
}

func (h *Struct2sql) setFieldFromTagOptWithoutVal(opt string, fieldName string) {
	if opt == "uniq" {
		h.fieldsUniq[fieldName] = true
		return
	}
	if strings.HasPrefix(opt, "db_type:") {
		dbTypeArr := strings.Split(opt, ":")
		typeUpperCase := strings.ToUpper(dbTypeArr[1])
		if typeUpperCase == "TEXT" || typeUpperCase == "BPCHAR" {
			h.fieldsOverwriteType[fieldName] = typeUpperCase
			return
		}
		m, _ := regexp.MatchString(`^(VARCHAR|CHARACTER VARYING|BPCHAR|CHAR|CHARACTER)\([0-9]+\)$`, typeUpperCase)
		if m {
			h.fieldsOverwriteType[fieldName] = typeUpperCase
			return
		}
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

// Mapping database column type to struct field type
func (h *Struct2sql) getDBColParams(n string, t string, uniq bool) string {
	dbColParams := ""
	if n == "ID" {
		dbColParams = "SERIAL PRIMARY KEY"
	} else if n == "Flags" {
		dbColParams = "BIGINT DEFAULT 0"
		// String types can be overwritten by a tag
	} else if h.fieldsOverwriteType[n] != "" {
		dbColParams = h.fieldsOverwriteType[n]+" DEFAULT ''"
	} else {
		switch t {
		case "string":
			dbColParams = "VARCHAR(255) DEFAULT ''"
		case "bool":
			dbColParams = "BOOLEAN DEFAULT false"
		case "int64":
			dbColParams = "BIGINT DEFAULT 0"
		case "int32":
			dbColParams = "INTEGER DEFAULT 0"
		case "int16":
			dbColParams = "SMALLINT DEFAULT 0"
		case "int8":
			dbColParams = "SMALLINT DEFAULT 0"
		case "int":
			dbColParams = "BIGINT DEFAULT 0"
		case "uint64":
			dbColParams = "BIGINT DEFAULT 0"
		case "uint32":
			dbColParams = "INTEGER DEFAULT 0"
		case "uint16":
			dbColParams = "SMALLINT DEFAULT 0"
		case "uint8":
			dbColParams = "SMALLINT DEFAULT 0"
		case "uint":
			dbColParams = "BIGINT DEFAULT 0"
		// TODO: Consider something different
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

func (h *Struct2sql) getQuerySet(values map[string]interface{}, valueFieldsToInclude map[string]bool) (string, int) {
	qSet := ""
	// Variable number in the query, the '$x'
	i := 1

	if len(values) == 0 {
		return "", 0
	}

	sorted := []string{}
	for k := range values {
		if h.dbFieldCols[k] == "" {
			continue
		}
		if len(valueFieldsToInclude) > 0 && !valueFieldsToInclude[k] {
			continue
		}
		sorted = append(sorted, h.dbFieldCols[k])
	}

	if len(sorted) > 0 {
		sort.Strings(sorted)

		for _, col := range sorted {
			qSet = h.addWithComma(qSet, fmt.Sprintf(col+"=$%d", i))
			i++
		}
	}	

	return qSet, i-1
}

func (h *Struct2sql) getQueryFilters(filters map[string]interface{}, filterFieldsToInclude map[string]bool, firstNumber int) string {
	qWhere := ""
	// Variable number in the query, the '$x'
	i := firstNumber

	if len(filters) == 0 {
		return ""
	}

	sorted := []string{}
	for k := range filters {
		if h.dbFieldCols[k] == "" {
			continue
		}
		if len(filterFieldsToInclude) > 0 && !filterFieldsToInclude[k] {
			continue
		}
		// _raw is a special entry that allows almost-raw SQL query
		if k == "_raw" {
			continue
		}
		sorted = append(sorted, h.dbFieldCols[k])
	}

	if len(sorted) > 0 {
		sort.Strings(sorted)

		for _, col := range sorted {
			qWhere = h.addWithAnd(qWhere, fmt.Sprintf(col+"=$%d", i))
			i++
		}
	}

	rawQueryArr, ok := filters["_raw"]
	if !ok || len(rawQueryArr.([]interface{})) == 0 {
		return qWhere
	}

	rawQuery := filters["_raw"].([]interface{})[0].(string)
	if rawQuery == "" {
		return qWhere
	}

	if rawQuery != "" {
		if qWhere != "" {
			qWhere = fmt.Sprintf("(%s)", qWhere)
			conjunction, ok := filters["_rawConjuction"].(int)
			if !ok || conjunction != RawConjuctionOR {
				qWhere += " AND "
			} else {
				qWhere += " OR "
			}
		}
		qWhere += "("

		reField := regexp.MustCompile(`\.[a-zA-Z0-9]+`)
		foundFields := reField.FindAllString(rawQuery, -1)
		alreadyReplaced := map[string]bool{}
		for _, f := range foundFields {
			if !alreadyReplaced[f] {
				fieldName := strings.Replace(f, ".", "", 1)

				// If field does not exist, it won't be processed
				if h.dbFieldCols[fieldName] == "" {
					continue
				}

				rawQuery = strings.ReplaceAll(rawQuery, f, h.dbFieldCols[fieldName])
				alreadyReplaced[f] = true
			}
		}

		for j := 1; j < len(filters["_raw"].([]interface{})); j++ {
			rt := reflect.TypeOf(filters["_raw"].([]interface{})[j])
			if rt.Kind() == reflect.Slice || rt.Kind() == reflect.Array {
				queryVal := ""
				for k := 0; k < reflect.ValueOf(filters["_raw"].([]interface{})[j]).Len(); k++ {
					if k == 0 {
						queryVal += fmt.Sprintf("$%d", i)
						i++
						continue
					}
					queryVal += fmt.Sprintf(",$%d", i)
					i++
				}
				rawQuery = strings.Replace(rawQuery, "?", queryVal, 1)
				continue
			}


			// Value is a single value so just replace ? with $x, eg $2
			rawQuery = strings.Replace(rawQuery, "?", fmt.Sprintf("$%d", i), 1)
			i++
		}

		qWhere += rawQuery + ")"
	}

	return qWhere
}
