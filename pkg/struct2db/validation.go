package struct2db

import (
	validator "github.com/mikolajgs/struct-validator"
)

// Validate checks object's fields. It returns result of validation as a bool and list of fields with invalid value
func (c Controller) Validate(obj interface{}, filters map[string]interface{}) (bool, map[string]int, error) {
	if filters != nil {
		valid, failedFields := validator.Validate(obj, &validator.ValidationOptions{
			ValidateWhenSuffix:   true,
			OverwriteFieldValues: filters,
			RestrictFields:       c.mapWithInterfacesToMapBool(filters),
			OverwriteTagName:     c.tagName,
		})
		return valid, failedFields, nil
	}

	valid, failedFields := validator.Validate(obj, &validator.ValidationOptions{
		ValidateWhenSuffix: true,
		OverwriteTagName:     c.tagName,
	})
	return valid, failedFields, nil
}
