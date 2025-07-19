package main

import (
	"fmt"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {
	outValue := reflect.ValueOf(out)

	if outValue.Kind() != reflect.Ptr {
		return fmt.Errorf("out is not a pointer")
	} else {
		outValue = outValue.Elem()
	}

	switch outValue.Kind() {
	case reflect.Struct:
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return fmt.Errorf("failed convert to map[string]interface{}")
		}

		for i := 0; i < outValue.NumField(); i++ {
			fieldName := outValue.Type().
				Field(i).
				Name

			value, ok := dataMap[fieldName]
			if !ok {
				return fmt.Errorf("field not found: %s", fieldName)
			}

			var err = i2s(
				value,
				outValue.Field(i).
					Addr().
					Interface(),
			)

			if err != nil {
				return fmt.Errorf("failed to process struct field %s: %s", fieldName, err)
			}
		}

	case reflect.Slice:
		dataSlice, ok := data.([]interface{})
		if !ok {
			return fmt.Errorf("failed convert to []interface{}")
		}

		for idx, value := range dataSlice {
			outData := reflect.New(outValue.Type().Elem())

			var err = i2s(value, outData.Interface())
			if err != nil {
				return fmt.Errorf("failed to process slice element %dataMap: %s", idx, err)
			}

			outValue.Set(reflect.Append(outValue, outData.Elem()))
		}

	case reflect.Int:
		floatData, ok := data.(float64)
		if !ok {
			return fmt.Errorf("failed convert to float64")
		}

		outValue.SetInt(int64(floatData))

	case reflect.String:
		stringData, ok := data.(string)
		if !ok {
			return fmt.Errorf("failed convert to string")
		}

		outValue.SetString(stringData)

	case reflect.Bool:
		boolData, ok := data.(bool)
		if !ok {
			return fmt.Errorf("failed convert to bool")
		}

		outValue.SetBool(boolData)

	default:
		return fmt.Errorf("unsupportd type: %s", outValue.Kind())
	}

	return nil
}
