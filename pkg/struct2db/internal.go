package struct2db

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	stsql "github.com/mikolajgs/struct-sql-postgres"
)

// getSQLGenerator returns a special StructSQL instance which reflects the struct type to get SQL queries etc.
func (c *Controller) getSQLGenerator(obj interface{}) (*stsql.StructSQL, *ErrController) {
	n := c.getSQLGeneratorName(obj)
	if c.sqlGenerators[n] == nil {
		h := stsql.NewStructSQL(obj, stsql.StructSQLOptions{
			DatabaseTablePrefix: c.dbTblPrefix,
			TagName: c.tagName,
		})
		if h.Err() != nil {
			return nil, &ErrController{
				Op:  "GetHelper",
				Err: fmt.Errorf("Error getting StructSQL: %w", h.Err()),
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

func (c Controller) runOnDelete(obj interface{}, constructors map[string]func() interface{}, tagName string, ids []int64, lastDepth int) *ErrController {
	val := reflect.ValueOf(obj).Elem()
	typ := val.Type()
	structName := typ.Name()
	parentIDField := structName + "ID"

	tagRegexp := regexp.MustCompile(`[a-zA-Z0-9_]+\:[a-zA-Z0-9_-]+`)

	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		// Only field which are slices of pointers to struct instances
		if valueField.Kind() != reflect.Slice || valueField.Type().Elem().Kind() != reflect.Ptr || valueField.Type().Elem().Elem().Kind() != reflect.Struct {
			continue
		}
		childStructName := valueField.Type().Elem().Elem().Name()

		// Check if constructor is passed in the options - if not then then ignore the child
		_, ok := constructors[childStructName]
		if !ok {
			continue
		}

		// Get 2db tag, loop through its value on determine action based on it
		tag := typ.Field(i).Tag.Get(tagName)
		if tag == "" {
			continue
		}

		tags := strings.Split(tag, " ")
		tagsMap := map[string]string{}
		for _, t := range tags {
			if m := tagRegexp.MatchString(t); m {
				mArr := strings.Split(t, ":")
				tagsMap[mArr[0]] = mArr[1]
			}
		}	

		// Perform delete
		if tagsMap["on_del"] == "del" {
			if tagsMap["del_field"] != "" {
				parentIDField = tagsMap["del_field"]
			}
			// Delete from children table where parent ID = id of deleted object
			errCtl := c.DeleteMultiple(constructors[childStructName], DeleteMultipleOptions{
				Filters: map[string]interface{}{
					"_raw": []interface{}{
						fmt.Sprintf(".%s IN (?)", parentIDField),
						ids,
					},
				},
				CascadeDeleteDepth: lastDepth + 1,
				Constructors: constructors,
			})
			if errCtl != nil {
				return &ErrController{
					Op: "CascadeDelete",
					Err: errors.New("Error from DeleteMultiple"),
				}
			}
		}

		// Perform update
		if tagsMap["on_del"] == "upd" {
			updField := tagsMap["del_upd_field"]
			updValue := tagsMap["del_upd_val"]
			if updField == "" {
				return &ErrController{
					Op: "CascadeDelete",
					Err: errors.New("Missing update field in tags"),
				}
			}

			if tagsMap["del_field"] != "" {
				parentIDField = tagsMap["del_field"]
			}
			// Update children table where parent ID = id of deleted object
			errCtl := c.UpdateMultiple(constructors[childStructName],
				map[string]interface{}{
					updField: updValue,
				},
				UpdateMultipleOptions{
					Filters: map[string]interface{}{
						"_raw": []interface{}{
							fmt.Sprintf(".%s IN (?)", parentIDField),
							ids,
						},
					},
					ConvertValuesFromString: true,
				},
			)
			if errCtl != nil {
				return &ErrController{
					Op: "CascadeDelete",
					Err: errors.New("Error from UpdateMultiple"),
				}
			}
		}
	}

	return nil
}
