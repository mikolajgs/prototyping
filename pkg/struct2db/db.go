package struct2db

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/mikolajgs/crud/pkg/struct2sql"
)

const RawConjuctionOR = 1
const RawConjuctionAND = 2

type SaveOptions struct {
	NoInsert bool
}

type GetOptions struct {
	Order []string
	Limit int
	Offset int
	Filters map[string]interface{}
	RowObjTransformFunc func(interface{}) interface{}
}

type DeleteOptions struct {
	Constructors map[string]func() interface{}
}

type DeleteMultipleOptions struct {
	Filters map[string]interface{}
}

type GetCountOptions struct {
	Filters map[string]interface{}
}

// Save takes object, validates its field values and saves it in the database.
// If ID is not present then an INSERT will be performed
// If ID is set then an "upsert" is performed
func (c Controller) Save(obj interface{}, options SaveOptions) *ErrController {
	h, err := c.getSQLGenerator(obj)
	if err != nil {
		return err
	}

	b, invalidFields, err2 := c.Validate(obj, nil)
	if err2 != nil {
		return &ErrController{
			Op:  "Validate",
			Err: fmt.Errorf("Error when trying to validate: %w", err2),
		}
	}

	if !b {
		return &ErrController{
			Op: "Validate",
			Err: &ErrValidation{
				Fields: invalidFields,
			},
		}
	}

	var err3 error
	if c.GetObjIDValue(obj) != 0 {
		// do no try to insert if NoInsert is set
		// TODO: error handling, we should check if object exists - for now nothing happens, UPDATE gets executed and updates nothing
		if options.NoInsert {
			_, err3 = c.dbConn.Exec(h.GetQueryUpdateById(), append(c.GetObjFieldInterfaces(obj, false), c.GetObjIDInterface(obj))...)
		} else {
			// try to insert - if ID already exists then try to update it
			_, err3 = c.dbConn.Exec(h.GetQueryInsertOnConflictUpdate(), append(c.GetObjFieldInterfaces(obj, true), c.GetObjFieldInterfaces(obj, false)...)...)
		}
	} else {
		err3 = c.dbConn.QueryRow(h.GetQueryInsert(), c.GetObjFieldInterfaces(obj, false)...).Scan(c.GetObjIDInterface(obj))
	}
	if err3 != nil {
		return &ErrController{
			Op:  "DBQuery",
			Err: fmt.Errorf("Error executing DB query: %w", err3),
		}
	}
	return nil
}

// Load sets object's fields with values from the database table with a specific id. If record does not exist
// in the database, all field values in the struct are zeroed
func (c Controller) Load(obj interface{}, id string) *ErrController {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return &ErrController{
			Op:  "IDToInt",
			Err: fmt.Errorf("Error converting string to int: %w", err),
		}
	}

	h, err2 := c.getSQLGenerator(obj)
	if err2 != nil {
		return err2
	}
	err3 := c.dbConn.QueryRow(h.GetQuerySelectById(), int64(idInt)).Scan(c.GetObjFieldInterfaces(obj, true)...)
	switch {
	case err3 == sql.ErrNoRows:
		c.ResetFields(obj)
		return nil
	case err3 != nil:
		return &ErrController{
			Op:  "DBQuery",
			Err: fmt.Errorf("Error executing DB query: %w", err),
		}
	default:
		return nil
	}
}

// Delete removes object from the database table and it does that only when ID field is set (greater than 0).
// Once deleted from the DB, all field values are zeroed
func (c Controller) Delete(obj interface{}, options DeleteOptions) *ErrController {
	h, err := c.getSQLGenerator(obj)
	if err != nil {
		return err
	}
	if c.GetObjIDValue(obj) == 0 {
		return nil
	}
	_, err2 := c.dbConn.Exec(h.GetQueryDeleteById(), c.GetObjIDInterface(obj))
	if err2 != nil {
		return &ErrController{
			Op:  "DBQuery",
			Err: fmt.Errorf("Error executing DB query: %w", err2),
		}
	}
	c.ResetFields(obj)
	return nil
}

// DeleteMultiple removes objects from the database based on specified filters
func (c Controller) DeleteMultiple(newObjFunc func() interface{}, options DeleteMultipleOptions) (*ErrController) {
	obj := newObjFunc()
	h, err := c.getSQLGenerator(obj)
	if err != nil {
		return err
	}

	if len(options.Filters) > 0 {
		b, invalidFields, err1 := c.Validate(obj, options.Filters)
		if err1 != nil {
			return &ErrController{
				Op:  "ValidateFilters",
				Err: fmt.Errorf("Error when trying to validate filters: %w", err1),
			}
		}

		if !b {
			return &ErrController{
				Op: "ValidateFilters",
				Err: &ErrValidation{
					Fields: invalidFields,
				},
			}
		}
	}

	_, err2 := c.dbConn.Exec(h.GetQueryDelete(options.Filters, nil), c.GetFiltersInterfaces(options.Filters)...)
	if err2 != nil {
		return &ErrController{
			Op:  "DBQuery",
			Err: fmt.Errorf("Error executing DB query: %w", err2),
		}
	}

	return nil
}

// Get runs a select query on the database with specified filters, order, limit and offset and returns a
// list of objects
func (c Controller) Get(newObjFunc func() interface{}, options GetOptions) ([]interface{}, *ErrController) {
	obj := newObjFunc()
	h, err := c.getSQLGenerator(obj)
	if err != nil {
		return nil, err
	}

	if len(options.Filters) > 0 {
		b, invalidFields, err1 := c.Validate(obj, options.Filters)
		if err1 != nil {
			return nil, &ErrController{
				Op:  "ValidateFilters",
				Err: fmt.Errorf("Error when trying to validate filters: %w", err1),
			}
		}

		if !b {
			return nil, &ErrController{
				Op: "ValidateFilters",
				Err: &ErrValidation{
					Fields: invalidFields,
				},
			}
		}
	}

	var v []interface{}
	rows, err2 := c.dbConn.Query(h.GetQuerySelect(options.Order, options.Limit, options.Offset, options.Filters, nil, nil), c.GetFiltersInterfaces(options.Filters)...)
	if err2 != nil {
		return nil, &ErrController{
			Op:  "DBQuery",
			Err: fmt.Errorf("Error executing DB query: %w", err2),
		}
	}
	defer rows.Close()

	for rows.Next() {
		newObj := newObjFunc()
		err3 := rows.Scan(c.GetObjFieldInterfaces(newObj, true)...)
		if err3 != nil {
			return nil, &ErrController{
				Op:  "DBQueryRowsScan",
				Err: fmt.Errorf("Error scanning DB query row: %w", err3),
			}
		}

		// If options.RowObjTransformFunc is defined then call it on the row
		if options.RowObjTransformFunc != nil {
			v = append(v, options.RowObjTransformFunc(newObj))
			continue
		}

		// Normal append
		v = append(v, newObj)
	}

	return v, nil
}

// GetCount runs a 'SELECT COUNT(*)' query on the database with specified filters, order, limit and offset and returns count of rows
func (c Controller) GetCount(newObjFunc func() interface{}, options GetCountOptions) (int64, *ErrController) {
	obj := newObjFunc()
	h, err := c.getSQLGenerator(obj)
	if err != nil {
		return 0, err
	}

	if len(options.Filters) > 0 {
		b, invalidFields, err1 := c.Validate(obj, options.Filters)
		if err1 != nil {
			return 0, &ErrController{
				Op:  "ValidateFilters",
				Err: fmt.Errorf("Error when trying to validate filters: %w", err1),
			}
		}

		if !b {
			return 0, &ErrController{
				Op: "ValidateFilters",
				Err: &ErrValidation{
					Fields: invalidFields,
				},
			}
		}
	}

	row := c.dbConn.QueryRow(h.GetQuerySelectCount(options.Filters, nil), c.GetFiltersInterfaces(options.Filters)...)
	var cnt int64
	err3 := row.Scan(&cnt)
	if err3 != nil {
		return 0, &ErrController{
			Op:  "DBQueryRowScan",
			Err: fmt.Errorf("Error scanning DB query row: %w", err3),
		}
	}

	return cnt, nil
}

// AddSQLGenerator adds Struct2sql object to sqlGenerators
func (c *Controller) AddSQLGenerator(obj interface{}, parentObj interface{}, overwrite bool) *ErrController {
	n := c.getSQLGeneratorName(obj)

	// If sql generator already exists and it should not be overwritten then finish
	if !overwrite {
		_, ok := c.sqlGenerators[n]
		if ok {
			return nil
		}
	}

	var sourceHelper *struct2sql.Struct2sql
	var forceName string
	if parentObj != nil {
		h, err := c.getSQLGenerator(parentObj)
		if err != nil {
			return &ErrController{
				Op:  "GetHelper",
				Err: fmt.Errorf("Error getting Struct2sql: %w", h.Err()),
			}
		}
		sourceHelper = h
		forceName = c.getSQLGeneratorName(parentObj)
	}

	h := struct2sql.NewStruct2sql(obj, c.dbTblPrefix, forceName, sourceHelper)
	if h.Err() != nil {
		return &ErrController{
			Op:  "GetHelper",
			Err: fmt.Errorf("Error getting Struct2sql: %w", h.Err()),
		}
	}
	c.sqlGenerators[n] = h
	return nil
}

// GetDBCol returns column name used in the database
func (c *Controller) GetFieldNameFromDBCol(obj interface{}, dbCol string) (string, *ErrController) {
	h, err := c.getSQLGenerator(obj)
	if err != nil {
		return "", err
	}
	fieldName := h.GetFieldNameFromDBCol(dbCol)
	return fieldName, nil
}
