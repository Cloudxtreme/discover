// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package discover

import (
	"net"
	"sync"
	"time"

	"github.com/fcavani/e"
)

type context struct {
	Ttl  time.Time
	Id   string
	Seq  uint16
	Addr *net.UDPAddr
}

type contexts struct {
	ctxs     map[string]*context
	duration time.Duration
	lck      sync.RWMutex
	chclose  chan chan struct{}
}

func newContexts(duration, interval time.Duration) *contexts {
	c := &contexts{
		ctxs:     make(map[string]*context),
		duration: duration,
		chclose:  make(chan chan struct{}),
	}
	go func() {
		for {
			select {
			case <-time.After(interval):
				c.lck.Lock()
				for id, ctx := range c.ctxs {
					if time.Now().After(ctx.Ttl) {
						delete(c.ctxs, id)
					}
				}
				c.lck.Unlock()
			case ch := <-c.chclose:
				ch <- struct{}{}
			}
		}
	}()
	return c
}

func (c *contexts) Close() {
	ch := make(chan struct{})
	c.chclose <- ch
	<-ch
}

const ErrCtxAlreadyRegistered = "context already registered"

func (c *contexts) Register(ctx *context) error {
	c.lck.Lock()
	defer c.lck.Unlock()
	_, found := c.ctxs[ctx.Id]
	if found {
		return e.New(ErrCtxAlreadyRegistered)
	}
	c.ctxs[ctx.Id] = ctx
	return nil
}

const ErrCtxNotFound = "context not found"

func (c *contexts) Del(id string) error {
	c.lck.Lock()
	defer c.lck.Unlock()
	_, found := c.ctxs[id]
	if !found {
		return e.New(ErrCtxNotFound)
	}
	delete(c.ctxs, id)
	return nil
}

func (c *contexts) Get(id string) (*context, error) {
	c.lck.RLock()
	defer c.lck.RUnlock()
	ctx, found := c.ctxs[id]
	if !found {
		return nil, e.New(ErrCtxNotFound)
	}
	ctx.Ttl = time.Now().Add(c.duration)
	return ctx, nil
}
