package gohm

import (
	"reflect"
	"sync"

	"github.com/garyburd/redigo/redis"
)

type BasicSet interface {
	// Per-type lock
	sync.Locker

	GetConn() redis.Conn
	FetchKey(conn redis.Conn) (string, error)
	Clean() error
	GetNamespace() string
}

func getConnAndKey(set BasicSet) (redis.Conn, string, error) {
	c := set.GetConn()
	key, err := set.FetchKey(c)
	if err != nil {
		c.Close()
		return nil, "", err
	}
	return c, key, nil
}

func cleanAndClose(set BasicSet, conn redis.Conn) error {
	err := set.Clean()
	if err != nil {
		return err
	}
	return conn.Close()
}

func SetInclude(set BasicSet, model interface{}) (bool, error) {
	return SetExists(set, modelID(model))
}

func SetExists(set BasicSet, id interface{}) (bool, error) {
	c, key, err := getConnAndKey(set)
	if err != nil {
		return false, err
	}
	defer cleanAndClose(set, c)
	ret, err := redis.Int(c.Do("SISMEMBER", key, toString(id)))
	if err != nil {
		return false, err
	}
	return ret == 1, nil
}

func SetIds(set BasicSet) ([]string, error) {
	c, key, err := getConnAndKey(set)
	if err != nil {
		return nil, err
	}
	defer cleanAndClose(set, c)
	return redis.Strings(c.Do("SMEMBERS", key))
}

func SetSize(set BasicSet) (int, error) {
	c, key, err := getConnAndKey(set)
	if err != nil {
		return 0, err
	}
	defer cleanAndClose(set, c)
	return redis.Int(c.Do("SCARD", key))
}

func setFetchData(set BasicSet, ids []interface{}) ([][]string, error) {
	set.Lock()
	defer set.Unlock()
	c := set.GetConn()
	defer c.Close()

	for i := range ids {
		c.Send("HGETALL", connectKeys(set.GetNamespace(), ids[i]))
	}
	c.Flush()
	resp := make([][]string, len(ids))
	for i := range resp {
		v, err := redis.Strings(c.Receive())
		if err != nil {
			return nil, err
		}
		resp[i] = append(v, "id", toString(ids[i]))
	}
	return resp, nil
}

func SetFetchByIds(set BasicSet, out interface{}, ids []interface{}) error {
	resp, err := setFetchData(set, ids)
	if err != nil {
		return err
	}
	return fillResponse(resp, out)
}

func SetFetch(set BasicSet, out interface{}) error {
	ids, err := SetIds(set)
	if err != nil {
		return err
	}
	interfaceIds := make([]interface{}, len(ids))
	for i := range ids {
		interfaceIds[i] = ids[i]
	}
	return SetFetchByIds(set, out, interfaceIds)
}

type Set struct {
	G         *Gohm
	Key       string
	Namespace string
	Model     reflect.Type
}

func NewSet(g *Gohm, key string, namespace string, model reflect.Type) Set {
	return Set{
		G:         g,
		Key:       key,
		Namespace: namespace,
		Model:     model,
	}
}

func (set Set) Lock() {
	set.G.Lock(set.Model.Name())
}

func (set Set) Unlock() {
	set.G.Unlock(set.Model.Name())
}

func (set Set) GetConn() redis.Conn {
	return set.G.RedisPool.Get()
}

func (set Set) FetchKey(conn redis.Conn) (string, error) {
	return set.Key, nil
}

func (set Set) Clean() error {
	return nil
}

func (set Set) GetNamespace() string {
	return set.Namespace
}

type MultiSet struct {
	G         *Gohm
	Namespace string
	Model     reflect.Type
	Command   Command
}

func NewMultiSet(g *Gohm, namespace string, model reflect.Type, command Command) MultiSet {
	return MultiSet{
		G:         g,
		Namespace: namespace,
		Model:     model,
		Command:   command,
	}
}

func (set MultiSet) Lock() {
	set.G.Lock(set.Model.Name())
}

func (set MultiSet) Unlock() {
	set.G.Unlock(set.Model.Name())
}

func (set MultiSet) GetConn() redis.Conn {
	return set.G.RedisPool.Get()
}

func (set MultiSet) FetchKey(conn redis.Conn) (string, error) {
	return set.Command.Call(connectKeys(set.Namespace, "tmp"), conn)
}

func (set MultiSet) Clean() error {
	return set.Command.Clean()
}

func (set MultiSet) GetNamespace() string {
	return set.Namespace
}
