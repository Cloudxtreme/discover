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
	"strconv"
	"strings"
	"time"

	"github.com/fcavani/e"
	"github.com/fcavani/log"
	utilNet "github.com/fcavani/net"
	"github.com/fcavani/rand"
)

// Client struct provide methods for discover a server in a multicast or broadcast network.
type Client struct {
	Intface
	MulticastAddr
	AddrVer
	// Port number
	Port string
	// BufSize is the buffer size am must be equal to the server.
	BufSize int
	// Timeout is the amount of time with the client try to connect the server.
	Timeout time.Duration
	// Deadline is the udp io deadline
	Deadline time.Duration
	// Keepalive is the periode of the keepalive package
	Keepalive time.Duration
	// Request function returns the data that will be send to the server.
	Request    func(dst *net.UDPAddr) (*Request, error)
	ServerName string
	//ServerKey is the server public key
	ServerKey *rsa.PublicKey
	// Name is the name of this client. This is used to pick the right public key.
	Name       string
	PrivateKey *rsa.PrivateKey
	// Id is the unique identification for this client
	Id     string
	stopKa chan chan struct{}
	conn   *net.UDPConn
}

// Discover funtion discovers the server and returns the data sent by the server.
func (c *Client) Discover() (*Response, error) {
	if c.Port == "" {
		c.Port = "3456"
	}
	if c.BufSize <= 0 {
		c.BufSize = 1024
	}
	if c.Timeout <= 0 {
		c.Timeout = 2 * time.Minute
	}
	if c.Deadline <= 0 {
		c.Deadline = 10 * time.Second
	}
	if c.Keepalive <= 0 {
		c.Keepalive = 10 * time.Second
	}
	var err error
	if c.Id == "" {
		c.Id, err = rand.Uuid()
		if err != nil {
			return nil, e.Forward(err)
		}
	}
	c.stopKa = make(chan chan struct{})
	c.InitMCast()
	err = c.getInt()
	if err != nil {
		return nil, e.Forward(err)
	}
	resp, err := c.getAddr()
	if err != nil {
		return nil, e.Forward(err)
	}
	return resp, nil
}

func (c *Client) getAddr() (*Response, error) {
	addrs, err := c.iface.Addrs()
	if err != nil {
		return nil, e.New(err)
	}
	for _, addr := range addrs {
		a := addr.String()
		i := strings.Index(a, "/")
		if i != -1 {
			a = a[:i]
		}
		if !c.AddrAllowed(a) {
			continue
		}
		resp, err := c.client(a)
		if e.Equal(err, ErrCantFindInt) {
			continue
		} else if err != nil {
			return nil, e.Forward(err)
		}
		return resp, nil
	}
	return nil, e.New("no addresses capable for listen udp")
}

func (c *Client) encode(typ msgType, val interface{}, dst *net.UDPAddr) error {
	log.ProtoLevel().Tag("client", "discover").Printf("Send request (%v) to %v from %v.", typ, dst, c.conn.LocalAddr())

	reqBuf := bytes.NewBuffer([]byte{})

	buf := make([]byte, binary.MaxVarintLen16)
	binary.PutUvarint(buf, uint64(typ))
	n, err := reqBuf.Write(buf)
	if err != nil {
		return e.Push(err, "error enconding message type")
	}
	if n != len(buf) {
		return e.Push(err, "error enconding message type")
	}

	enc := gob.NewEncoder(reqBuf)
	err = enc.Encode(val)
	if err != nil {
		return e.Push(err, e.New("error encoding"))
	}

	msg, err := NewMsg(c.Name, c.ServerName, c.PrivateKey, c.ServerKey, reqBuf.Bytes())
	if err != nil {
		return e.Push(err, "erro cryptographing the value")
	}

	reqBuf = bytes.NewBuffer([]byte{})
	enc = gob.NewEncoder(reqBuf)
	err = enc.Encode(msg)
	if err != nil {
		return e.Push(err, e.New("error encoding"))
	}

	if reqBuf.Len() > c.BufSize {
		return e.New("value to encode is too big %v", reqBuf.Len())
	}
	err = c.conn.SetDeadline(time.Now().Add(c.Deadline))
	if err != nil {
		return e.New(err)
	}
	_, _, err = c.conn.WriteMsgUDP(reqBuf.Bytes(), nil, dst)
	if err != nil {
		return e.New(err)
	}
	err = c.conn.SetDeadline(time.Time{})
	if err != nil {
		return e.New(err)
	}
	return nil
}

func (c *Client) response() (*Response, error) {
	log.ProtoLevel().Tag("client", "discover").Printf("Waiting response...")
	buf := make([]byte, c.BufSize)
	err := c.conn.SetDeadline(time.Now().Add(c.Deadline))
	if err != nil {
		return nil, e.New(err)
	}
	n, addr, err := c.conn.ReadFromUDP(buf)
	if err != nil {
		return nil, e.New(err)
	}
	log.ProtoLevel().Tag("client", "discover").Printf("Response from %v with size %v.", addr, n)
	err = c.conn.SetDeadline(time.Time{})
	if err != nil {
		return nil, e.New(err)
	}

	dec := gob.NewDecoder(bytes.NewReader(buf[:n]))
	var msg Msg
	err = dec.Decode(&msg)
	if err != nil {
		return nil, e.Push(err, e.New("error decoding response"))
	}

	if msg.Err != nil {
		return nil, e.Forward(msg.Err)
	}

	if msg.From != c.ServerName {
		return nil, e.New("wrong server name")
	}
	if msg.To != c.Name {
		return nil, e.New("message isn't for me")
	}

	buf, err = msg.Message(c.ServerKey, c.PrivateKey)
	if err != nil {
		return nil, e.Push(err, e.New("error decrypting response"))
	}

	dec = gob.NewDecoder(bytes.NewReader(buf))
	var resp Response
	err = dec.Decode(&resp)
	if err != nil {
		return nil, e.Push(err, e.New("error decoding response"))
	}
	return &resp, nil
}

const ErrCantFindInt = "can't find an interface with the right capabilites"

func (c *Client) client(addr string) (*Response, error) {
	ip, err := ipport(c.Interface, addr, "0")
	if err != nil {
		return nil, e.Push(err, ErrCantFindInt)
	}
	client, err := net.ResolveUDPAddr("udp", ip)
	if err != nil {
		return nil, e.Push(err, ErrCantFindInt)
	}
	c.conn, err = net.ListenUDP("udp", client)
	if err != nil {
		return nil, e.Push(err, ErrCantFindInt)
	}
	var dst *net.UDPAddr
	if c.iface.Flags&net.FlagLoopback == net.FlagLoopback {
		ip, err := ipport(c.Interface, addr, c.Port)
		if err != nil {
			return nil, e.Push(err, ErrCantFindInt)
		}
		dst, err = net.ResolveUDPAddr("udp", ip)
		if err != nil {
			return nil, e.Push(err, ErrCantFindInt)
		}
	} else if !c.NotMulticast && c.iface.Flags&net.FlagMulticast == net.FlagMulticast {
		dst, err = c.multicast(c.conn.LocalAddr())
		if err != nil {
			return nil, e.Push(err, ErrCantFindInt)
		}
	} else if c.iface.Flags&net.FlagBroadcast == net.FlagBroadcast {
		dst, err = broadcast(c.conn.LocalAddr(), c.Port)
		if err != nil {
			return nil, e.Push(err, ErrCantFindInt)
		}
	} else {
		return nil, e.Push(e.New("interface isn't suported: %v", c.iface.Flags), ErrCantFindInt)
	}
	log.ProtoLevel().Tag("discover", "client").Printf("Local ip %v.", c.conn.LocalAddr())
	log.ProtoLevel().Tag("discover", "client").Printf("Try to contact server in %v.", dst)
	now := time.Now()
	end := now.Add(c.Timeout)
	for d := now; d.Before(end) || d.Equal(end); d = time.Now() {
		req, err := c.Request(dst)
		if err != nil {
			return nil, e.Forward(err)
		}

		req.Id = c.Id
		req.Ip = c.conn.LocalAddr().String()

		err = c.encode(protoReq, req, dst)
		if e.Contains(err, "i/o timeout") {
			log.Errorf("Error %v -> %v: %v", c.conn.LocalAddr(), dst, err)
			continue
		} else if err != nil {
			return nil, e.Forward(err)
		}

		resp, err := c.response()
		if e.Contains(err, "i/o timeout") {
			log.Errorf("Error %v -> %v: %v", c.conn.LocalAddr(), dst, err)
			continue
		} else if err != nil {
			return nil, e.Forward(err)
		}

		c.Id = resp.Id

		err = c.encode(protoConfirm, resp.Id, dst)
		if e.Contains(err, "i/o timeout") {
			log.Errorf("Error %v -> %v: %v", c.conn.LocalAddr(), dst, err)
			continue
		} else if err != nil {
			return nil, e.Forward(err)
		}

		rp, err := c.response()
		if e.Contains(err, "i/o timeout") {
			log.Errorf("Error %v -> %v: %v", c.conn.LocalAddr(), dst, err)
			continue
		} else if err != nil {
			return nil, e.Forward(err)
		}

		if rp.Id != resp.Id {
			return nil, e.New("protocol fail wrong response")
		}

		go func(dst *net.UDPAddr) {
			for {
				select {
				case <-time.After(c.Keepalive):
					log.ProtoLevel().Tag("client", "discover").Printf("Send keep alive to %v", dst)
					err := c.keepalive(dst)
					if err != nil {
						log.Tag("client", "discover").Errorf("Keep alive to %v failed: %v", dst, err)
						return
					}
				case ch := <-c.stopKa:
					ch <- struct{}{}
					return
				}
			}
		}(dst)

		return resp, nil
	}
	return nil, e.New("can't find the server")
}

func ipport(in, ip, port string) (string, error) {
	if utilNet.IsValidIpv4(ip) {
		return ip + ":" + port, nil
	} else if utilNet.IsValidIpv6(ip) {
		return "[" + ip + "%" + in + "]:" + port, nil
	} else {
		return "", e.New("invalid ip address")
	}
}

func broadcast(addr net.Addr, port string) (*net.UDPAddr, error) {
	p, err := strconv.ParseInt(port, 10, 32)
	if err != nil {
		return nil, e.New(err)
	}
	udpAddr := &net.UDPAddr{Port: int(p)}
	a, ok := addr.(*net.UDPAddr)
	if !ok {
		return nil, e.New("addr isn't an *UDPAddr")
	}
	if utilNet.IsValidIpv4(a.IP.String()) {
		udpAddr.IP = net.IPv4bcast
	} else if utilNet.IsValidIpv6(a.IP.String()) {
		udpAddr.IP = net.IPv6linklocalallnodes
		udpAddr.Zone = a.Zone
	} else {
		return nil, e.New("invalid ip address")
	}
	return udpAddr, nil
}

func (c *Client) multicast(addr net.Addr) (*net.UDPAddr, error) {
	host, _, err := utilNet.SplitHostPort(addr.String())
	if err != nil {
		return nil, e.Forward(err)
	}
	if utilNet.IsValidIpv4(host) {
		addr, err := net.ResolveUDPAddr("udp", c.McIpv4+":"+c.Port)
		if err != nil {
			return nil, e.Forward(err)
		}
		return addr, nil
	} else if utilNet.IsValidIpv6(host) {
		addr, err := net.ResolveUDPAddr("udp", c.McIpv6+":"+c.Port)
		if err != nil {
			return nil, e.Forward(err)
		}
		return addr, nil
	} else {
		return nil, e.New("invalid ip address")
	}
}

func (c *Client) keepalive(dst *net.UDPAddr) error {
	err := c.encode(protoKeepAlive, c.Id, dst)
	if err != nil {
		return e.Forward(err)
	}
	_, err = c.response()
	if err != nil {
		return e.Forward(err)
	}
	return nil
}

func (c *Client) Close() error {
	ch := make(chan struct{})
	c.stopKa <- ch
	<-ch
	return e.New(c.conn.Close())
}
