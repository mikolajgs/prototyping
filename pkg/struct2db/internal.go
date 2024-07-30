package struct2db

import (
	"fmt"
	"reflect"

	"github.com/mikolajgs/crud/pkg/struct2sql"
)

// getSQLGenerator returns a special Struct2sql instance which reflects the struct type to get SQL queries etc.
func (c *Controller) getSQLGenerator(obj interface{}) (*struct2sql.Struct2sql, *ErrController) {
	n := c.getSQLGeneratorName(obj)
	if c.sqlGenerators[n] == nil {
		h := struct2sql.NewStruct2sql(obj, c.dbTblPrefix, "", nil)
		if h.Err() != nil {
			return nil, &ErrController{
				Op:  "GetHelper",
				Err: fmt.Errorf("Error getting Struct2sql: %w", h.Err()),
			}
		}
		c.sqlGenerators[n] = h
	}
	return c.sqlGenerators[n], nil
}

func (c *Controller) getSQLGeneratorName(obj interface{}) string {
	v := reflect.ValueOf(obj)
	i := reflect.Indirect(v)
	s := i.Type()
	n := s.Name()
	return n
}

func (c Controller) mapWithInterfacesToMapBool(m map[string]interface{}) map[string]bool {
	o := map[string]bool{}
	for k, _ := range m {
		o[k] = true
	}
	return o
}
