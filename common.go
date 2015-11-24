// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package discover

import (
	"net"

	"github.com/fcavani/e"
	utilNet "github.com/fcavani/net"
)

type MulticastAddr struct {
	McIpv4 string
	McIpv6 string
}

func (m *MulticastAddr) InitMCast() {
	if m.McIpv4 == "" {
		m.McIpv4 = "224.0.0.1"
	}
	if m.McIpv6 == "" {
		m.McIpv6 = "[ff00::1]"
	}
}

const ErrNoInt = "no interface"

//Discover returns the interface name with the capabilite.
func Discover(flag net.Flags) (string, error) {
	ints, err := net.Interfaces()
	if err != nil {
		return "", e.New(err)
	}
	for _, in := range ints {
		if in.Flags&flag == flag {
			addrs, err := in.Addrs()
			if err != nil {
				continue
			}
			if len(addrs) == 0 {
				continue
			}
			return in.Name, nil
		}
	}
	return "", e.New(ErrNoInt)
}

func getInterface(in net.Interface) ([]net.Addr, string) {
	addrs, err := in.Addrs()
	if err != nil {
		return nil, ""
	}
	if len(addrs) == 0 {
		return nil, ""
	}
	return addrs, in.Name
}

type AddrVer uint8

const (
	Any AddrVer = iota
	Ipv4
	Ipv6
)

func (a AddrVer) AddrAllowed(ip string) bool {
	if utilNet.IsValidIpv6(ip) && (a == Any || a == Ipv6) {
		return true
	} else if utilNet.IsValidIpv4(ip) && (a == Any || a == Ipv4) {
		return true
	}
	return false
}

func (a AddrVer) Proto() string {
	switch a {
	case Any:
		return "udp"
	case Ipv4:
		return "udp4"
	case Ipv6:
		return "udp6"
	default:
		return "udp"
	}
	panic("not here")
}

// Request is sent by the client and contains the client
// ip and a payload.
type Request struct {
	Ip   string
	Id   string
	Data []byte
}

// Response is sent by the server to the client with
// a id, a sequence number, incoming order of the client, the ip address of the server
// and a payload.
type Response struct {
	Id   string
	Seq  uint16
	Ip   string
	Data []byte
}
