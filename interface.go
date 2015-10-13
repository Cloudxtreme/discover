// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package discover

import (
	"net"

	"github.com/fcavani/e"
)

type Intface struct {
	Interface    string
	NotMulticast bool
	iface        *net.Interface
}

func (i *Intface) getInt() error {
	if i.Interface != "" && i.iface == nil {
		ints, err := net.Interfaces()
		if err != nil {
			return e.New(err)
		}
		for _, in := range ints {
			if in.Name == i.Interface {
				i.iface = &in
				return nil
			}
		}
		return e.New("none interface with this name")
	} else if i.Interface == "" && i.iface == nil {
		ints, err := net.Interfaces()
		if err != nil {
			return e.New(err)
		}
		var intName string
		for _, in := range ints {
			if in.Flags&net.FlagMulticast == net.FlagMulticast || in.Flags&net.FlagBroadcast == net.FlagBroadcast || in.Flags&net.FlagBroadcast == net.FlagBroadcast {
				_, intName = getInterface(in)
				if intName == "" {
					continue
				}
				i.iface = &in
				break
			}
		}
	}
	return nil
}
