// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package discover

import (
	"log"
	"net"
	"strings"

	"github.com/fcavani/e"
	utilNet "github.com/fcavani/util/net"
)

// Server wait for a client and send some data to it.
type Server struct {
	Intface
	MulticastAddr
	AddrVer
	//Port number
	Port string
	// BufSize is the buffer size am must be equal to the client.
	BufSize int
	// Protocol function receive data from client and return something to this client.
	Protocol func(addr *net.UDPAddr, recv []byte) (msg []byte, err error)
	conn     *net.UDPConn
}

// Do method starts a goroutine that waites for the clients, and make responses with the
// Protocol function.
func (a *Server) Do() error {
	if a.Port == "" {
		a.Port = "3456"
	}
	if a.BufSize <= 0 {
		a.BufSize = 512
	}
	a.InitMCast()
	err := a.getInt()
	if err != nil {
		return e.Forward(err)
	}
	a.conn, err = a.bind()
	if err != nil {
		return e.Forward(err)
	}
	buf := make([]byte, a.BufSize)
	go func() {
		for {
			n, addr, err := a.conn.ReadFromUDP(buf)
			if e.Contains(err, "use of closed network connection") {
				return
			} else if err != nil {
				log.Printf("Server - ReadFromUDP (%v) failed: %v", addr, e.Trace(e.New(err)))
				continue
			}
			msg, err := a.Protocol(addr, buf[:n])
			if err != nil {
				log.Printf("Server - Protocol fail for %v with error: %v", addr, e.Trace(e.New(err)))
				_, _, err := a.conn.WriteMsgUDP([]byte("protocol fail"), nil, addr)
				if e.Contains(err, "use of closed network connection") {
					return
				}
				continue
			}
			if len(msg) > a.BufSize {
				log.Printf("Server - Protocol fail for %v message is too big (%v).", addr, len(msg))
				continue
			}
			n, oob, err := a.conn.WriteMsgUDP(msg, nil, addr)
			if e.Contains(err, "use of closed network connection") {
				return
			} else if err != nil {
				log.Printf("Server - WriteMsgUDP (%v) failed: %v", addr, e.Trace(e.New(err)))
				continue
			}
			if oob != 0 {
				log.Printf("Server - WriteMsgUDP to %v failed: %v, %v", addr, n, oob)
				continue
			}
			if n != len(msg) {
				log.Printf("Server - WriteMsgUDP to %v failed: %v, %v", addr, n, oob)
				continue
			}
		}
	}()
	return nil
}

// Close terminates the server.
func (a *Server) Close() error {
	err := a.conn.Close()
	if err != nil {
		return e.Forward(err)
	}
	return nil
}

func (a *Server) bind() (conn *net.UDPConn, err error) {
	if !a.NotMulticast && a.iface.Flags&net.FlagMulticast == net.FlagMulticast {
		gaddr, err := a.groupAddr()
		if err != nil {
			return nil, e.Forward(err)
		}
		conn, err = net.ListenMulticastUDP(a.Proto(), a.iface, gaddr)
		if err != nil {
			return nil, e.New(err)
		}
	} else {
		server, err := net.ResolveUDPAddr(a.Proto(), ":"+a.Port)
		if err != nil {
			return nil, e.New(err)
		}
		conn, err = net.ListenUDP(a.Proto(), server)
		if err != nil {
			return nil, e.New(err)
		}
	}
	return
}

func (a *Server) haveIpv6() (bool, error) {
	ipv4 := false
	addrs, err := a.iface.Addrs()
	if err != nil {
		return false, e.New(err)
	}
	for _, addr := range addrs {
		a := addr.String()
		i := strings.Index(a, "/")
		if i != -1 {
			a = a[:i]
		}
		if utilNet.IsValidIpv6(a) {
			return true, nil
		} else if utilNet.IsValidIpv4(a) {
			ipv4 = true
		}
	}
	if ipv4 {
		return false, nil
	}
	return false, e.New("no valid ip address")
}

func (a *Server) groupAddr() (*net.UDPAddr, error) {
	ipv6, err := a.haveIpv6()
	if err != nil {
		return nil, e.Forward(err)
	}
	if ipv6 && (a.AddrVer == Any || a.AddrVer == Ipv6) {
		addr, err := net.ResolveUDPAddr("udp", a.McIpv6+":"+a.Port)
		if err != nil {
			return nil, e.New(err)
		}
		return addr, nil
	} else {
		addr, err := net.ResolveUDPAddr("udp", a.McIpv4+":"+a.Port)
		if err != nil {
			return nil, e.New(err)
		}
		return addr, nil
	}
	return nil, e.New("invalid ip address")
}
