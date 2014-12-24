package gohm

import (
	"reflect"
)

type query struct {
	G       *Gohm
	ModelType reflect.Type
	Queries map[string]interface{}
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

func (q query) filters() ([]string, error) {
	modelName := q.ModelType.Name()
	ret := make([]string, 0)
	for k, v := range q.Queries {
		// TODO: right now we only support one query value per key
		if !modelHasIndex(q.ModelType, k) {
			return nil, IndexNotFoundError
		}
		ret = append(ret, connectKeys(modelName, "indices", k, v))
	}
	return ret, nil
}

func (q query) FetchByIds(v interface{}, ids []interface{}) error {
	q.ModelType = fetchTypeFromReturnInterface(v)
	modelName := q.ModelType.Name()
	return NewSet(q.G, connectKeys(modelName, "all"), modelName).FetchByIds(v, ids)
}

func (q query) Fetch(v interface{}) error {
	q.ModelType = fetchTypeFromReturnInterface(v)
	modelName := q.ModelType.Name()
	filters, err := q.filters()
	if err != nil {
		return err
	}
	switch len(filters) {
	case 0:
		return NewSet(q.G, connectKeys(modelName, "all"), modelName).Fetch(v)
	case 1:
		return NewSet(q.G, filters[0], modelName).Fetch(v)
	default:
		return NotImplementedError
	}
}
