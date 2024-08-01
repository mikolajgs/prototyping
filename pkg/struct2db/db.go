package struct2db

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/mikolajgs/crud/pkg/struct2sql"
)

// Save takes object, validates its field values and saves it in the database.
// If ID field is already set (it's greater than 0) then the function assumes that record with such ID already
// exists in the database and the function with execute an "UPDATE" query. Otherwise it will be "INSERT". After
// inserting, new record ID is set to struct's ID field
func (c Controller) Save(obj interface{}) *ErrController {
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
	if c.GetModelIDValue(obj) != 0 {
		_, err3 = c.dbConn.Exec(h.GetQueryUpdateById(), append(c.GetModelFieldInterfaces(obj), c.GetModelIDInterface(obj))...)
	} else {
		err3 = c.dbConn.QueryRow(h.GetQueryInsert(), c.GetModelFieldInterfaces(obj)...).Scan(c.GetModelIDInterface(obj))
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
	err3 := c.dbConn.QueryRow(h.GetQuerySelectById(), int64(idInt)).Scan(append(append(make([]interface{}, 0), c.GetModelIDInterface(obj)), c.GetModelFieldInterfaces(obj)...)...)
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
func (c Controller) Delete(obj interface{}) *ErrController {
	h, err := c.getSQLGenerator(obj)
	if err != nil {
		return err
	}
	if c.GetModelIDValue(obj) == 0 {
		return nil
	}
	_, err2 := c.dbConn.Exec(h.GetQueryDeleteById(), c.GetModelIDInterface(obj))
	if err2 != nil {
		return &ErrController{
			Op:  "DBQuery",
			Err: fmt.Errorf("Error executing DB query: %w", err2),
		}
	}
	c.ResetFields(obj)
	return nil
}

// Get runs a select query on the database with specified filters, order, limit and offset and returns a
// list of objects
func (c Controller) Get(newObjFunc func() interface{}, order []string, limit int, offset int, filters map[string]interface{}) ([]interface{}, *ErrController) {
	obj := newObjFunc()
	h, err := c.getSQLGenerator(obj)
	if err != nil {
		return nil, err
	}

	if len(filters) > 0 {
		b, invalidFields, err1 := c.Validate(obj, filters)
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
	rows, err2 := c.dbConn.Query(h.GetQuerySelect(order, limit, offset, filters, nil, nil), c.GetFiltersInterfaces(filters)...)
	if err2 != nil {
		return nil, &ErrController{
			Op:  "DBQuery",
			Err: fmt.Errorf("Error executing DB query: %w", err2),
		}
	}
	defer rows.Close()

	for rows.Next() {
		newObj := newObjFunc()
		err3 := rows.Scan(append(append(make([]interface{}, 0), c.GetModelIDInterface(newObj)), c.GetModelFieldInterfaces(newObj)...)...)
		if err3 != nil {
			return nil, &ErrController{
				Op:  "DBQueryRowsScan",
				Err: fmt.Errorf("Error scanning DB query row: %w", err3),
			}
		}
		v = append(v, newObj)
	}

	return v, nil
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

// DeleteHorizontal deletes an object along with linked objects that would usually be selected with JOIN
// TODO: Re-phrase it
func (c Controller) DeleteHorizontal(obj interface{}, lnks ...interface{}) *ErrController {
	return nil
}
