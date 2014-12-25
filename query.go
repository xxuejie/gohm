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

func (q query) Model(v interface{}) query {
	q.ModelType = fetchTypeFromReturnInterface(v)
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

func (q query) Set() (Set, error) {
	if q.ModelType == nil {
		return Set{}, ModelTypeUnknownError
	}
	modelName := q.ModelType.Name()
	filters, err := q.filters()
	if err != nil {
		return Set{}, err
	}
	switch len(filters) {
	case 0:
		return NewSet(q.G, connectKeys(modelName, "all"), modelName), nil
	case 1:
		return NewSet(q.G, filters[0], modelName), nil
	default:
		return Set{}, NotImplementedError
	}
}

func (q query) FetchByIds(v interface{}, ids []interface{}) error {
	set, err := q.Model(v).Set()
	if err != nil {
		return err
	}
	return set.FetchByIds(v, ids)
}

func (q query) Fetch(v interface{}) error {
	set, err := q.Model(v).Set()
	if err != nil {
		return err
	}
	return set.Fetch(v)
}

func (q query) Size() (int, error) {
	set, err := q.Set()
	if err != nil {
		return 0, err
	}
	return set.Size()
}

func (q query) Exists(id interface{}) (bool, error) {
	set, err := q.Set()
	if err != nil {
		return false, err
	}
	return set.Exists(id)
}

func (q query) Include(v interface{}) (bool, error) {
	set, err := q.Model(v).Set()
	if err != nil {
		return false, err
	}
	return set.Include(v)
}
