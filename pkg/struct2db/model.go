package struct2db

import (
	"reflect"
	"sort"
)

// GetModelIDInterface returns an interface{} to ID field of an object
func (c *Controller) GetModelIDInterface(obj interface{}) interface{} {
	return reflect.ValueOf(obj).Elem().FieldByName("ID").Addr().Interface()
}

// GetModelIDValue returns value of ID field (int64) of an object
func (c *Controller) GetModelIDValue(obj interface{}) int64 {
	return reflect.ValueOf(obj).Elem().FieldByName("ID").Int()
}

// GetModelFieldInterfaces returns list of interfaces to object's fields without the ID field
func (c Controller) GetModelFieldInterfaces(obj interface{}) []interface{} {
	val := reflect.ValueOf(obj).Elem()

	var v []interface{}
	for i := 1; i < val.NumField(); i++ {
		valueField := val.Field(i)
		if valueField.Kind() != reflect.Int64 && valueField.Kind() != reflect.Int && valueField.Kind() != reflect.String {
			continue
		}
		v = append(v, valueField.Addr().Interface())
	}
	return v
}

// GetFiltersInterfaces returns list of interfaces from filters map (used in querying)
func (c Controller) GetFiltersInterfaces(mf map[string]interface{}) []interface{} {
	var xi []interface{}

	if len(mf) == 0 {
		return xi
	}

	sorted := []string{}
	for k := range mf {
		if k == "_raw" || k == "_rawConjuction" {
			continue
		}
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	for _, v := range sorted {
		xi = append(xi, mf[v])
	}

	// Get pointers to values from raw query
	_, ok := mf["_raw"]
	if !ok {
		return xi
	}

	rt := reflect.TypeOf(mf["_raw"])
	if rt.Kind() != reflect.Slice && rt.Kind() != reflect.Array {
		return xi
	}
	if reflect.ValueOf(mf["_raw"]).Len() < 2 {
		return xi
	}

	for i:=1; i<reflect.ValueOf(mf["_raw"]).Len(); i++ {
		rt2 := reflect.TypeOf(mf["_raw"].([]interface{})[i])
		if rt2.Kind() == reflect.Slice || rt2.Kind() == reflect.Array {
			// TODO: Replace with generics
			valInts, ok := mf["_raw"].([]interface{})[i].([]int)
			if ok {
				for j:=0; j<len(valInts); j++ {
					xi = append(xi, valInts[j])
				}
				continue
			}
			valInt64s, ok := mf["_raw"].([]interface{})[i].([]int64)
			if ok {
				for j:=0; j<len(valInt64s); j++ {
					xi = append(xi, valInt64s[j])
				}
				continue
			}
			valBools, ok := mf["_raw"].([]interface{})[i].([]bool)
			if ok {
				for j:=0; j<len(valBools); j++ {
					xi = append(xi, valBools[j])
				}
				continue
			}
			valStrings, ok := mf["_raw"].([]interface{})[i].([]bool)
			if ok {
				for j:=0; j<len(valStrings); j++ {
					xi = append(xi, valStrings[j])
				}
			}
		} else {
			xi = append(xi, mf["_raw"].([]interface{})[i])
		}
	}

	return xi
}

// ResetFields zeroes object's field values
func (c Controller) ResetFields(obj interface{}) {
	val := reflect.ValueOf(obj).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		if valueField.Kind() == reflect.Ptr {
			valueField.Set(reflect.Zero(valueField.Type()))
		}
		if valueField.Kind() == reflect.Int64 {
			valueField.SetInt(0)
		}
		if valueField.Kind() == reflect.Int {
			valueField.SetInt(0)
		}
		if valueField.Kind() == reflect.String {
			valueField.SetString("")
		}
	}
}
