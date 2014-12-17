package gohm

import (
	"github.com/garyburd/redigo/redis"
)

type Set struct {
	G      *Gohm
	Key       string
	Namespace string
}

func NewSet(g *Gohm, key string, namespace string) Set {
	return Set{
		G:      g,
		Key:       key,
		Namespace: namespace,
	}
}

func (set Set) Include(model interface{}) (bool, error) {
	if err := validateModel(model); err != nil {
		return false, err
	}
	return set.Exists(modelID(model))
}

func (set Set) Exists(id interface{}) (bool, error) {
	c := set.G.RedisPool.Get()
	defer c.Close()
	ret, err := redis.Int(c.Do("SISMEMBER", set.Key, toString(id)))
	if err != nil {
		return false, err
	}
	return ret == 1, nil
}

func (set Set) Ids() ([]string, error) {
	c := set.G.RedisPool.Get()
	defer c.Close()
	return redis.Strings(c.Do("SMEMBERS", set.Key))
}

func (set Set) fetchData() ([][]string, error) {
	// TODO: Add per-type lock
	set.G.Lock()
	defer set.G.Unlock()
	c := set.G.RedisPool.Get()
	defer c.Close()

	ids, err := set.Ids()
	if err != nil {
		return nil, err
	}
	for i := range ids {
		c.Send("HGETALL", connectKeys(set.Namespace, ids[i]))
	}
	c.Flush()
	resp := make([][]string, len(ids))
	for i := range resp {
		v, err := redis.Strings(c.Receive())
		if err != nil {
			return nil, err
		}
		resp[i] = append(v, "id", ids[i])
	}
	return resp, nil
}

func (set Set) Fetch(out interface{}) error {
	resp, err := set.fetchData()
	if err != nil {
		return err
	}
	return fillResponse(resp, out)
}
