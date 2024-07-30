package crud

import (
	"fmt"
)

// initHelpers creates all the Struct2sql objects. For HTTP endpoints, it is necessary to create these first
func (c *Controller) initHelpersForHTTPHandler(newObjFunc func() interface{}, newObjCreateFunc func() interface{}, newObjReadFunc func() interface{}, newObjUpdateFunc func() interface{}, newObjDeleteFunc func() interface{}, newObjListFunc func() interface{}) *ErrController {
	obj := newObjFunc()

	cErr := c.struct2db.AddSQLGenerator(newObjFunc(), obj, false)
	if cErr != nil {
		return &ErrController{
			Op:  "AddSQLGenerator",
			Err: fmt.Errorf("Error adding SQL generator: %w", cErr.Unwrap()),
		}
	}

	for _, f := range []func() interface{}{
		newObjCreateFunc,
		newObjReadFunc,
		newObjUpdateFunc,
		newObjDeleteFunc,
		newObjListFunc,
	} {
		if f == nil {
			continue
		}

		cErr = c.struct2db.AddSQLGenerator(f(), obj, false)
		if cErr != nil {
			return &ErrController{
				Op:  "AddSQLGenerator",
				Err: fmt.Errorf("Error adding SQL generator: %w", cErr.Unwrap()),
			}
		}
	}

	return nil
}

func (c Controller) mapWithInterfacesToMapBool(m map[string]interface{}) map[string]bool {
	o := map[string]bool{}
	for k, _ := range m {
		o[k] = true
	}
	return o
}
