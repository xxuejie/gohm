package gohm

import (
	"errors"
	"github.com/garyburd/redigo/redis"
	"github.com/pote/go-msgpack"
	"github.com/pote/redisurl"
	"sync"
)

var (
	IndexNotFoundError = errors.New("Index is not found!")
	NotImplementedError = errors.New("Not implemented!")
	MissingIdError = errors.New("Missing ID!")
)

type Gohm struct {
	// Global lock in case we don't have per-type lock
	sync.Mutex

	RedisPool *redis.Pool
	LuaSave   *redis.Script
	LuaDelete *redis.Script
}

func NewGohm(r ...*redis.Pool) (*Gohm, error) {
	if len(r) < 1 {
		pool, err := redisurl.NewPool(3, 200, "240s")
		if err != nil {
			return &Gohm{}, err
		}

		return NewGohmWithPool(pool), nil
	} else {
		return NewGohmWithPool(r[0]), nil
	}
}

func NewGohmWithPool(pool *redis.Pool) *Gohm {
	g := &Gohm{
		RedisPool: pool,
	}

	g.LuaSave = redis.NewScript(0, LUA_SAVE)
	g.LuaDelete = redis.NewScript(0, LUA_DELETE)

	return g
}

func (g *Gohm) Save(model interface{}) error {
	if err := validateModel(model); err != nil {
		return err
	}

	modelData := modelReflectValue(model)
	modelType := modelData.Type()

	// Prepare Ohm-scripts `features` parameter.
	features := map[string]string{
		"name": modelType.Name(),
	}
	if modelID(model) != "" {
		features["id"] = modelID(model)
	}
	ohmFeatures, err := msgpack.Marshal(features)
	if err != nil {
		return err
	}

	// Prepare Ohm-scripts `attributes` parameter.
	attrs := []string{}
	attrIndexMap := modelAttrIndexMap(modelType)
	for attr, index := range attrIndexMap {
		attrs = append(attrs, attr)
		attrs = append(attrs, modelData.Field(index).String())
	}
	ohmAttrs, err := msgpack.Marshal(attrs)
	if err != nil {
		return err
	}

	// Prepare Ohm-scripts `indices` parameter.
	indices := map[string][]string{}
	indexIndexMap := modelIndexIndexMap(modelType)
	for attr, index := range indexIndexMap {
		val := modelData.Field(index).String()
		if len(val) > 0 {
			indices[attr] = []string{val}
		}
	}
	ohmIndices, err := msgpack.Marshal(indices)
	if err != nil {
		return err
	}

	// Prepare Ohm-scripts `uniques` parameter.
	uniques := map[string]string{}
	for attr, index := range modelUniqueIndexMap(modelType) {
		val := modelData.Field(index).String()
		if len(val) > 0 {
			uniques[attr] = val
		}
	}
	ohmUniques, err := msgpack.Marshal(uniques)
	if err != nil {
		return err
	}

	conn := g.RedisPool.Get()
	defer conn.Close()
	id, err := redis.String(g.LuaSave.Do(conn, ohmFeatures, ohmAttrs, ohmIndices, ohmUniques))
	if err != nil {
		return err
	}
	modelSetID(id, model)

	return nil
}

func (g *Gohm) Update(model interface{}, attrs map[string]interface{}) error {
	if err := validateModel(model); err != nil {
		return err
	}

	arr := make([]string, len(attrs) * 2)
	i := 0
	for k, v := range attrs {
		arr[i] = k
		i = i + 1
		arr[i] = toString(v)
		i = i + 1
	}

	modelLoadAttrs(arr, model)
	return g.Save(model)
}

func (g *Gohm) Delete(model interface{}) error {
	if err := validateModel(model); err != nil {
		return err
	}

	modelData := modelReflectValue(model)
	modelType := modelData.Type()
	modelId := modelID(model)
	if modelId == "" {
		return MissingIdError
	}

	// Prepare Ohm-scripts `features` parameter.
	features := map[string]string{
		"name": modelType.Name(),
		"id":   modelId,
		"key":  connectKeys(modelType.Name(), modelId),
	}
	ohmFeatures, err := msgpack.Marshal(features)
	if err != nil {
		return err
	}

	// Prepare Ohm-scripts `uniques` parameter.
	uniques := map[string]string{}
	for attr, index := range modelUniqueIndexMap(modelType) {
		val := modelData.Field(index).String()
		if len(val) > 0 {
			uniques[attr] = val
		}
	}

	ohmUniques, err := msgpack.Marshal(uniques)
	if err != nil {
		return err
	}

	// TODO: implements tracked
	ohmTracked, err := msgpack.Marshal([]string{})
	if err != nil {
		return err
	}

	conn := g.RedisPool.Get()
	defer conn.Close()
	id, err := redis.String(g.LuaDelete.Do(conn, ohmFeatures, ohmUniques, ohmTracked))
	if err != nil {
		return err
	}
	modelSetID(id, model)

	return nil
}

func (g *Gohm) FetchById(model interface{}, id interface{}) (bool, error) {
	if err := validateModel(model); err != nil {
		return false, err
	}

	conn := g.RedisPool.Get()
	defer conn.Close()

	attrs, err := redis.Strings(conn.Do(
		"HGETALL", connectKeys(modelType(model), id)))
	if err != nil || len(attrs) == 0 {
		return false, err
	}

	modelLoadAttrs(append(attrs, "id", toString(id)), model)
	return true, nil
}

func (g *Gohm) Counter(model interface{}, key string) (int64, error) {
	if err := validateModel(model); err != nil {
		return 0, err
	}
	id := modelID(model)
	if len(id) <= 0 {
		return 0, MissingIdError
	}

	conn := g.RedisPool.Get()
	defer conn.Close()

	resp, err := conn.Do(
		"HGET", connectKeys(modelType(model), id, "counters"), key)
	if resp == nil && err == nil {
		return 0, nil
	}
	return redis.Int64(resp, err)
}

func (g *Gohm) ClearCounter(model interface{}, key string) error {
	if err := validateModel(model); err != nil {
		return err
	}
	id := modelID(model)
	if len(id) <= 0 {
		return MissingIdError
	}

	conn := g.RedisPool.Get()
	defer conn.Close()

	_, err := conn.Do(
		"HDEL", connectKeys(modelType(model), id, "counters"), key)
	return err
}

func (g *Gohm) Incr(model interface{}, key string, step int64) (int64, error) {
	if err := validateModel(model); err != nil {
		return 0, err
	}
	id := modelID(model)
	if len(id) <= 0 {
		return 0, MissingIdError
	}

	conn := g.RedisPool.Get()
	defer conn.Close()

	return redis.Int64(conn.Do(
		"HINCRBY", connectKeys(modelType(model), id, "counters"), key, step))
}

func (g *Gohm) Decr(model interface{}, key string, step int64) (int64, error) {
	return g.Incr(model, key, - step)
}
