package gohm

import (
	"reflect"

	"github.com/garyburd/redigo/redis"
)

type MutableSet struct {
	Set Set
}

func NewMutableSet(g *Gohm, key string, namespace string, model reflect.Type) MutableSet {
	return MutableSet{
		Set: NewSet(g, key, namespace, model),
	}
}

func (set MutableSet) Lock() {
	set.Set.Lock()
}

func (set MutableSet) Unlock() {
	set.Set.Unlock()
}

func (set MutableSet) GetConn() redis.Conn {
	return set.Set.GetConn()
}

func (set MutableSet) FetchKey(conn redis.Conn) (string, error) {
	return set.Set.FetchKey(conn)
}

func (set MutableSet) Clean() error {
	return set.Set.Clean()
}

func (set MutableSet) GetNamespace() string {
	return set.Set.GetNamespace()
}

func (set MutableSet) AddById(id string) error {
	c, key, err := getConnAndKey(set)
	if err != nil {
		return err
	}
	defer cleanAndClose(set, c)
	_, err = c.Do("SADD", key, id)
	return err
}

func (set MutableSet) DeleteById(id string) error {
	c, key, err := getConnAndKey(set)
	if err != nil {
		return err
	}
	defer cleanAndClose(set, c)
	_, err = c.Do("SREM", key, id)
	return err
}

type MutableSetBuilder struct {
	G               *Gohm
	BaseModelValue  interface{}
	ValueModelValue interface{}
	NameValue       string
}

func (builder MutableSetBuilder) BaseModel(v interface{}) MutableSetBuilder {
	builder.BaseModelValue = v
	return builder
}

func (builder MutableSetBuilder) ValueModel(v interface{}) MutableSetBuilder {
	builder.ValueModelValue = v
	return builder
}

func (builder MutableSetBuilder) Name(name interface{}) MutableSetBuilder {
	builder.NameValue = toString(name)
	return builder
}

func (builder MutableSetBuilder) Set() (MutableSet, error) {
	if builder.ValueModelValue == nil || builder.BaseModelValue == nil {
		return MutableSet{}, ModelTypeUnknownError
	}
	valueModelType := fetchTypeFromReturnInterface(builder.ValueModelValue)
	if err := validateModel(reflect.New(valueModelType).Interface()); err != nil {
		return MutableSet{}, err
	}
	baseModelType := fetchTypeFromReturnInterface(builder.BaseModelValue)
	if err := validateModel(reflect.New(baseModelType).Interface()); err != nil {
		return MutableSet{}, err
	}
	baseModelName := baseModelType.Name()
	return NewMutableSet(builder.G, connectKeys(baseModelName, modelID(builder.BaseModelValue), builder.NameValue), baseModelName, baseModelType), nil
}

func (builder MutableSetBuilder) FetchByIds(v interface{}, ids []interface{}) error {
	set, err := builder.ValueModel(v).Set()
	if err != nil {
		return err
	}
	return SetFetchByIds(set, v, ids)
}

func (builder MutableSetBuilder) Fetch(v interface{}) error {
	set, err := builder.ValueModel(v).Set()
	if err != nil {
		return err
	}
	return SetFetch(set, v)
}

func (builder MutableSetBuilder) Ids() ([]string, error) {
	set, err := builder.Set()
	if err != nil {
		return nil, err
	}
	return SetIds(set)
}

func (builder MutableSetBuilder) Size() (int, error) {
	set, err := builder.Set()
	if err != nil {
		return 0, err
	}
	return SetSize(set)
}

func (builder MutableSetBuilder) Empty() (bool, error) {
	size, err := builder.Size()
	if err != nil {
		return false, err
	}
	return size == 0, nil
}

func (builder MutableSetBuilder) Exists(id interface{}) (bool, error) {
	set, err := builder.Set()
	if err != nil {
		return false, err
	}
	return SetExists(set, id)
}

func (builder MutableSetBuilder) Include(v interface{}) (bool, error) {
	set, err := builder.ValueModel(v).Set()
	if err != nil {
		return false, err
	}
	return SetInclude(set, v)
}

func (builder MutableSetBuilder) Add(v interface{}) error {
	return builder.ValueModel(v).AddById(modelID(v))
}

func (builder MutableSetBuilder) AddById(id interface{}) error {
	set, err := builder.Set()
	if err != nil {
		return err
	}
	return set.AddById(toString(id))
}

func (builder MutableSetBuilder) Delete(v interface{}) error {
	return builder.ValueModel(v).DeleteById(modelID(v))
}

func (builder MutableSetBuilder) DeleteById(id interface{}) error {
	set, err := builder.Set()
	if err != nil {
		return err
	}
	return set.DeleteById(toString(id))
}

func (g *Gohm) MutableSet(v interface{}, name interface{}) MutableSetBuilder {
	return MutableSetBuilder{
		G: g,
	}.BaseModel(v).Name(name)
}
