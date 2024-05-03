package websocketcmd

import (
	"sync"
)

type writerCache struct {
	mu       sync.Mutex
	aWriters map[string]*writer
	oWriters map[string]*writer
}

func newConnCache() *writerCache {
	return &writerCache{
		mu:       sync.Mutex{},
		aWriters: map[string]*writer{},
		oWriters: map[string]*writer{},
	}
}

func (c *writerCache) getA(peerName string) (*writer, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	conn, ok := c.aWriters[peerName]
	return conn, ok
}

func (c *writerCache) setA(peerName string, newW *writer, close bool) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	w, ok := c.aWriters[peerName]
	if ok && close {
		err = w.conn.Close()
	}
	c.aWriters[peerName] = newW
	return err
}

func (c *writerCache) getO(peerName string) (*writer, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	conn, ok := c.oWriters[peerName]
	return conn, ok
}

func (c *writerCache) setO(peerName string, newW *writer, close bool) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	w, ok := c.oWriters[peerName]
	if ok && close {
		err = w.conn.Close()
	}
	c.oWriters[peerName] = newW
	return err
}
