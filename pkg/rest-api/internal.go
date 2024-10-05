package restapi

import (
	"fmt"
)

// initHelpers creates all the Struct2sql objects. For HTTP endpoints, it is necessary to create these first
func (c *Controller) initHelpers(newObjFunc func() interface{}, options HandlerOptions) *ErrController {
	obj := newObjFunc()

	var forceName string
	if options.ForceName != "" {
		forceName = options.ForceName
	}

	cErr := c.struct2db.AddSQLGenerator(obj, nil, false, forceName, false)
	if cErr != nil {
		return &ErrController{
			Op:  "AddSQLGenerator",
			Err: fmt.Errorf("Error adding SQL generator: %w", cErr.Unwrap()),
		}
	}

	if options.CreateConstructor != nil {
		cErr = c.struct2db.AddSQLGenerator(options.CreateConstructor(), obj, false, "", true)
		if cErr != nil {
			return &ErrController{
				Op:  "AddSQLGenerator",
				Err: fmt.Errorf("Error adding SQL generator: %w", cErr.Unwrap()),
			}
		}
	}

	if options.ReadConstructor != nil {
		cErr = c.struct2db.AddSQLGenerator(options.ReadConstructor(), obj, false, "", true)
		if cErr != nil {
			return &ErrController{
				Op:  "AddSQLGenerator",
				Err: fmt.Errorf("Error adding SQL generator: %w", cErr.Unwrap()),
			}
		}
	}

	if options.UpdateConstructor != nil {
		cErr = c.struct2db.AddSQLGenerator(options.UpdateConstructor(), obj, false, "", true)
		if cErr != nil {
			return &ErrController{
				Op:  "AddSQLGenerator",
				Err: fmt.Errorf("Error adding SQL generator: %w", cErr.Unwrap()),
			}
		}
	}

	if options.ListConstructor != nil {
		cErr = c.struct2db.AddSQLGenerator(options.ListConstructor(), obj, false, "", true)
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
