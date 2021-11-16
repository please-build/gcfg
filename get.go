package gcfg

import (
	"fmt"
	"reflect"
)

// Get retrieves the values of a config field.
func Get(config interface{}, section, subsection, name string) ([]string, error) {
	field, err := parseField(section, subsection, name)
	if err != nil {
		return nil, err
	}

	configPtr := reflect.ValueOf(config)
	if configPtr.Kind() != reflect.Ptr || configPtr.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("Config must be a pointer to a struct")
	}
	configValue := configPtr.Elem()

	sectionValue, _ := fieldFold(configValue, section)
	if !sectionValue.IsValid() {
		return nil, fmt.Errorf("Section does not exist: %s", section)
	}

	if subsection != "" {
		if sectionValue.Kind() != reflect.Map || sectionValue.Type().Key().Kind() != reflect.String {
			return nil, fmt.Errorf("Subsection does not exist: %s", subsection)
		}

		if sectionValue.Type().Elem().Kind() == reflect.String {
			// We encode subsections and variables using the `subsection variable` format
			// in map keys, if we define subsections on the ini file where the underlying
			// type is map[string]string.
			key := reflect.ValueOf(subsection + " " + name)
			res := sectionValue.MapIndex(key)
			if !res.IsValid() {
				return nil, fmt.Errorf("Settable field not defined: %s", field)
			}
			return []string{res.String()}, nil
		}
		if sectionValue.Type().Elem().Kind() != reflect.Ptr || sectionValue.Type().Elem().Elem().Kind() != reflect.Struct {
			return nil, fmt.Errorf("Invalid unsettable field: %s", field)
		}
		res := sectionValue.MapIndex(reflect.ValueOf(subsection))
		if !res.IsValid() {
			return nil, fmt.Errorf("Settable field not defined: %s", field)
		}
		sectionValue = res.Elem()
	} else if sectionValue.Kind() == reflect.Map && sectionValue.Type().Key().Kind() == reflect.String && sectionValue.Type().Elem().Kind() == reflect.String {
		res := sectionValue.MapIndex(reflect.ValueOf(name))
		if !res.IsValid() {
			return nil, fmt.Errorf("Settable field not defined: %s", field)
		}
		return []string{res.String()}, nil
	} else if sectionValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("Invalid unsettable field: %s", field)
	}

	variableValue, _ := fieldFold(sectionValue, name)
	if !variableValue.IsValid() {
		var err error
		if variableValue, err = getExtraData(sectionValue, name, field); err != nil {
			return nil, err
		}
	}

	// If the field is a slice.
	if isMultiVal(variableValue) {
		m := make([]string, 0, variableValue.Len())
		for i := 0; i < variableValue.Len(); i++ {
			res, err := iniValue(variableValue.Index(i))
			if err != nil {
				return nil, fmt.Errorf("Failed to retrieve field %s: %s", field, err)
			}
			m = append(m, res)
		}
		return m, nil
	}

	res, err := iniValue(variableValue)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve field %s: %s", field, err)
	}
	return []string{res}, nil
}

// Tries to obtain `key` through the `gcfg:"extra_values"` functionality.
func getExtraData(sectionValue reflect.Value, key, field string) (reflect.Value, error) {
	extraDataValue := findExtraDataField(sectionValue)
	if extraDataValue == nil {
		return reflect.Value{}, fmt.Errorf("Invalid unsettable field: %s", field)
	}
	if extraDataValue.Kind() == reflect.Map && extraDataValue.Type().Key().Kind() == reflect.String {
		if extraDataValue.Type().Elem().Kind() == reflect.String ||
			extraDataValue.Type().Elem().Kind() == reflect.Slice && extraDataValue.Type().Elem().Elem().Kind() == reflect.String {
			res := extraDataValue.MapIndex(reflect.ValueOf(key))
			if !res.IsValid() {
				return reflect.Value{}, fmt.Errorf("Settable field not defined: %s", field)
			}
			return res, nil
		}
	}
	return reflect.Value{}, fmt.Errorf("Invalid unsettable field: %s", field)
}

func parseField(section, subsection, name string) (string, error) {
	if section == "" || name == "" {
		return "", fmt.Errorf("Both section and name must be non-empty")
	}

	field := section
	if subsection != "" {
		field += "." + subsection
	}
	return field + "." + name, nil
}
