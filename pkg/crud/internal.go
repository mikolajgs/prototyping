package crud

import (
	"fmt"
	"reflect"

	"github.com/mikolajgs/crud/pkg/struct2sql"
)

// initHelpers creates all the Struct2sql objects. For HTTP endpoints, it is necessary to create these first
func (c *Controller) initHelpersForHTTPHandler(newObjFunc func() interface{}, newObjCreateFunc func() interface{}, newObjReadFunc func() interface{}, newObjUpdateFunc func() interface{}, newObjDeleteFunc func() interface{}, newObjListFunc func() interface{}) *ErrController {
	obj := newObjFunc()
	v := reflect.ValueOf(obj)
	i := reflect.Indirect(v)
	s := i.Type()
	forceName := s.Name()

	h, cErr := c.struct2db.GetHelper(obj)
	if cErr != nil {
		return cErr
	}

	cErr = c.initHelper(newObjCreateFunc, forceName, h)
	if cErr != nil {
		return cErr
	}
	cErr = c.initHelper(newObjReadFunc, forceName, h)
	if cErr != nil {
		return cErr
	}
	cErr = c.initHelper(newObjUpdateFunc, forceName, h)
	if cErr != nil {
		return cErr
	}
	cErr = c.initHelper(newObjDeleteFunc, forceName, h)
	if cErr != nil {
		return cErr
	}
	cErr = c.initHelper(newObjListFunc, forceName, h)
	if cErr != nil {
		return cErr
	}

	return nil
}

func (c *Controller) initHelper(newObjFunc func() interface{}, forceName string, sourceHelper *struct2sql.Struct2sql) *ErrController {
	if newObjFunc == nil {
		return nil
	}

	obj := newObjFunc()

	h, err := c.struct2db.RegisterHelper(obj, forceName, sourceHelper)
	if err != nil {
		return &ErrController{
			Op:  "InitStruct2sqlWithForcedName",
			Err: fmt.Errorf("Error initialising Helper with forced name: %w", h.Err()),
		}
	}
	return nil
}

func (c Controller) isKeyInMap(k string, m map[string]interface{}) bool {
	for _, key := range reflect.ValueOf(m).MapKeys() {
		if key.String() == k {
			return true
		}
	}
	return false
}

func (c Controller) mapWithInterfacesToMapBool(m map[string]interface{}) map[string]bool {
	o := map[string]bool{}
	for k, _ := range m {
		o[k] = true
	}
	return o
}
