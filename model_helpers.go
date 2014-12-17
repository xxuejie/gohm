package gohm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var NoStructError error = errors.New(`model is not a struct`)
var NoIDError error = errors.New(`model does not have an ohm:"id" tagged field`)
var NonStringIDError error = errors.New(`model's ohm:"id" field is not a string`)
var NonExportedAttrError error = errors.New(`can't put ohm tags in unexported fields`)

// If you plan on calling any of the Model helpers available in this package
// make sure you always run ValidateModel on your model, or you run a pretty
// big risk of raising a panic: gohm uses *a lot* of reflection, which is very
// prone to panics when the type received doesn't follow certain assumptions.
func validateModel(model interface{}) error {
	var hasID bool
	modelData := reflect.ValueOf(model).Elem()
	modelType := modelData.Type()

	if modelData.Kind().String() != `struct` {
		return NoIDError
	}

	for i := 0; i < modelData.NumField(); i++ {
		if !modelData.Field(i).CanSet() {
			return NonExportedAttrError
		}

		if modelType.Field(i).Tag.Get(`ohm`) == `id` {
			hasID = true
		}

		if modelType.Field(i).Type.Name() != `string` {
			return NonStringIDError
		}
	}

	if !hasID {
		return NoIDError
	}

	return nil
}

func modelReflectValue(model interface{}) reflect.Value {
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
}

func modelReflectType(model interface{}) reflect.Type {
	return modelReflectValue(model).Type()
}

func modelAttrIndexMap(model interface{}) map[string]int {
	attrs := map[string]int{}
	typeData := modelReflectType(model)
	for i := 0; i < typeData.NumField(); i++ {
		field := typeData.Field(i)
		tag := strings.Split(field.Tag.Get(`ohm`), ` `)[0]
		if tag != `` && tag != `-` && tag != "id" {
			attrs[tag] = i
		}
	}

	return attrs
}

func modelIndexIndexMap(model interface{}) map[string]int {
	indices := map[string]int{}
	typeData := modelReflectType(model)
	for i := 0; i < typeData.NumField(); i++ {
		field := typeData.Field(i)
		tags := strings.Split(field.Tag.Get(`ohm`), ` `)
		key := tags[0]
		if key != `` && key != `-` && key != `id` {
			if stringInSlice(`index`, tags[1:]) {
				indices[key] = i
			}
		}
	}

	return indices
}

func modelUniqueIndexMap(model interface{}) map[string]int {
	uniques := map[string]int{}
	typeData := modelReflectType(model)
	for i := 0; i < typeData.NumField(); i++ {
		field := typeData.Field(i)
		tags := strings.Split(field.Tag.Get(`ohm`), ` `)
		key := tags[0]
		if key != `` && key != `-` && key != `id` {
			if stringInSlice(`unique`, tags[1:]) {
				uniques[key] = i
			}
		}
	}

	return uniques
}

func modelKey(model interface{}) (key string) {
	key = fmt.Sprintf("%v:%v", modelType(model), modelID(model))

	return
}

func modelID(model interface{}) (id string) {
	modelData := modelReflectValue(model)
	idFieldName := modelIDFieldName(model)
	id = modelData.FieldByName(idFieldName).String()

	return
}

func modelHasAttribute(model interface{}, attribute string) bool {
	attrIndexMap := modelAttrIndexMap(model)
	for attr, _ := range attrIndexMap {
		if attribute == attr {
			return true
		}
	}

	return false
}

func modelIDFieldName(model interface{}) (fieldName string) {
	modelData := modelReflectValue(model)
	modelType := modelData.Type()

	for i := 0; i < modelData.NumField(); i++ {
		field := modelType.Field(i)
		tag := field.Tag.Get(`ohm`)
		if tag == `id` {
			fieldName = field.Name
			break
		}
	}

	return
}

func modelSetID(id string, model interface{}) {
	modelReflectValue(model).FieldByName(modelIDFieldName(model)).SetString(id)
}

func modelType(model interface{}) string {
	return modelReflectType(model).Name()
}

func modelLoadAttrs(attrs []string, model interface{}) {
	modelData := modelReflectValue(model)
	modelType := modelData.Type()
	fmt.Printf("Load attrs of type: %s\n", modelType.Name())
	attrIndexMap := modelAttrIndexMap(model)
	for i := 0; i < len(attrs); i = i + 2 {
		attrName := attrs[i]
		attrValue := attrs[i+1]
		attrIndex := attrIndexMap[attrName]

		if attrName == "id" {
			modelSetID(attrValue, model)
		} else if modelHasAttribute(model, attrName) {
			attrValueValue := reflect.ValueOf(attrValue)
			typedAttrValue := attrValueValue.Convert(modelType.Field(attrIndex).Type)
			modelData.Field(attrIndex).Set(typedAttrValue)
		}
	}
}
