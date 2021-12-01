package fsm

import (
	"errors"
	"reflect"
)

func GetTargetType(types []interface{}, expectedFields []string) (interface{}, error) {
	var result interface{} = nil
	for _, i := range types {
		// get current type fields list
		v := reflect.ValueOf(i)
		typeOfS := v.Type()
		currentTypeFields := make([]string, v.NumField())
		for i := 0; i < v.NumField(); i++ {
			currentTypeFields[i] = typeOfS.Field(i).Name
		}

		// compare current vs expected fields
		foundAllFields := true
		for _, expectedField := range expectedFields {
			foundCurrentField := false
			for _, currentTypeField := range currentTypeFields {
				if currentTypeField == expectedField {
					foundCurrentField = true
				}
			}
			if !foundCurrentField {
				foundAllFields = false
			}
		}

		if foundAllFields && len(expectedFields) == len(currentTypeFields) {
			result = i
			break
		}
	}

	if result == nil {
		return nil, errors.New("unknown type!")
	}

	return result, nil
}
