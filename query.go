package gohm

import (
	"reflect"
)

type query struct {
	G          *Gohm
	ValueModel interface{}
	Queries    map[string]interface{}
}

func (g *Gohm) All() query {
	return query{
		G:       g,
		Queries: make(map[string]interface{}),
	}
}

func (q query) Find(k string, v interface{}) query {
	q.Queries[k] = v
	return q
}

func (q query) Model(v interface{}) query {
	q.ValueModel = v
	return q
}

func (q query) filters(modelType reflect.Type) ([]string, error) {
	modelName := modelType.Name()
	ret := make([]string, 0)
	for k, v := range q.Queries {
		// TODO: right now we only support one query value per key
		if !modelHasIndex(modelType, k) {
			return nil, IndexNotFoundError
		}
		ret = append(ret, connectKeys(modelName, "indices", k, v))
	}
	return ret, nil
}

func (q query) Set() (BasicSet, error) {
	if q.ValueModel == nil {
		return nil, ModelTypeUnknownError
	}
	// TODO: Add validation here once we figure out how to deal with slice
	// if err := validateModel(q.ValueModel); err != nil {
	// 	return Set{}, err
	// }
	modelType := fetchTypeFromReturnInterface(q.ValueModel)
	modelName := modelType.Name()
	filters, err := q.filters(modelType)
	if err != nil {
		return Set{}, err
	}
	switch len(filters) {
	case 0:
		return NewSet(q.G, connectKeys(modelName, "all"), modelName), nil
	case 1:
		return NewSet(q.G, filters[0], modelName), nil
	default:
		return NewMultiSet(q.G, modelName, NewCommand("sinterstore", filters...)), nil
	}
}

func (q query) FetchByIds(v interface{}, ids []interface{}) error {
	set, err := q.Model(v).Set()
	if err != nil {
		return err
	}
	return SetFetchByIds(set, v, ids)
}

func (q query) Fetch(v interface{}) error {
	set, err := q.Model(v).Set()
	if err != nil {
		return err
	}
	return SetFetch(set, v)
}

func (q query) Size() (int, error) {
	set, err := q.Set()
	if err != nil {
		return 0, err
	}
	return SetSize(set)
}

func (q query) Exists(id interface{}) (bool, error) {
	set, err := q.Set()
	if err != nil {
		return false, err
	}
	return SetExists(set, id)
}

func (q query) Include(v interface{}) (bool, error) {
	set, err := q.Model(v).Set()
	if err != nil {
		return false, err
	}
	return SetInclude(set, v)
}
