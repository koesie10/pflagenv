package pflagenv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"
)

func newStringMap(val map[string]string, ref *map[string]string) *stringMap {
	sm := new(stringMap)
	sm.value = ref
	*sm.value = val
	return sm
}

type stringMap struct {
	value   *map[string]string
	changed bool
}

func (m *stringMap) Set(val string) error {
	s, err := readAsCSV(val)
	if err != nil {
		return err
	}
	v, err := readAsMap(s)
	if err != nil {
		return err
	}
	if !m.changed {
		*m.value = v
	} else {
		mapValue := *m.value

		for k, v := range v {
			mapValue[k] = v
		}

		*m.value = mapValue
	}
	m.changed = true
	return nil
}

func (m *stringMap) String() string {
	var vals []string
	for k, v := range *m.value {
		vals = append(vals, fmt.Sprintf("%s=%s", k, v))
	}

	str, _ := writeAsCSV(vals)
	return "[" + str + "]"
}

func (m *stringMap) Type() string {
	return "stringSlice"
}

func newInt64Map(val map[string]int64, ref *map[string]int64) *int64Map {
	sm := new(int64Map)
	sm.value = ref
	*sm.value = val
	return sm
}

type int64Map struct {
	value   *map[string]int64
	changed bool
}

func (m *int64Map) Set(val string) error {
	s, err := readAsCSV(val)
	if err != nil {
		return err
	}
	sv, err := readAsMap(s)
	if err != nil {
		return err
	}
	v, err := convertToInt64Map(sv)
	if err != nil {
		return err
	}
	if !m.changed {
		*m.value = v
	} else {
		mapValue := *m.value

		for k, v := range v {
			mapValue[k] = v
		}

		*m.value = mapValue
	}
	m.changed = true
	return nil
}

func (m *int64Map) String() string {
	var vals []string
	for k, v := range *m.value {
		vals = append(vals, fmt.Sprintf("%s=%d", k, v))
	}

	str, _ := writeAsCSV(vals)
	return "[" + str + "]"
}

func (m *int64Map) Type() string {
	return "stringSlice"
}

func convertToInt64Map(val map[string]string) (map[string]int64, error) {
	result := make(map[string]int64, len(val))

	for k, v := range val {
		intVal, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid int in %q: %w", v, err)
		}

		result[k] = intVal
	}

	return result, nil
}

func readAsMap(val []string) (map[string]string, error) {
	result := make(map[string]string, len(val))

	for _, v := range val {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid key=value format in %q", v)
		}

		result[parts[0]] = parts[1]
	}

	return result, nil
}

func readAsCSV(val string) ([]string, error) {
	if val == "" {
		return nil, nil
	}
	stringReader := strings.NewReader(val)
	csvReader := csv.NewReader(stringReader)
	return csvReader.Read()
}

func writeAsCSV(vals []string) (string, error) {
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)
	err := w.Write(vals)
	if err != nil {
		return "", err
	}
	w.Flush()
	return strings.TrimSuffix(b.String(), "\n"), nil
}

func StringMapHook() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if t.Kind() != reflect.Map {
			return data, nil
		}

		if t.Key().Kind() != reflect.String || t.Elem().Kind() != reflect.String {
			return data, nil
		}

		switch v := data.(type) {
		case []string:
			return readAsMap(v)
		case string:
			mapValues, err := readAsCSV(v)
			if err != nil {
				return nil, err
			}
			return readAsMap(mapValues)
		case map[string]string:
			return v, nil
		default:
			return nil, fmt.Errorf("unsupported map type %T", data)
		}
	}
}

func Int64MapHook() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if t.Kind() != reflect.Map {
			return data, nil
		}

		if t.Key().Kind() != reflect.String || t.Elem().Kind() != reflect.Int64 {
			return data, nil
		}

		switch v := data.(type) {
		case []string:
			mapValues, err := readAsMap(v)
			if err != nil {
				return nil, err
			}
			return convertToInt64Map(mapValues)
		case string:
			mapValues, err := readAsCSV(v)
			if err != nil {
				return nil, err
			}
			values, err := readAsMap(mapValues)
			if err != nil {
				return nil, err
			}
			return convertToInt64Map(values)
		case map[string]int64:
			return v, nil
		default:
			return nil, fmt.Errorf("unsupported map type %T", data)
		}
	}
}
