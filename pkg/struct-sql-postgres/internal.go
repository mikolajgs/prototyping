package structsqlpostgres

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

func (h *StructSQL) setDefaultTags(src *StructSQL) {
	if src != nil {
		h.defaultFieldsTags = make(map[string]map[string]string)
		h.defaultFieldsTags = src.getFieldsTags()
	}
}

func (h *StructSQL) setDependenciesTags() {
	h.dependenciesFieldsTags = make(map[string]map[string]map[string]string)
	for k, v := range h.dependencies {
		h.dependenciesFieldsTags[k] = v.getFieldsTags()
	}
}

func (h *StructSQL) getFieldsTags() map[string]map[string]string {
	return h.fieldsTags
}

func (h *StructSQL) reflectStruct(u interface{}, dbTablePrefix string, forceName string) {
	h.reflectStructForValidation(u)
	h.reflectStructForDBQueries(u, dbTablePrefix, forceName)
}

func (h *StructSQL) reflectStructForDBQueries(u interface{}, dbTablePrefix string, forceName string) {
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

	var colsWithTypes, cols, vals, valsWithoutID, colsWithoutID, colVals, colValsAgain string
	idCol := h.dbColPrefix + "_id"
	if h.hasDependencies {
		idCol = fmt.Sprintf("t1.%s", idCol)
	}

	reDep := regexp.MustCompile(`^[a-zA-Z0-9]+_[a-zA-Z0-9]+`)
	innerJoins := ""
	joinedTables := map[string]string{}

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

		// Process field named 'xx_yy' which could be a joined struct field
		if h.hasDependencies && reDep.MatchString(f.Name) {
			fieldNameArr := strings.Split(f.Name, "_")

			_, ok := h.dependencies[fieldNameArr[0]]
			if ok {
				col := h.dependencies[fieldNameArr[0]].dbFieldCols[fieldNameArr[1]]
				tbl := h.dependencies[fieldNameArr[0]].dbTbl

				if _, ok2 := joinedTables[fieldNameArr[0]]; !ok2 {
					alias := fmt.Sprintf("t%d", len(joinedTables)+2)
					innerJoins += fmt.Sprintf(
						" INNER JOIN %s %s ON %s=%s.%s", 
						tbl, alias,
						h.dbFieldCols[fieldNameArr[0]+"ID"],
						alias, h.dependencies[fieldNameArr[0]].dbFieldCols["ID"],
					)
					joinedTables[fieldNameArr[0]] = alias
				}

				dbCol := fmt.Sprintf("%s.%s", joinedTables[fieldNameArr[0]], col)

				h.dbFieldCols[f.Name] = dbCol
				h.dbCols[dbCol] = f.Name
				
				cols = h.addWithComma(cols, dbCol)
				colsWithoutID = h.addWithComma(colsWithoutID, dbCol) // not used for now
				valWithoutIDCnt++
				valCnt++

				h.fields = append(h.fields, f.Name)

				continue
			}
		}

		// Continue when field does not come from joined struct
		dbCol := h.getDBCol(f.Name)
		if h.hasDependencies {
			dbCol = "t1."+dbCol
		}
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
	h.queryInsert = fmt.Sprintf("INSERT INTO %s(%s) VALUES (%s) RETURNING %s", h.dbTbl, colsWithoutID, valsWithoutID, idCol)
	h.queryUpdateById = fmt.Sprintf("UPDATE %s SET %s WHERE %s = $%d", h.dbTbl, colVals, idCol, valCnt)
	h.queryInsertOnConflictUpdate = fmt.Sprintf("INSERT INTO %s(%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s RETURNING %s", h.dbTbl, cols, vals, idCol, colValsAgain, idCol)
	h.queryDeletePrefix = fmt.Sprintf("DELETE FROM %s", h.dbTbl)
	h.queryUpdatePrefix = fmt.Sprintf("UPDATE %s SET", h.dbTbl)

	if h.hasDependencies {
		h.querySelectById = fmt.Sprintf("SELECT %s FROM %s t1%s WHERE %s = $1", cols, h.dbTbl, innerJoins, idCol)
		h.querySelectPrefix = fmt.Sprintf("SELECT %s FROM %s t1%s", cols, h.dbTbl, innerJoins)
		h.querySelectCountPrefix = fmt.Sprintf("SELECT COUNT(*) AS cnt FROM %s t1%s", h.dbTbl, innerJoins)
	} else {
		h.querySelectById = fmt.Sprintf("SELECT %s FROM %s%s WHERE %s = $1", cols, h.dbTbl, innerJoins, idCol)
		h.querySelectPrefix = fmt.Sprintf("SELECT %s FROM %s%s", cols, h.dbTbl, innerJoins)
		h.querySelectCountPrefix = fmt.Sprintf("SELECT COUNT(*) AS cnt FROM %s%s", h.dbTbl, innerJoins)
	}

}

func (h *StructSQL) reflectStructForValidation(u interface{}) {
	v := reflect.ValueOf(u)
	i := reflect.Indirect(v)
	s := i.Type()

	h.fieldsDefaultValue = make(map[string]string)
	h.fieldsUniq = make(map[string]bool)
	h.fieldsTags = make(map[string]map[string]string)
	h.fieldsOverwriteType = make(map[string]string)

	reDep := regexp.MustCompile(`^[a-zA-Z0-9]+_[a-zA-Z0-9]+`)

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

		// If field is like 'xxx_yyy', try to set tags from dependencies for 'xxx'
		if h.dependenciesFieldsTags != nil && reDep.MatchString(f.Name) {
			fieldNameArr := strings.Split(f.Name, "_")

			// Mark hasDependencies
			_, ok := h.dependenciesFieldsTags[fieldNameArr[0]]
			if !ok {
				continue
			}
			h.hasDependencies = true

			_, ok = h.dependenciesFieldsTags[fieldNameArr[0]][fieldNameArr[1]]
			if !ok {
				continue
			}

			if crudTag == "" && h.dependenciesFieldsTags[fieldNameArr[0]][fieldNameArr[1]][h.tagName] != "" {
				crudTag = h.dependenciesFieldsTags[fieldNameArr[0]][fieldNameArr[1]][h.tagName]
			}
			if crudValTag == "" && h.dependenciesFieldsTags[fieldNameArr[0]][fieldNameArr[1]][h.tagName+"_val"] != "" {
				crudValTag = h.dependenciesFieldsTags[fieldNameArr[0]][fieldNameArr[1]][h.tagName+"_val"]
			}
		}

		// Go through tag values and parse out the ones we're interested in
		h.setFieldFromTag(crudTag, f.Name)
		if h.err != nil {
			return
		}

		if crudValTag != "" {
			h.fieldsDefaultValue[f.Name] = crudValTag
		}

		// Store original field tags (non-overwritten one) so they can be easily returned and used as
		// defaultFieldsTags in another struct
		h.fieldsTags[f.Name] = make(map[string]string)
		h.fieldsTags[f.Name][h.tagName] = f.Tag.Get(h.tagName)
		h.fieldsTags[f.Name][h.tagName+"_val"] = f.Tag.Get(h.tagName+"_val")
	}
}

func (h *StructSQL) setFieldFromTag(tag string, fieldName string) {
	opts := strings.SplitN(tag, " ", -1)
	for _, opt := range opts {
		h.setFieldFromTagOptWithoutVal(opt, fieldName)
	}
}

func (h *StructSQL) setFieldFromTagOptWithoutVal(opt string, fieldName string) {
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

func (h *StructSQL) getDBCol(n string) string {
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
func (h *StructSQL) getDBColParams(n string, t string, uniq bool) string {
	dbColParams := ""
	if n == "ID" {
		dbColParams = "SERIAL PRIMARY KEY"
	} else if n == "Flags" {
		dbColParams = "BIGINT NOT NULL DEFAULT 0"
		// String types can be overwritten by a tag
	} else if h.fieldsOverwriteType[n] != "" {
		dbColParams = h.fieldsOverwriteType[n]+" NOT NULL DEFAULT ''"
	} else {
		switch t {
		case "string":
			dbColParams = "VARCHAR(255) NOT NULL DEFAULT ''"
		case "bool":
			dbColParams = "BOOLEAN NOT NULL DEFAULT false"
		case "int64":
			dbColParams = "BIGINT NOT NULL DEFAULT 0"
		case "int32":
			dbColParams = "INTEGER NOT NULL DEFAULT 0"
		case "int16":
			dbColParams = "SMALLINT NOT NULL DEFAULT 0"
		case "int8":
			dbColParams = "SMALLINT NOT NULL DEFAULT 0"
		case "int":
			dbColParams = "BIGINT NOT NULL DEFAULT 0"
		case "uint64":
			dbColParams = "BIGINT NOT NULL DEFAULT 0"
		case "uint32":
			dbColParams = "INTEGER NOT NULL DEFAULT 0"
		case "uint16":
			dbColParams = "SMALLINT NOT NULL DEFAULT 0"
		case "uint8":
			dbColParams = "SMALLINT NOT NULL DEFAULT 0"
		case "uint":
			dbColParams = "BIGINT NOT NULL DEFAULT 0"
		// TODO: Consider something different
		default:
			dbColParams = "VARCHAR(255) NOT NULL DEFAULT ''"
		}
	}
	if uniq {
		dbColParams += " UNIQUE"
	}
	return dbColParams
}

func (h *StructSQL) addWithComma(s string, v string) string {
	if s != "" {
		s += ","
	}
	s += v
	return s
}

func (h *StructSQL) addWithAnd(s string, v string) string {
	if s != "" {
		s += " AND "
	}
	s += v
	return s
}

func (h *StructSQL) getUnderscoredName(s string) string {
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

func (h *StructSQL) getPluralName(s string) string {
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

func (h *StructSQL) getQueryOrder(order []string, orderFieldsToInclude map[string]bool) string {
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

func (h *StructSQL) getQueryLimitOffset(limit int, offset int) string {
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

func (h *StructSQL) getQuerySet(values map[string]interface{}, valueFieldsToInclude map[string]bool) (string, int) {
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

func (h *StructSQL) getQueryFilters(filters map[string]interface{}, filterFieldsToInclude map[string]bool, firstNumber int) string {
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

		reField := regexp.MustCompile(`\.[a-zA-Z0-9_]+`)
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
