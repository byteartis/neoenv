package neoenv

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const tagKey = "env"

// Load reads the environment variables and populates the struct defined in T.
// T must be a struct otherwise it will thrown an error.
// Each nested level within the struct is sepearated by '__'.
func Load[T any]() (*T, error) {
	in := new(T)

	v := reflect.ValueOf(in)
	el := v.Elem()
	if el.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected pointer to struct, got pointer to %T", in)
	}

	return in, parse("", el)
}

func parse(prefix string, v reflect.Value) error {
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := v.Type().Field(i)

		tagValue := fieldType.Tag.Get(tagKey)
		if tagValue == "" {
			tagValue = normalizeKey(fieldType.Name)
		}

		key := buildKey(prefix, tagValue)

		if field.Kind() == reflect.Struct {
			if err := parse(key, field); err != nil {
				return err
			}
		} else {
			v := getEnv(key)
			if v == "" {
				continue
			}

			if err := setField(field, v); err != nil {
				return err
			}
		}
	}

	return nil
}

func buildKey(prefix, key string) string {
	if prefix == "" {
		return key
	}

	return prefix + "__" + key
}

func getEnv(k string) string {
	return os.Getenv(strings.ToUpper(k))
}

func setField(field reflect.Value, value string) error {
	//nolint:exhaustive // we are handling all the types
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(v)
	case reflect.Slice:
		err := parseSlice(field, value)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported type %s", field.Kind())
	}

	return nil
}

var (
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

func normalizeKey(str string) string {
	normalizedStr := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	normalizedStr = matchAllCap.ReplaceAllString(normalizedStr, "${1}_${2}")
	return strings.ToLower(normalizedStr)
}

//revive:disable:cognitive-complexity // we are handling all the types we want
func parseSlice(field reflect.Value, value string) error {
	items := strings.Split(value, ",")
	size := len(items)
	s := reflect.MakeSlice(field.Type(), size, size)

	//nolint:exhaustive // we are handling all the types
	switch k := field.Type().Elem().Kind(); k {
	case reflect.String:
		for i, item := range items {
			s.Index(i).SetString(item)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		for i, item := range items {
			v, err := strconv.ParseInt(item, 10, 64)
			if err != nil {
				return err
			}
			s.Index(i).SetInt(v)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		for i, item := range items {
			v, err := strconv.ParseUint(item, 10, 64)
			if err != nil {
				return err
			}
			s.Index(i).SetUint(v)
		}
	case reflect.Float32, reflect.Float64:
		for i, item := range items {
			v, err := strconv.ParseFloat(item, 64)
			if err != nil {
				return err
			}
			s.Index(i).SetFloat(v)
		}
	default:
		return fmt.Errorf("unsupported slice type %s", k)
	}

	field.Set(s)

	return nil
}
