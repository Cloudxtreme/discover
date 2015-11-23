// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package discover

import (
	"crypto/rsa"
	"sync"

	"github.com/fcavani/e"
)

type PubKeys struct {
	Keys map[string]*rsa.PublicKey
	lck  sync.RWMutex
}

const ErrKeyNotFound = "key not found for this id"

func NewPubKeys() *PubKeys {
	return &PubKeys{
		Keys: make(map[string]*rsa.PublicKey),
	}
}

func (p *PubKeys) Get(id string) (*rsa.PublicKey, error) {
	p.lck.RLock()
	defer p.lck.RUnlock()
	key, found := p.Keys[id]
	if !found {
		return nil, e.New(ErrKeyNotFound)
	}
	return key, nil
}

func (p *PubKeys) Delete(id string) error {
	p.lck.Lock()
	defer p.lck.Unlock()
	_, found := p.Keys[id]
	if !found {
		return e.New(ErrKeyNotFound)
	}
	delete(p.Keys, id)
	return nil
}

func (p *PubKeys) Put(id string, key *rsa.PublicKey) {
	p.lck.Lock()
	defer p.lck.Unlock()
	p.Keys[id] = key
}
