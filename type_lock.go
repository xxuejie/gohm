package gohm

import (
	"sync"
)

func (g *Gohm) Lock(name string) {
	g.fetchTypeLock(name).Lock()
}

func (g *Gohm) Unlock(name string) {
	g.fetchTypeLock(name).Unlock()
}

func (g *Gohm) fetchTypeLock(name string) *sync.Mutex {
	g.TypeLockLock.Lock()
	defer g.TypeLockLock.Unlock()

	lock, ok := g.TypeLocks[name]
	if !ok {
		lock = &sync.Mutex{}
		g.TypeLocks[name] = lock
	}
	return lock
}
