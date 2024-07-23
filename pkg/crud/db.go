package crud

import (
	"database/sql"
	"fmt"
	"strconv"
)

// SaveToDB takes object, validates its field values and saves it in the database.
// If ID field is already set (it's greater than 0) then the function assumes that record with such ID already
// exists in the database and the function with execute an "UPDATE" query. Otherwise it will be "INSERT". After
// inserting, new record ID is set to struct's ID field
func (c Controller) SaveToDB(obj interface{}) *ErrController {
	h, err := c.getHelper(obj)
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

// SetFromDB sets object's fields with values from the database table with a specific id. If record does not exist
// in the database, all field values in the struct are zeroed
func (c Controller) SetFromDB(obj interface{}, id string) *ErrController {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return &ErrController{
			Op:  "IDToInt",
			Err: fmt.Errorf("Error converting string to int: %w", err),
		}
	}

	h, err2 := c.getHelper(obj)
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

// DeleteFromDB removes object from the database table and it does that only when ID field is set (greater than 0).
// Once deleted from the DB, all field values are zeroed
func (c Controller) DeleteFromDB(obj interface{}) *ErrController {
	h, err := c.getHelper(obj)
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

// GetFromDB runs a select query on the database with specified filters, order, limit and offset and returns a
// list of objects
func (c Controller) GetFromDB(newObjFunc func() interface{}, order []string, limit int, offset int, filters map[string]interface{}) ([]interface{}, *ErrController) {
	obj := newObjFunc()
	h, err := c.getHelper(obj)
	if err != nil {
		return nil, err
	}

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
