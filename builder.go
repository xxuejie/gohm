package gohm

import (
	"sync"

	"github.com/garyburd/redigo/redis"
	"github.com/pote/redisurl"
)

type builder struct {
	RedisPool *redis.Pool
	Callbacks map[string][]CallbackFunc
}

func NewBuilder() *builder {
	return &builder{}
}

func (b *builder) WithPool(r *redis.Pool) *builder {
	b.RedisPool = r
	return b
}

func (b *builder) WithCallback(model interface{}, callback CallbackFunc) *builder {
	fullName := typeFullName(modelReflectType(model))
	if b.Callbacks == nil {
		b.Callbacks = make(map[string][]CallbackFunc)
	}
	callbacks, ok := b.Callbacks[fullName]
	if !ok {
		callbacks = make([]CallbackFunc, 0)
	}
	b.Callbacks[fullName] = append(callbacks, callback)
	return b
}

func (b *builder) Build() (*Gohm, error) {
	if b.RedisPool == nil {
		var err error
		b.RedisPool, err = redisurl.NewPool(3, 200, "240s")
		if err != nil {
			return &Gohm{}, err
		}
	}
	return &Gohm{
		RedisPool: b.RedisPool,
		LuaSave: redis.NewScript(0, LUA_SAVE),
		LuaDelete: redis.NewScript(0, LUA_DELETE),
		Callbacks: b.Callbacks,
		TypeLocks: make(map[string]*sync.Mutex),
	}, nil
}
