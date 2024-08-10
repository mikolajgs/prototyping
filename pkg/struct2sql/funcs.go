package struct2sql

import "reflect"

func GetStructName(u interface{}) string {
	v := reflect.ValueOf(u)
	i := reflect.Indirect(v)
	s := i.Type()
	return s.Name()
}

func GetStructFieldNames(u interface{}) []string {
	v := reflect.ValueOf(u)
	i := reflect.Indirect(v)
	s := i.Type()

	names := []string{}

	for j := 0; j < s.NumField(); j++ {
		field := s.Field(j)
		if field.Type.Kind() != reflect.Int64 && field.Type.Kind() != reflect.String && field.Type.Kind() != reflect.Int {
			continue
		}
		names = append(names, field.Name)
	}

	return names
}

func GetStructNamesFromContructors(objFuncs... func() interface{}) []string {
	names := []string{}
	for _, objFunc := range objFuncs {
		o := objFunc()
		v := reflect.ValueOf(o)
		i := reflect.Indirect(v)
		s := i.Type()
		names = append(names, s.Name())
	}
	return names
}
