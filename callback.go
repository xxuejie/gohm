package gohm

import (
	"fmt"
	"reflect"
)

type CallbackFunc func(model interface{}, attrs *map[string]string)

func typeFullName(modelType reflect.Type) string {
	return fmt.Sprintf("%s.%s", modelType.PkgPath(), modelType.Name())
}

func (g *Gohm) fetchSaveCallbacksFromMap(modelType reflect.Type) []CallbackFunc {
	fullName := typeFullName(modelType)
	callbacks, ok := g.Callbacks[fullName]
	if !ok {
		return nil
	}
	return callbacks
}

func pickData(attr string, index int, callbackAttrs map[string]string, modelData reflect.Value) string {
	val, ok := callbackAttrs[attr]
	if ok {
		return val
	}
	return modelData.Field(index).String()
}
