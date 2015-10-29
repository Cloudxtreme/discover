// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package discover

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/fcavani/e"
)

func TestServerMultiCast(t *testing.T) {
	in, err := Discover(net.FlagMulticast)
	if e.Equal(err, ErrNoInt) {
		t.Log("No multicast capable interface, may be this is travis.cl. Skip the test.")
		return
	} else if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}

	server := &Server{}
	server.Interface = in
	server.Port = "3333"
	server.Protocol = func(addr *net.UDPAddr, req *Request) (resp *Response, err error) {
		if string(req.Data) != "request" {
			return nil, e.New("protocol error")
		}
		t.Log(req)
		return &Response{
			Data: []byte("msg"),
		}, nil
	}
	err = server.Do()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	defer server.Close()

	client := &Client{}
	client.Interface = in
	client.Port = server.Port
	client.Request = func(dst *net.UDPAddr) (*Request, error) {
		return &Request{
			Data: []byte("request"),
		}, nil
	}
	resp, err := client.Discover()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	if string(resp.Data) != "msg" {
		t.Fatal("received wrong message", string(resp.Data))
	}
	t.Log(resp)
}

func TestServerLocalhost(t *testing.T) {
	in, err := Discover(net.FlagLoopback)
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}

	server := &Server{}
	server.Interface = in
	server.Protocol = func(addr *net.UDPAddr, req *Request) (resp *Response, err error) {
		if string(req.Data) != "request" {
			return nil, e.New("protocol error")
		}
		t.Log(req)
		return &Response{
			Data: []byte("msg"),
		}, nil
	}
	err = server.Do()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	defer server.Close()

	client := &Client{}
	client.Interface = in
	client.Port = server.Port
	client.Request = func(dst *net.UDPAddr) (*Request, error) {
		return &Request{
			Data: []byte("request"),
		}, nil
	}
	resp, err := client.Discover()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	if string(resp.Data) != "msg" {
		t.Fatal("received wrong message", string(resp.Data))
	}
	t.Log(resp)
}

func TestServerBroadcast(t *testing.T) {
	in, err := Discover(net.FlagBroadcast)
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}

	server := &Server{}
	server.Interface = in
	server.NotMulticast = true
	server.Protocol = func(addr *net.UDPAddr, req *Request) (resp *Response, err error) {
		if string(req.Data) != "request" {
			return nil, e.New("protocol error")
		}
		t.Log(req)
		return &Response{
			Data: []byte("msg"),
		}, nil
	}
	err = server.Do()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	defer server.Close()

	client := &Client{}
	client.Interface = in
	client.NotMulticast = true
	client.Port = server.Port
	client.Request = func(dst *net.UDPAddr) (*Request, error) {
		return &Request{
			Data: []byte("request"),
		}, nil
	}
	resp, err := client.Discover()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	if string(resp.Data) != "msg" {
		t.Fatal("received wrong message", string(resp.Data))
	}
	t.Log(resp)
}

func TestServerAny(t *testing.T) {
	time.Sleep(2 * time.Second)
	server := &Server{}
	server.Protocol = func(addr *net.UDPAddr, req *Request) (resp *Response, err error) {
		if string(req.Data) != "request" {
			return nil, e.New("protocol error")
		}
		t.Log(req)
		return &Response{
			Data: []byte("msg"),
		}, nil
	}
	err := server.Do()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	defer server.Close()

	client := &Client{}
	client.Port = server.Port
	client.Request = func(dst *net.UDPAddr) (*Request, error) {
		return &Request{
			Data: []byte("request"),
		}, nil
	}
	resp, err := client.Discover()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	if string(resp.Data) != "msg" {
		t.Fatal("received wrong message", string(resp.Data))
	}
	t.Log(resp)
}

func TestServerIpv4lo(t *testing.T) {
	in, err := Discover(net.FlagLoopback)
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}

	server := &Server{}
	server.Interface = in
	server.AddrVer = Ipv4
	server.Protocol = func(addr *net.UDPAddr, req *Request) (resp *Response, err error) {
		if string(req.Data) != "request" {
			return nil, e.New("protocol error")
		}
		t.Log(req)
		return &Response{
			Data: []byte("msg"),
		}, nil
	}
	err = server.Do()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	defer server.Close()

	client := &Client{}
	client.Interface = in
	client.AddrVer = Ipv4
	client.Port = server.Port
	client.Request = func(dst *net.UDPAddr) (*Request, error) {
		return &Request{
			Data: []byte("request"),
		}, nil
	}
	resp, err := client.Discover()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	if string(resp.Data) != "msg" {
		t.Fatal("received wrong message", string(resp.Data))
	}
	t.Log(resp)
}

func TestServerIpv4bc(t *testing.T) {
	in, err := Discover(net.FlagBroadcast)
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}

	server := &Server{}
	server.Interface = in
	server.AddrVer = Ipv4
	server.Protocol = func(addr *net.UDPAddr, req *Request) (resp *Response, err error) {
		if string(req.Data) != "request" {
			return nil, e.New("protocol error")
		}
		t.Log(req)
		return &Response{
			Data: []byte("msg"),
		}, nil
	}
	err = server.Do()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	defer server.Close()

	client := &Client{}
	client.Interface = in
	client.AddrVer = Ipv4
	client.Port = server.Port
	client.Request = func(dst *net.UDPAddr) (*Request, error) {
		return &Request{
			Data: []byte("request"),
		}, nil
	}
	resp, err := client.Discover()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	if string(resp.Data) != "msg" {
		t.Fatal("received wrong message", string(resp.Data))
	}
	t.Log(resp)
}

func TestServerIpv4mc(t *testing.T) {
	in, err := Discover(net.FlagMulticast)
	if e.Equal(err, ErrNoInt) {
		t.Log("No multicast capable interface, may be this is travis.cl. Skip the test.")
		return
	} else if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}

	server := &Server{}
	server.Interface = in
	server.AddrVer = Ipv4
	server.Protocol = func(addr *net.UDPAddr, req *Request) (resp *Response, err error) {
		if string(req.Data) != "request" {
			return nil, e.New("protocol error")
		}
		t.Log(req)
		return &Response{
			Data: []byte("msg"),
		}, nil
	}
	err = server.Do()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	defer server.Close()

	client := &Client{}
	client.Interface = in
	client.AddrVer = Ipv4
	client.Port = server.Port
	client.Request = func(dst *net.UDPAddr) (*Request, error) {
		return &Request{
			Data: []byte("request"),
		}, nil
	}
	resp, err := client.Discover()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	if string(resp.Data) != "msg" {
		t.Fatal("received wrong message", string(resp.Data))
	}
	t.Log(resp)
}

func TestServerFail(t *testing.T) {
	server := &Server{}
	server.Interface = ":)"
	server.AddrVer = Any
	server.Protocol = func(addr *net.UDPAddr, req *Request) (resp *Response, err error) {
		if string(req.Data) != "request" {
			return nil, e.New("protocol error")
		}
		t.Log(req)
		return &Response{
			Data: []byte("msg"),
		}, nil
	}
	err := server.Do()
	if err != nil && !e.Equal(err, "none interface with this name") {
		t.Fatal(e.Trace(e.Forward(err)))
	}
}

func TestClientFail(t *testing.T) {
	client := &Client{}
	client.Interface = ":)"
	client.AddrVer = Any
	client.Port = "6464"
	client.Request = func(dst *net.UDPAddr) (*Request, error) {
		return &Request{
			Data: []byte("request"),
		}, nil
	}
	_, err := client.Discover()
	if err != nil && !e.Equal(err, "none interface with this name") {
		t.Fatal(e.Trace(e.Forward(err)))
	}

	client = &Client{}
	client.AddrVer = Any
	client.Port = "6465"
	client.Timeout = 1 * time.Second
	client.Deadline = 100 * time.Millisecond
	client.Request = func(dst *net.UDPAddr) (*Request, error) {
		return &Request{
			Data: []byte("request"),
		}, nil
	}
	_, err = client.Discover()
	if err != nil && !e.Equal(err, "no addresses capable for listen udp") {
		t.Fatal(e.Trace(e.Forward(err)))
	}
}

func TestServerProtocolFail(t *testing.T) {
	in, err := Discover(net.FlagLoopback)
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}

	server := &Server{}
	server.Interface = in
	server.AddrVer = Ipv4
	server.Protocol = func(addr *net.UDPAddr, req *Request) (resp *Response, err error) {
		if string(req.Data) != "request" {
			return nil, e.New("protocol error")
		}
		t.Log(req)
		return &Response{
			Data: []byte("msg"),
		}, nil
	}
	err = server.Do()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	defer server.Close()

	client := &Client{}
	client.Interface = in
	client.AddrVer = Ipv4
	client.Port = server.Port
	client.Request = func(dst *net.UDPAddr) (*Request, error) {
		return &Request{
			Data: []byte("request"),
		}, nil
	}
	_, err = client.Discover()
	if err != nil && e.Find(err, "protocol fail") < 0 {
		t.Fatal(e.Trace(e.Forward(err)))
	}
}

func TestClientAgain(t *testing.T) {
	in, err := Discover(net.FlagMulticast)
	if e.Equal(err, ErrNoInt) {
		t.Log("No multicast capable interface, may be this is travis.cl. Skip the test.")
		return
	} else if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}

	server := &Server{}
	server.Interface = in
	server.Port = "3333"
	server.Protocol = func(addr *net.UDPAddr, req *Request) (resp *Response, err error) {
		if string(req.Data) != "request" {
			return nil, e.New("protocol error")
		}
		t.Log(req)
		return &Response{
			Data: []byte("msg"),
		}, nil
	}
	err = server.Do()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	defer server.Close()

	client := &Client{}
	client.Interface = in
	client.Port = server.Port
	client.Request = func(dst *net.UDPAddr) (*Request, error) {
		return &Request{
			Data: []byte("request"),
		}, nil
	}
	resp1, err := client.Discover()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	if string(resp1.Data) != "msg" {
		t.Fatal("received wrong message", string(resp1.Data))
	}
	t.Log(resp1)

	resp2, err := client.Discover()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	if string(resp2.Data) != "msg" {
		t.Fatal("received wrong message", string(resp2.Data))
	}
	t.Log(resp2)

	if resp1.Id != resp2.Id || resp1.Seq != resp2.Seq {
		t.Fatal("not the same session")
	}
}

// Example demonstrate discovery in work.
func Example() {
	server := &Server{}
	server.Protocol = func(addr *net.UDPAddr, req *Request) (resp *Response, err error) {
		if string(req.Data) != "request" {
			return nil, e.New("protocol error")
		}
		return &Response{
			Data: []byte("msg"),
		}, nil
	}
	err := server.Do()
	if err != nil {
		fmt.Println(err)
	}
	defer server.Close()

	client := &Client{}
	client.Port = server.Port
	client.Request = func(dst *net.UDPAddr) (*Request, error) {
		return &Request{
			Data: []byte("request"),
		}, nil
	}
	resp, err := client.Discover()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(resp.Data))
	//Output:
	//msg
}
