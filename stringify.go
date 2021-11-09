package gcfg

import (
	"encoding"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Stringify returns the ini format representation of `config`.
func Stringify(config interface{}) (string, error) {
	configPtr := reflect.ValueOf(config)
	if configPtr.Kind() != reflect.Ptr || configPtr.Elem().Kind() != reflect.Struct {
		return "", fmt.Errorf("Config must be a pointer to a struct")
	}
	configValue := configPtr.Elem()

	var s string
	for i := 0; i < configValue.NumField(); i++ {
		fieldValue := configValue.Field(i)
		fieldStruct := configValue.Type().Field(i)
		if !fieldStruct.IsExported() {
			continue
		}

		iniFieldName := iniKey(fieldStruct)

		if fieldValue.Kind() == reflect.Struct {
			res, err := stringifyStructFields(fieldValue)
			if err != nil {
				return "", err
			}

			s += fmt.Sprintf("[%s]\n%s\n", iniFieldName, res)
		} else if fieldValue.Kind() == reflect.Map {
			if fieldStruct.Type.Key().Kind() != reflect.String {
				return "", fmt.Errorf("The map keys must to be of string type, instead they are of %s type", fieldStruct.Type.Key().Kind())
			}

			if fieldStruct.Type.Elem().Kind() == reflect.String {
				// We encode subsections and variables using the `subsection variable` format
				// in map keys, if we define subsections on the ini file where the underlying
				// type is map[string]string. We need to decode this from the config object
				// to convert it back to the ini format.
				mapTree := make(map[string]map[string]string)
				iter := fieldValue.MapRange()
				for iter.Next() {
					parts := strings.Split(iter.Key().String(), " ")
					if len(parts) > 2 {
						panic("Invalid map key value")
					}

					var subsection string
					var variable string
					if len(parts) == 1 {
						variable = parts[0]
					} else {
						subsection = parts[0]
						variable = parts[1]
					}
					if _, exists := mapTree[subsection]; !exists {
						mapTree[subsection] = make(map[string]string)
					}
					mapTree[subsection][variable] = iter.Value().String()
				}

				for subsection, variables := range mapTree {
					if subsection == "" {
						s += fmt.Sprintf("[%s]\n", iniFieldName)
					} else {
						s += fmt.Sprintf("[%s \"%s\"]\n", iniFieldName, subsection)
					}
					for variable, value := range variables {
						s += fmt.Sprintf("%s=%s\n", variable, value)
					}
					s += "\n"
				}
			} else if fieldStruct.Type.Elem().Kind() == reflect.Ptr && fieldStruct.Type.Elem().Elem().Kind() == reflect.Struct {
				iter := fieldValue.MapRange()
				for iter.Next() {
					res, err := stringifyStructFields(iter.Value().Elem())
					if err != nil {
						return "", err
					}

					if iter.Key().String() == "" {
						s += fmt.Sprintf("[%s]\n", iniFieldName)
					} else {
						s += fmt.Sprintf("[%s \"%s\"]\n", iniFieldName, iter.Key())
					}
					s += res + "\n"
				}
			} else {
				return "", fmt.Errorf("The map values must either be of string or *struct type, instead they are of %s type", fieldStruct.Type.Elem().Kind())
			}
		}
	}

	return s, nil
}

func stringifyStructFields(value reflect.Value) (string, error) {
	if value.Kind() != reflect.Struct {
		return "", fmt.Errorf("Expected a struct type from %v, but instead got %s\n", value, value.Kind())
	}

	var s string
	for i := 0; i < value.NumField(); i++ {
		fieldValue := value.Field(i)
		fieldStruct := value.Type().Field(i)
		if !fieldStruct.IsExported() {
			continue
		}

		if fieldStruct.Tag.Get("gcfg") == "extra_values" {
			if fieldStruct.Type != reflect.TypeOf(map[string]string{}) && fieldStruct.Type != reflect.TypeOf(map[string][]string{}) {
				return "", fmt.Errorf("Expected either a map[string]string or map[string][]string type, but instead got %s\n", fieldStruct.Type)
			}

			iter := fieldValue.MapRange()
			for iter.Next() {
				key := iter.Key()
				value := iter.Value()
				if value.Kind() == reflect.String {
					s += fmt.Sprintf("%s=%s\n", key, value)
				} else {
					for j := 0; j < value.Len(); j++ {
						s += fmt.Sprintf("%s=%s\n", key, value.Index(j))
					}
				}
			}
		} else {
			if err := iterateMaybeSlice(fieldValue, func(innerValue reflect.Value) error {
				res, err := iniValue(innerValue)
				if err != nil {
					return err
				}
				s += fmt.Sprintf("%s=%s\n", iniKey(fieldStruct), res)
				return nil
			}); err != nil {
				return "", err
			}
		}
	}

	return s, nil
}

// This function implements the inversion of `fieldFold`.
func iniKey(fieldStruct reflect.StructField) string {
	tag := newTag(fieldStruct.Tag.Get("gcfg"))
	if tag.ident != "" {
		return tag.ident
	}

	name := fieldStruct.Name
	firstRune, firstRuneSize := utf8.DecodeRuneInString(name)
	if firstRune == 'X' {
		secondRune, _ := utf8.DecodeRuneInString(name[firstRuneSize:])
		if unicode.IsLetter(secondRune) && !unicode.IsLower(secondRune) && !unicode.IsUpper(secondRune) {
			name = name[firstRuneSize:]
		}
	}

	return strings.ToLower(strings.Replace(name, "_", "-", -1))
}

// This function implements the inversion of `setters`.
func iniValue(value reflect.Value) (string, error) {
	if bigIntValue, ok := value.Interface().(big.Int); ok {
		return bigIntValue.String(), nil
	}

	if v, exists := value.Interface().(encoding.TextMarshaler); exists {
		res, err := v.MarshalText()
		if err != nil {
			return "", fmt.Errorf("Failed to marshal text: %s", err)
		}

		return string(res), nil
	}

	switch value.Kind() {
	case reflect.String:
		return value.String(), nil

	case reflect.Bool:
		if value.Bool() {
			return "true", nil
		}
		return "false", nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(value.Int(), 10), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(value.Uint(), 10), nil
	}

	return "", fmt.Errorf("Unable to stringify value: %+v", value.Interface())
}

func iterateMaybeSlice(value reflect.Value, callback func(reflect.Value) error) error {
	if value.Kind() == reflect.Slice {
		for i := 0; i < value.Len(); i++ {
			if err := callback(value.Index(i)); err != nil {
				return err
			}
		}
	} else {
		if err := callback(value); err != nil {
			return err
		}
	}
	return nil
}
