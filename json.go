package gcfg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

// RawJSON returns the config raw encoded JSON value.
// This value implements both Marshaler and Unmarshaler interfaces.
func RawJSON(config interface{}) ([]byte, error) {
	configPtr := reflect.ValueOf(config)
	if configPtr.Kind() != reflect.Ptr || configPtr.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("Config must be a pointer to a struct")
	}
	configValue := configPtr.Elem()

	res := []byte("{")
	for i := 0; i < configValue.NumField(); i++ {
		fieldValue := configValue.Field(i)
		fieldStruct := configValue.Type().Field(i)
		if !fieldValue.CanInterface() {
			continue
		}

		iniFieldName := iniKey(fieldStruct)

		if fieldValue.Kind() == reflect.Struct {
			innerRes, err := marshalStruct(fieldValue)
			if err != nil {
				return nil, err
			}
			res = append(res, fmt.Sprintf("\"%s\":%s,", iniFieldName, innerRes)...)
		} else if fieldValue.Kind() == reflect.Map {
			if fieldStruct.Type.Key().Kind() != reflect.String {
				return nil, fmt.Errorf("The map keys must to be of string type, instead they are of %s type", fieldStruct.Type.Key().Kind())
			}

			if fieldStruct.Type.Elem().Kind() == reflect.String {
				innerRes, err := json.Marshal(fieldValue.Interface())
				if err != nil {
					return nil, err
				}
				res = append(res, fmt.Sprintf("\"%s\":%s,", iniFieldName, innerRes)...)
			} else if fieldStruct.Type.Elem().Kind() == reflect.Ptr && fieldStruct.Type.Elem().Elem().Kind() == reflect.Struct {
				res = append(res, `"`+iniFieldName+`":{`...)
				iter := fieldValue.MapRange()
				for iter.Next() {
					innerRes, err := marshalStruct(iter.Value().Elem())
					if err != nil {
						return nil, err
					}
					res = append(res, fmt.Sprintf("\"%s\":%s,", iter.Key(), innerRes)...)
				}
				res = append(bytes.TrimSuffix(res, []byte(",")), `},`...)
			} else {
				return nil, fmt.Errorf("The map values must either be of string or *struct type, instead they are of %s type", fieldStruct.Type.Elem().Kind())
			}
		}
	}
	res = append(bytes.TrimSuffix(res, []byte(",")), '}')

	return json.RawMessage(res), nil
}

func marshalStruct(fieldValue reflect.Value) ([]byte, error) {
	if fieldValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("Expected a struct type from %v, but instead got %s\n", fieldValue, fieldValue.Kind())
	}

	res := []byte("{")
	for i := 0; i < fieldValue.NumField(); i++ {
		subfieldValue := fieldValue.Field(i)
		subfieldStruct := fieldValue.Type().Field(i)
		if !subfieldValue.CanInterface() {
			continue
		}

		if subfieldStruct.Tag.Get("gcfg") == "extra_values" {
			if subfieldStruct.Type != reflect.TypeOf(map[string]string{}) && subfieldStruct.Type != reflect.TypeOf(map[string][]string{}) {
				return nil, fmt.Errorf("Expected either a map[string]string or map[string][]string type, but instead got %s\n", subfieldStruct.Type)
			}

			iter := subfieldValue.MapRange()
			for iter.Next() {
				innerRes, err := json.Marshal(iter.Value().Interface())
				if err != nil {
					return nil, err
				}
				res = append(res, fmt.Sprintf("\"%s\":%s,", iter.Key(), innerRes)...)
			}
			res = bytes.TrimSuffix(res, []byte(","))
		} else {
			innerRes, err := json.Marshal(subfieldValue.Interface())
			if err != nil {
				return nil, err
			}
			res = append(res, fmt.Sprintf("\"%s\":%s", iniKey(subfieldStruct), innerRes)...)
		}

		res = append(res, ',')
	}

	return append(bytes.TrimSuffix(res, []byte(",")), '}'), nil
}
