package struct2sql

// Struct2sql reflects the object to generate and cache PostgreSQL queries (CREATE TABLE, INSERT, UPDATE etc.).
// Database table and column names are lowercase with underscore and they are generated from field names.
// Struct2sql is created within Controller and there is no need to instantiate it
type Struct2sql struct {
	queryDropTable    string
	queryCreateTable  string
	queryInsert       string
	queryUpdateById   string
	queryInsertOnConflictUpdate string
	querySelectById   string
	queryDeleteById   string
	querySelectPrefix string
	querySelectCountPrefix string
	queryDeletePrefix string
	queryUpdatePrefix string

	dbTbl       string
	dbColPrefix string
	dbFieldCols map[string]string
	dbCols      map[string]string
	url         string
	fields      []string

	fieldsDefaultValue map[string]string
	fieldsUniq         map[string]bool
	fieldsTags         map[string]map[string]string
	fieldsOverwriteType map[string]string

	flags int

	defaultFieldsTags map[string]map[string]string

	err *ErrStruct2sql

	tagName string
}

const RawConjuctionOR = 1
const RawConjuctionAND = 2

type Struct2sqlOptions struct {
	DatabaseTablePrefix string
	ForceName string
	SourceStruct2sql *Struct2sql
	TagName string
}

// NewStruct2sql takes object and database table name prefix as arguments and returns Struct2sql instance.
func NewStruct2sql(obj interface{}, options Struct2sqlOptions) *Struct2sql {
	h := &Struct2sql{}

	h.tagName = "2sql"
	if options.TagName != "" {
		h.tagName = options.TagName
	}

	h.setDefaultTags(options.SourceStruct2sql)
	h.reflectStruct(obj, options.DatabaseTablePrefix, options.ForceName)
	return h
}

// Err returns error that occurred when reflecting struct
func (h *Struct2sql) Err() *ErrStruct2sql {
	return h.err
}

// GetFlags returns Struct2sql flags.
func (h *Struct2sql) GetFlags() int {
	return h.flags
}

// GetQueryDropTable returns a DROP TABLE query.
func (h Struct2sql) GetQueryDropTable() string {
	return h.queryDropTable
}

// GetQueryCreateTable return a CREATE TABLE query.
// Columns in the query are ordered the same way as they are defined in the struct, eg. SELECT field1_column, field2_column, ... etc.
func (h Struct2sql) GetQueryCreateTable() string {
	return h.queryCreateTable
}

// GetQueryInsert returns an INSERT query.
// Columns in the INSERT query are ordered the same way as they are defined in the struct, eg. SELECT field1_column, field2_column, ... etc.
func (h *Struct2sql) GetQueryInsert() string {
	return h.queryInsert
}

// GetQueryUpdateById returns an UPDATE query with WHERE condition on ID field.
// Columns in the UPDATE query are ordered the same way as they are defined in the struct, eg. SELECT field1_column, field2_column, ... etc.
func (h *Struct2sql) GetQueryUpdateById() string {
	return h.queryUpdateById
}

// GetQueryInsertOnConflictUpdate returns an "upsert" query, which will INSERT data when it does not exist or UPDATE it otherwise.
// Columns in the query are ordered the same way as they are defined in the struct, eg. SELECT field1_column, field2_column, ... etc.
func (h *Struct2sql) GetQueryInsertOnConflictUpdate() string {
	return h.queryInsertOnConflictUpdate
}

// GetQuerySelectById returns a SELECT query with WHERE condition on ID field.
// Columns in the SELECT query are ordered the same way as they are defined in the struct, eg. SELECT field1_column, field2_column, ... etc.
func (h *Struct2sql) GetQuerySelectById() string {
	return h.querySelectById
}

// GetQueryDeleteById returns a DELETE query with WHERE condition on ID field.
func (h *Struct2sql) GetQueryDeleteById() string {
	return h.queryDeleteById
}

// GetQuerySelect returns a SELECT query with WHERE condition built from 'filters' (field-value pairs).
// Struct fields in 'filters' argument are sorted alphabetically. Hence, when used with database connection, their values (or pointers to it) must be sorted as well.
// Columns in the SELECT query are ordered the same way as they are defined in the struct, eg. SELECT field1_column, field2_column, ... etc.
func (h *Struct2sql) GetQuerySelect(order []string, limit int, offset int, filters map[string]interface{}, orderFieldsToInclude map[string]bool, filterFieldsToInclude map[string]bool) string {
	s := h.querySelectPrefix

	qOrder := h.getQueryOrder(order, orderFieldsToInclude)
	qLimitOffset := h.getQueryLimitOffset(limit, offset)
	qWhere := h.getQueryFilters(filters, filterFieldsToInclude, 1)

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

// GetQuerySelectCount returns a SELECT COUNT(*) query to count rows with WHERE condition built from 'filters' (field-value pairs).
// Struct fields in 'filters' argument are sorted alphabetically. Hence, when used with database connection, their values (or pointers to it) must be sorted as well.
func (h *Struct2sql) GetQuerySelectCount(filters map[string]interface{}, filterFieldsToInclude map[string]bool) string {
	s := h.querySelectCountPrefix
	qWhere := h.getQueryFilters(filters, filterFieldsToInclude, 1)
	if qWhere != "" {
		s += " WHERE " + qWhere
	}
	return s
}

// GetQueryDelete return a DELETE query with WHERE condition built from 'filters' (field-value pairs).
// Struct fields in 'filters' argument are sorted alphabetically. Hence, when used with database connection, their values (or pointers to it) must be sorted as well.
func (h *Struct2sql) GetQueryDelete(filters map[string]interface{}, filterFieldsToInclude map[string]bool) string {
	s := h.queryDeletePrefix
	qWhere := h.getQueryFilters(filters, filterFieldsToInclude, 1)
	if qWhere != "" {
		s += " WHERE " + qWhere
	}
	return s
}

// GetQueryUpdate returns an UPDATE query where specified struct fields (columns) are updated and rows match specific WHERE condition built from 'filters' (field-value pairs).
// Struct fields in 'values' and 'filters' arguments, are sorted alphabetically. Hence, when used with database connection, their values (or pointers to it) must be sorted as well.
func (h *Struct2sql) GetQueryUpdate(values map[string]interface{}, filters map[string]interface{}, valueFieldsToInclude map[string]bool, filterFieldsToInclude map[string]bool) string {
	s := h.queryUpdatePrefix

	qSet, lastVarNumber := h.getQuerySet(values, valueFieldsToInclude)
	s += " " + qSet

	qWhere := h.getQueryFilters(filters, filterFieldsToInclude, lastVarNumber+1)
	if qWhere != "" {
		s += " WHERE " + qWhere
	}

	return s
}

// GetFieldNameFromDBCol returns field name from a database column.
func (h *Struct2sql) GetFieldNameFromDBCol(n string) string {
	return h.dbCols[n]
}
