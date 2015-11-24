// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package discover

import (
	"bytes"
	"crypto/rsa"
	"encoding/binary"
	"encoding/gob"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/fcavani/e"
	"github.com/fcavani/log"
	utilNet "github.com/fcavani/net"
	"github.com/fcavani/rand"
)

type msgType uint16

const (
	protoReq msgType = iota
	protoConfirm
	protoKeepAlive
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
	Protocol func(addr *net.UDPAddr, req *Request) (resp *Response, err error)
	// PrivateKey is the server private key
	PrivateKey *rsa.PrivateKey
	// PubKeys hold all pubkeys that will be used.
	PubKeys *PubKeys
	// Duration time of one session
	Duration time.Duration
	// Name is the server name. Used to identify the key
	Name   string
	conn   *net.UDPConn
	seq    []*net.UDPAddr
	lckSeq sync.Mutex
	ctxs   *contexts
}

func (a *Server) sendErr(addr *net.UDPAddr, er error) {
	respBuf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(respBuf)
	msg := &Msg{
		Err: er,
	}
	err := enc.Encode(msg)
	if err != nil {
		log.Tag("discover", "server").Error("Error encoding erro response:", err)
		return
	}
	if respBuf.Len() > a.BufSize {
		log.Tag("discover", "server").Error("Error encoding erro response: error response is too long", respBuf.Len())
		return
	}
	_, _, err = a.conn.WriteMsgUDP(respBuf.Bytes(), nil, addr)
	if err != nil {
		log.Tag("discover", "server").Error("Error sending erro response:", err)
	}
}

// Do method starts a goroutine that waites for the clients, and make responses with the
// Protocol function.
func (a *Server) Do() error {
	if a.Port == "" {
		a.Port = "0"
	}
	if a.BufSize <= 0 {
		a.BufSize = 1024
	}
	if a.Duration == 0 {
		a.Duration = 24 * time.Hour
	}
	if a.Name == "" {
		a.Name = "master"
	}
	a.seq = make([]*net.UDPAddr, 0)
	a.ctxs = newContexts(a.Duration, 300*time.Second)
	a.InitMCast()
	err := a.getInt()
	if err != nil {
		return e.Forward(err)
	}
	a.conn, err = a.bind()
	if err != nil {
		return e.Forward(err)
	}
	go func() {
		for {
			buf := make([]byte, a.BufSize)
			n, addr, err := a.conn.ReadFromUDP(buf)
			if e.Contains(err, "use of closed network connection") {
				return
			} else if err != nil {
				log.Tag("discover", "server").Printf("Server - ReadFromUDP (%v) failed: %v", addr, e.Trace(e.New(err)))
				continue
			}

			dec := gob.NewDecoder(bytes.NewReader(buf[:n]))
			var msg Msg
			err = dec.Decode(&msg)
			if err != nil {
				log.Tag("discover", "server").Printf("Can't decode data from %v.", addr)
				continue
			}

			pubkey, err := a.PubKeys.Get(msg.From)
			if err != nil {
				log.Tag("discover", "server").Printf("Invalid %v sender from %v.", msg.From, addr)
				continue
			}

			buf, err = msg.Message(pubkey, a.PrivateKey)
			if err != nil {
				log.Tag("discover", "server").Printf("Invalid message from %v: %v.", addr, err)
				continue
			}

			if len(buf) < binary.MaxVarintLen16 {
				log.Tag("discover", "server").Printf("Read insulficient data from %v.", addr)
				continue
			}
			typ, b := binary.Uvarint(buf[:binary.MaxVarintLen16])
			if b <= 0 {
				log.Tag("discover", "server").Print("Invalid package type from %v.", addr)
				continue
			}
			switch msgType(typ) {
			case protoConfirm:
				go a.confirm(addr, msg.From, pubkey, buf[binary.MaxVarintLen16:])
			case protoReq:
				go a.request(addr, msg.From, pubkey, buf[binary.MaxVarintLen16:])
			case protoKeepAlive:
				go a.keepalive(addr, msg.From, pubkey, buf[binary.MaxVarintLen16:])
			default:
				log.Tag("discover", "server").Errorf("Protocol error. (%v)", typ)
			}
		}
	}()
	return nil
}

func (a *Server) sendResp(resp *Response, to string, tokey *rsa.PublicKey, addr *net.UDPAddr) {
	respBuf := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(respBuf)
	err := enc.Encode(resp)
	if err != nil {
		log.Tag("discover", "server").Printf("Server - Protocol fail for %v with error: %v", addr, e.Trace(e.New(err)))
		a.sendErr(addr, e.Push(err, e.New("error enconding response")))
		return
	}

	msg, err := NewMsg(a.Name, to, a.PrivateKey, tokey, respBuf.Bytes())
	if err != nil {
		log.Tag("discover", "server").Printf("Server - Protocol fail for %v with error: %v", addr, e.Trace(e.New(err)))
		a.sendErr(addr, e.Push(err, e.New("error creating new response message")))
		return
	}

	respBuf = bytes.NewBuffer([]byte{})
	enc = gob.NewEncoder(respBuf)
	err = enc.Encode(msg)
	if err != nil {
		log.Tag("discover", "server").Printf("Server - Protocol fail for %v with error: %v", addr, e.Trace(e.New(err)))
		a.sendErr(addr, e.Push(err, e.New("error enconding response")))
		return
	}

	if respBuf.Len() > a.BufSize {
		log.Tag("discover", "server").Printf("Server - Protocol fail for %v message is too big (%v).", addr, respBuf.Len())
		a.sendErr(addr, e.Push(err, e.New("response is too long %v", respBuf.Len())))
		return
	}
	n, oob, err := a.conn.WriteMsgUDP(respBuf.Bytes(), nil, addr)
	if e.Contains(err, "use of closed network connection") {
		return
	} else if err != nil {
		log.Tag("discover", "server").Printf("Server - WriteMsgUDP (%v) failed: %v", addr, e.Trace(e.New(err)))
		return
	}
	if oob != 0 {
		log.Tag("discover", "server").Printf("Server - WriteMsgUDP to %v failed: %v, %v", addr, n, oob)
		return
	}
	if n != respBuf.Len() {
		log.Tag("discover", "server").Printf("Server - WriteMsgUDP to %v failed: %v, %v", addr, n, oob)
		return
	}
}

func (a *Server) request(addr *net.UDPAddr, to string, tokey *rsa.PublicKey, buf []byte) {
	dec := gob.NewDecoder(bytes.NewReader(buf))
	var req Request
	err := dec.Decode(&req)
	if err != nil {
		log.Tag("discover", "server").Printf("Server - Protocol fail for %v with error: %v", addr, e.Trace(e.New(err)))
		a.sendErr(addr, e.Push(err, e.New("error decoding request")))
		return
	}
	resp, err := a.Protocol(addr, &req)
	if err != nil {
		log.Tag("discover", "server").Printf("Server - Protocol fail for %v with error: %v", addr, e.Trace(e.New(err)))
		a.sendErr(addr, e.Push(err, e.New("protocol error")))
		return
	}

	ctx, err := a.ctxs.Get(req.Id)
	if err != nil {
		uuid, err := rand.Uuid()
		if err != nil {
			log.Tag("discover", "server").Printf("Server - Protocol fail for %v with error: %v", addr, e.Trace(e.New(err)))
			a.sendErr(addr, e.Push(err, e.New("protocol error")))
			return
		}
		resp.Id = uuid
		resp.Ip = addr.String()
		a.lckSeq.Lock()
		resp.Seq = uint16(len(a.seq))
		a.lckSeq.Unlock()
		err = a.ctxs.Register(&context{
			Id:   uuid,
			Seq:  resp.Seq,
			Addr: addr,
		})
		if err != nil {
			log.Tag("discover", "server").Printf("Server - Protocol fail for %v with error: %v", addr, e.Trace(e.New(err)))
			a.sendErr(addr, e.Push(err, e.New("protocol error")))
			return
		}
	} else {
		resp.Id = ctx.Id
		resp.Ip = ctx.Addr.String()
		resp.Seq = ctx.Seq
	}
	a.sendResp(resp, to, tokey, addr)
}

func (a *Server) confirm(addr *net.UDPAddr, to string, tokey *rsa.PublicKey, buf []byte) {
	dec := gob.NewDecoder(bytes.NewReader(buf))
	var id string
	err := dec.Decode(&id)
	if err != nil {
		log.Tag("discover", "server").Printf("Server - Protocol fail for %v with error: %v", addr, e.Trace(e.New(err)))
		a.sendErr(addr, e.Push(err, e.New("error decoding id")))
		return
	}
	ctx, err := a.ctxs.Get(id)
	if err != nil {
		log.Tag("discover", "server").Printf("Server - Protocol fail for %v with error: %v", addr, e.Trace(e.New(err)))
		a.sendErr(addr, e.Push(err, e.New("id is invalid")))
		return
	}
	a.lckSeq.Lock()
	a.seq = append(a.seq, ctx.Addr)
	a.lckSeq.Unlock()
	a.sendResp(&Response{
		Id:  ctx.Id,
		Ip:  ctx.Addr.String(),
		Seq: ctx.Seq,
	}, to, tokey, addr)
}

func (a *Server) keepalive(addr *net.UDPAddr, to string, tokey *rsa.PublicKey, buf []byte) {
	dec := gob.NewDecoder(bytes.NewReader(buf))
	var id string
	err := dec.Decode(&id)
	if err != nil {
		log.Tag("discover", "server").Printf("Server - Protocol fail for %v with error: %v", addr, e.Trace(e.New(err)))
		a.sendErr(addr, e.Push(err, e.New("error decoding id")))
		return
	}
	ctx, err := a.ctxs.Get(id)
	if err != nil {
		log.Tag("discover", "server").Printf("Server - Protocol fail for %v with error: %v", addr, e.Trace(e.New(err)))
		a.sendErr(addr, e.Push(err, e.New("id is invalid")))
		return
	}
	a.sendResp(&Response{
		Id:  ctx.Id,
		Ip:  ctx.Addr.String(),
		Seq: ctx.Seq,
	}, to, tokey, addr)
}

// Close terminates the server.
func (a *Server) Close() error {
	err := a.conn.Close()
	if err != nil {
		return e.Forward(err)
	}
	a.ctxs.Close()
	return nil
}

func (s *Server) ipver(addr net.Addr) {
	a := addr.String()
	if utilNet.IsValidIpv6(a) {
		s.AddrVer = Ipv6
	} else if utilNet.IsValidIpv4(a) {
		s.AddrVer = Ipv4
	}
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
	a.ipver(conn.LocalAddr())
	_, a.Port, err = utilNet.SplitHostPort(conn.LocalAddr().String())
	if err != nil {
		return nil, e.Forward(err)
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
	}
	addr, err := net.ResolveUDPAddr("udp", a.McIpv4+":"+a.Port)
	if err != nil {
		return nil, e.New(err)
	}
	return addr, nil
}
