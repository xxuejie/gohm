package gohm

import (
	"fmt"
	"reflect"
)

func fetchTypeFromReturnInterface(out interface{}) reflect.Type {
	v := reflect.ValueOf(out)
	switch v.Kind() {
	case reflect.Ptr:
		return fetchTypeFromReturnInterface(reflect.Indirect(v).Interface())
	case reflect.Slice:
		return fetchTypeFromReturnInterface(reflect.New(v.Type().Elem()).Interface())
	default:
		return v.Type()
	}
}

func fetchTypeNameFromReturnInterface(out interface{}) string {
	return fetchTypeFromReturnInterface(out).Name()
}

// Inspired from https://github.com/eaigner/jet/blob/master/mapper.go
func fillResponse(resp [][]string, out interface{}) error {
	v := reflect.ValueOf(out)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("Cannot unpack result to non-pointer!")
	}
	for i := range resp {
		err := fillValue(resp[i], v)
		if err != nil {
			return err
		}
	}
	return nil
}

func fillValue(resp []string, out reflect.Value) error {
	switch out.Kind() {
	case reflect.Ptr:
		fmt.Printf("PTR:\n")
		if out.IsNil() {
			out.Set(reflect.New(out.Type().Elem()))
		}
		return fillValue(resp, reflect.Indirect(out))
	case reflect.Slice:
		return fillSlice(resp, out)
	case reflect.Struct:
		return fillStruct(resp, out)
	case reflect.Map:
		return fillMap(resp, out)
	}
	return fmt.Errorf("Type %T (%s) is not supported in response!", out, out.Kind())
}

func fillSlice(resp []string, out reflect.Value) error {
	elem := reflect.Indirect(reflect.New(out.Type().Elem()))
	err := fillValue(resp, elem)
	if err != nil {
		return err
	}
	out.Set(reflect.Append(out, elem))
	return nil
}

func fillStruct(resp []string, out reflect.Value) error {
	// we would check if the model is valid outside
	modelLoadAttrs(resp, out.Addr().Interface())
	return nil
}

func fillMap(resp []string, out reflect.Value) error {
	if out.IsNil() {
		out.Set(reflect.MakeMap(out.Type()))
	}
	for i := 0; i < len(resp); i = i + 2 {
		key := resp[i]
		val := resp[i+1]
		out.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(val))
	}
	return nil
}
