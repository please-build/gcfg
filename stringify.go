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
		if !fieldValue.CanInterface() {
			continue
		}

		iniFieldName := iniKey(fieldStruct)

		if fieldValue.Kind() == reflect.Struct {
			iniVariableLines, err := stringifyStructFields(fieldValue)
			if err != nil {
				return "", err
			}

			s += iniSectionLine(iniFieldName, "")
			s += iniVariableLines
			s += "\n"
		} else if fieldValue.Kind() == reflect.Map {
			if fieldStruct.Type.Key().Kind() != reflect.String {
				return "", fmt.Errorf("The map keys must to be of string type, instead they are of %s type", fieldStruct.Type.Key().Kind())
			}

			if fieldStruct.Type.Elem().Kind() == reflect.String {
				for subsection, variables := range decodeStringMap(fieldValue) {
					s += iniSectionLine(iniFieldName, subsection)
					for variable, value := range variables {
						s += iniVariableLine(variable, value)
					}
					s += "\n"
				}
			} else if fieldStruct.Type.Elem().Kind() == reflect.Ptr && fieldStruct.Type.Elem().Elem().Kind() == reflect.Struct {
				iter := fieldValue.MapRange()
				for iter.Next() {
					iniVariableLines, err := stringifyStructFields(iter.Value().Elem())
					if err != nil {
						return "", err
					}

					s += iniSectionLine(iniFieldName, iter.Key().String())
					s += iniVariableLines
					s += "\n"
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
		if !fieldValue.CanInterface() {
			continue
		}

		if fieldStruct.Tag.Get("gcfg") == "extra_values" {
			if fieldStruct.Type != reflect.TypeOf(map[string]string{}) && fieldStruct.Type != reflect.TypeOf(map[string][]string{}) {
				return "", fmt.Errorf("Expected either a map[string]string or map[string][]string type, but instead got %s\n", fieldStruct.Type)
			}

			iter := fieldValue.MapRange()
			for iter.Next() {
				iterateMaybeSlice(iter.Value(), func(innerValue reflect.Value) error {
					s += iniVariableLine(iter.Key().String(), innerValue.String())
					return nil
				})
			}
		} else {
			if err := iterateMaybeSlice(fieldValue, func(innerValue reflect.Value) error {
				res, err := iniValue(innerValue)
				if err != nil {
					return err
				}
				s += iniVariableLine(iniKey(fieldStruct), res)
				return nil
			}); err != nil {
				return "", err
			}
		}
	}

	return s, nil
}

func iniSectionLine(section, subsection string) string {
	if subsection == "" {
		return fmt.Sprintf("[%s]\n", section)
	}
	return fmt.Sprintf("[%s \"%s\"]\n", section, subsection)
}

func iniVariableLine(variable, value string) string {
	return fmt.Sprintf("%s = %s\n", variable, value)
}

// This function implements the inversion of `fieldFold` in `set.go`. The order of operations must be maintained.
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

// This function implements the inversion of `setters` in `set.go`. The order of operations must be maintained.
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

// We encode subsections and variables using the `subsection variable` format
// in map keys, if we define subsections on the ini file where the underlying
// type is map[string]string.
func decodeStringMap(value reflect.Value) map[string]map[string]string {
	if value.Kind() != reflect.Map || value.Type().Key().Kind() != reflect.String || value.Type().Elem().Kind() != reflect.String {
		panic(fmt.Sprintf("Value must be of map[string]string type, instead is was: %s", value.Type()))
	}

	mapTree := make(map[string]map[string]string)
	iter := value.MapRange()
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
	return mapTree
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
