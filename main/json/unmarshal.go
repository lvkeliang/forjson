package json

import (
	"container/list"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func SplitJson(json string) []string {
	rect := make([]string, 0, 10)
	stack := list.New()
	beginIndex := 0
	for i, r := range json {
		if r == rune('{') || r == rune('[') {
			stack.PushBack(struct{}{})
		} else if r == rune('}') || r == rune(']') {
			ele := stack.Back()
			if ele != nil {
				stack.Remove(ele)
			}
		} else if r == rune(',') {
			if stack.Len() == 0 {
				rect = append(rect, json[beginIndex:i])
				beginIndex = i + 1
			}
		}
	}
	rect = append(rect, json[beginIndex:])
	return rect
}

func JsonUnmarshal(data []byte, v any) error {
	str := string(data)
	str = strings.TrimLeft(str, " ")
	str = strings.TrimRight(str, " ")
	if len(str) == 0 {
		return nil
	}

	typ := reflect.TypeOf(v)
	value := reflect.ValueOf(v)
	if typ.Kind() != reflect.Ptr {
		return errors.New("must pass pointer parameter")
	}

	typ = typ.Elem()
	value = value.Elem()

	switch typ.Kind() {
	case reflect.String:
		if str[0] == '"' && str[len(str)-1] == '"' {
			value.SetString(str[1 : len(str)-1])
		} else {
			return fmt.Errorf("invalid json part: %s", str)
		}
	case reflect.Bool:
		if boo, err := strconv.ParseBool(str); err == nil {
			value.SetBool(boo)
		} else {
			return err
		}
	case reflect.Float32, reflect.Float64:
		if flt, err := strconv.ParseFloat(str, 64); err == nil {
			value.SetFloat(flt)
		} else if flt, err := strconv.ParseFloat(str, 32); err == nil {
			value.SetFloat(flt)
		} else {
			return err
		}
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		if str[0] == '"' && str[len(str)-1] == '"' {
			str = str[1 : len(str)-1]
		}
		if i, err := strconv.ParseInt(str, 10, 64); err != nil {
			return err
		} else {
			value.SetInt(i)
		}
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		if str[0] == '"' && str[len(str)-1] == '"' {
			str = str[1 : len(str)-1]
		}
		if i, err := strconv.ParseUint(str, 10, 64); err != nil {
			return err
		} else {
			value.SetUint(i)
		}
	case reflect.Map:
		if str[0] == '{' && str[len(str)-1] == '}' {
			if len(str) > 2 {
				arr := SplitJson(str[1 : len(str)-1]) //去除前后的{}
				if len(arr) > 0 {
					mapValue := reflect.ValueOf(v).Elem()                //别忘了，v是指针
					mapValue.Set(reflect.MakeMapWithSize(typ, len(arr))) //通过反射创建map
					kType := typ.Key()                                   //获取map的key的Type
					vType := typ.Elem()                                  //获取map的value的Type
					for i := 0; i < len(arr); i++ {
						brr := strings.SplitN(arr[i], ":", 2)
						if len(brr) != 2 {
							return fmt.Errorf("invalid json part: %s", arr[i])
						}

						kValue := reflect.New(kType) //根据Type创建指针型的Value
						if err := JsonUnmarshal([]byte(brr[0]), kValue.Interface()); err != nil {
							return err
						}
						vValue := reflect.New(vType) //根据Type创建指针型的Value
						if err := JsonUnmarshal([]byte(brr[1]), vValue.Interface()); err != nil {
							return err
						}
						mapValue.SetMapIndex(kValue.Elem(), vValue.Elem()) //往map里面赋值
					}
				}
			}
		} else if str != "null" {
			return fmt.Errorf("invalid json part: %s", str)
		}

	case reflect.Array, reflect.Slice:
		if str[0] == '[' && str[len(str)-1] == ']' {
			arr := SplitJson(str[1 : len(str)-1])
			if len(arr) > 0 {
				slice := reflect.ValueOf(v).Elem()
				slice.Set(reflect.MakeSlice(typ, len(arr), len(arr))) //通过反射创建slice
				for i := 0; i < len(arr); i++ {
					eleValue := slice.Index(i)
					eleType := eleValue.Type()
					if eleType.Kind() != reflect.Ptr {
						eleValue = eleValue.Addr()
					}
					if err := JsonUnmarshal([]byte(arr[i]), eleValue.Interface()); err != nil {
						return err
					}
				}
			}
		} else if str != "null" {
			return fmt.Errorf("invalid json part: %s", str)
		}

	case reflect.Struct:
		if str[0] == '{' && str[len(str)-1] == '}' {
			if len(str) > 2 {
				arr := SplitJson(str[1 : len(str)-1])
				fieldCount := typ.NumField()
				//建立json tag到FieldName的映射关系
				tag2Field := make(map[string]string, fieldCount)
				for i := 0; i < fieldCount; i++ {
					fieldType := typ.Field(i)
					name := fieldType.Name
					if len(fieldType.Tag.Get(tagName)) > 0 {
						name = fieldType.Tag.Get(tagName)
					}
					tag2Field[name] = fieldType.Name
				}

				for _, fieldString := range arr {
					keyValue := strings.SplitN(fieldString, ":", 2)
					if len(keyValue) == 2 {
						tag := strings.Trim(keyValue[0], " ")
						if tag[0] == '"' && tag[len(tag)-1] == '"' {
							tag = tag[1 : len(tag)-1]
							if fieldName, exists := tag2Field[tag]; exists {
								fieldValue := value.FieldByName(fieldName)
								fieldType := fieldValue.Type()
								if fieldType.Kind() != reflect.Ptr {
									//如果内嵌不是指针，则声明时已经用0值初始化了，此处只需要根据json改写它的值
									fieldValue = fieldValue.Addr() //确保fieldValue指向指针类型，因为接下来要把fieldValue传给Unmarshal
									if err := JsonUnmarshal([]byte(keyValue[1]), fieldValue.Interface()); err != nil {
										return err
									}
								} else {
									//如果内嵌的是指针，则需要通过New()创建一个实例(申请内存空间)。不能给New()传指针型的Type，所以调一下Elem()
									newValue := reflect.New(fieldType.Elem()) //newValue代表的是指针
									if err := JsonUnmarshal([]byte(keyValue[1]), newValue.Interface()); err != nil {
										return err
									}
									value.FieldByName(fieldName).Set(newValue) //把newValue赋给value的Field
								}

							} else {
								fmt.Printf("cannot find field: %s\n", tag)
							}
						} else {
							return fmt.Errorf("invalid json part: %s", tag)
						}
					} else {
						return fmt.Errorf("invalid json part: %s", fieldString)
					}
				}
			}
		} else if str != "null" {
			return fmt.Errorf("invalid json part: %s", str)
		}
		return nil

	default:
		return nil
	}
	return nil
}
