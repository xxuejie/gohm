package gohm

import (
	"errors"
	"github.com/garyburd/redigo/redis"
	"github.com/pote/go-msgpack"
	"github.com/pote/redisurl"
	"reflect"
	"sync"
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

	modelData := reflect.ValueOf(model).Elem()
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
	attrIndexMap := modelAttrIndexMap(model)
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
	indexIndexMap := modelIndexIndexMap(model)
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
	for attr, index := range modelUniqueIndexMap(model) {
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

func (g *Gohm) Delete(model interface{}) error {
	if err := validateModel(model); err != nil {
		return err
	}

	modelData := reflect.ValueOf(model).Elem()
	modelType := modelData.Type()
	modelId := modelID(model)
	if modelId == "" {
		return errors.New("MissingID!")
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
	for attr, index := range modelUniqueIndexMap(model) {
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

	modelLoadAttrs(attrs, model)
	// TODO: Support free conversion between string and int for ID field
	modelSetID(toString(id), model)
	return true, nil
}
