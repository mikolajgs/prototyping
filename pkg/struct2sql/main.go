package struct2sql

import (
	"fmt"
	"sort"
	"strings"
)

// Struct2sql reflects the object to generate and cache PostgreSQL queries (CREATE TABLE, INSERT, UPDATE etc.).
// Database table and column names are lowercase with underscore and they are generated from field names.
// Struct2sql is created within Controller and there is no need to instantiate it
type Struct2sql struct {
	queryDropTable    string
	queryCreateTable  string
	queryInsert       string
	queryUpdateById   string
	querySelectById   string
	queryDeleteById   string
	querySelectPrefix string

	dbTbl       string
	dbColPrefix string
	dbFieldCols map[string]string
	dbCols      map[string]string
	url         string
	fields      []string

	fieldsDefaultValue map[string]string
	fieldsUniq         map[string]bool
	fieldsTags         map[string]map[string]string

	fieldsFlags map[string]int

	flags int

	defaultFieldsTags map[string]map[string]string

	err *ErrStruct2sql
}

const TypeInt64 = 64
const TypeInt = 128
const TypeString = 256

// NewStruct2sql takes object and database table name prefix as arguments and returns Struct2sql instance
func NewStruct2sql(obj interface{}, dbTblPrefix string, forceName string, sourceStruct2sql *Struct2sql) *Struct2sql {
	h := &Struct2sql{}
	h.setDefaultTags(sourceStruct2sql)
	h.reflectStruct(obj, dbTblPrefix, forceName)
	return h
}

// Err returns error that occurred when reflecting struct
func (h *Struct2sql) Err() *ErrStruct2sql {
	return h.err
}

// GetFlags returns flags
func (h *Struct2sql) GetFlags() int {
	return h.flags
}

// GetQueryDropTable returns drop table query
func (h Struct2sql) GetQueryDropTable() string {
	return h.queryDropTable
}

// GetQueryCreateTable return create table query
func (h Struct2sql) GetQueryCreateTable() string {
	return h.queryCreateTable
}

// GetQueryInsert returns insert query
func (h *Struct2sql) GetQueryInsert() string {
	return h.queryInsert
}

// GetQueryUpdateById returns update query
func (h *Struct2sql) GetQueryUpdateById() string {
	return h.queryUpdateById
}

// GetQuerySelectById returns select query
func (h *Struct2sql) GetQuerySelectById() string {
	return h.querySelectById
}

// GetQueryDeleteById returns delete query
func (h *Struct2sql) GetQueryDeleteById() string {
	return h.queryDeleteById
}

func (h *Struct2sql) GetQuerySelect(order []string, limit int, offset int, filters map[string]interface{}, orderFieldsToInclude map[string]bool, filterFieldsToInclude map[string]bool) string {
	s := h.querySelectPrefix

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

	qLimitOffset := ""
	if limit > 0 {
		if offset > 0 {
			qLimitOffset = fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
		} else {
			qLimitOffset = fmt.Sprintf("LIMIT %d", limit)
		}
	}

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

	if qWhere != "" {
		s += " WHERE " + qWhere
	}
	if qOrder != "" {
		s += " ORDER BY " + qOrder
	}
	if qLimitOffset != "" {
		s += " " + qLimitOffset
	}
	return s
}

func (h *Struct2sql) GetFieldNameFromDBCol(n string) string {
	return h.dbCols[n]
}
