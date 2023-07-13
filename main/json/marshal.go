package json

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const tagName = "name"

func isSupport(kind reflect.Kind) bool {
	switch kind {
	case reflect.Chan, reflect.Complex128, reflect.Complex64, reflect.Func, reflect.Invalid:
		return false
	default:
		return true
	}
}

func parseParam(obj any) (jsonString string, err error) {
	// 对空接口返回null
	if obj == nil {
		return "null", nil
	}

	objKind := reflect.TypeOf(obj).Kind()

	if !isSupport(objKind) {
		return "", errors.New("this type is unsupported")
	}

	switch objKind {
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Float32,
		reflect.Float64:
		return fmt.Sprintf("%v", reflect.ValueOf(obj)), nil

	case reflect.String:
		return fmt.Sprintf("\"%v\"", reflect.ValueOf(obj)), nil

	case reflect.Map:
		var resultList []string

		for _, mapKey := range reflect.ValueOf(obj).MapKeys() {
			key := mapKey.Interface()
			value := reflect.ValueOf(obj).MapIndex(mapKey)

			if !value.CanInterface() || !isSupport(reflect.TypeOf(value).Kind()) {
				continue
			}

			result, err := parseParam(value.Interface())
			if err != nil {
				return "", err
			}

			resultList = append(resultList, fmt.Sprintf("\"%v\":%v", key, result))
		}

		return fmt.Sprintf("{%v}", strings.Join(resultList, ",")), nil

	case reflect.Array, reflect.Slice:
		var values = reflect.ValueOf(obj)
		length := values.Len()
		result := make([]string, length)

		for i := 0; i < length; i++ {
			result[i], err = parseParam(values.Index(i).Interface())
			// fmt.Println(parseParam(values.Index(i)))
			if err != nil {
				return "", err
			}
		}

		return fmt.Sprintf("[%v]", strings.Join(result, ",")), nil
	case reflect.Struct:

		fields := reflect.ValueOf(obj)
		fieldNum := fields.NumField()

		var resultList []string

		for i := 0; i < fieldNum; i++ {
			field := fields.Field(i)
			fieldType := fields.Type()

			if !field.CanInterface() || !isSupport(reflect.TypeOf(field).Kind()) {
				continue
			}

			key := fieldType.Field(i).Name
			tag := fieldType.Field(i).Tag.Get(tagName)
			fmt.Println("tag:" + tag)

			if tag != "" {
				key = tag
			}

			result, err := parseParam(field.Interface())
			if err != nil {
				return "", err
			}

			resultList = append(resultList, fmt.Sprintf("\"%v\":%v", key, result))
		}
		return fmt.Sprintf("{%v}", strings.Join(resultList, ",")), nil
	default:
		return "", nil
	}
}

func JsonMarshal(obj any) ([]byte, error) {
	result, err := parseParam(obj)
	if err != nil {
		return nil, err
	}
	return []byte(result), err
}
