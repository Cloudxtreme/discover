// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package discover

import (
	"errors"
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
	server.Protocol = func(addr *net.UDPAddr, recv []byte) (msg []byte, err error) {
		if string(recv) != "request" {
			return nil, e.New("protocol error")
		}
		return []byte("msg"), nil
	}
	err = server.Do()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	defer server.Close()

	client := &Client{}
	client.Interface = in
	client.Request = func(dst *net.UDPAddr) ([]byte, error) {
		return []byte("request"), nil
	}
	buf, err := client.Discover()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	if string(buf) != "msg" {
		t.Fatal("received wrong message", string(buf))
	}
}

func TestServerLocalhost(t *testing.T) {
	in, err := Discover(net.FlagLoopback)
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}

	server := &Server{}
	server.Interface = in
	server.Protocol = func(addr *net.UDPAddr, recv []byte) (msg []byte, err error) {
		if string(recv) != "request" {
			return nil, e.New("protocol error")
		}
		return []byte("msg"), nil
	}
	err = server.Do()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	defer server.Close()

	client := &Client{}
	client.Interface = in
	client.Request = func(dst *net.UDPAddr) ([]byte, error) {
		return []byte("request"), nil
	}
	buf, err := client.Discover()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	if string(buf) != "msg" {
		t.Fatal("received wrong message")
	}
}

func TestServerBroadcast(t *testing.T) {
	in, err := Discover(net.FlagBroadcast)
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}

	server := &Server{}
	server.Interface = in
	server.NotMulticast = true
	server.Protocol = func(addr *net.UDPAddr, recv []byte) (msg []byte, err error) {
		if string(recv) != "request" {
			return nil, e.New("protocol error")
		}
		return []byte("msg"), nil
	}
	err = server.Do()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	defer server.Close()

	client := &Client{}
	client.Interface = in
	client.NotMulticast = true
	client.Request = func(dst *net.UDPAddr) ([]byte, error) {
		return []byte("request"), nil
	}
	buf, err := client.Discover()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	if string(buf) != "msg" {
		t.Fatal("received wrong message")
	}
}

func TestServerAny(t *testing.T) {
	time.Sleep(2 * time.Second)
	server := &Server{}
	server.Protocol = func(addr *net.UDPAddr, recv []byte) (msg []byte, err error) {
		if string(recv) != "request" {
			return nil, e.New("protocol error")
		}
		return []byte("msg"), nil
	}
	err := server.Do()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	defer server.Close()

	client := &Client{}
	client.Request = func(dst *net.UDPAddr) ([]byte, error) {
		return []byte("request"), nil
	}
	buf, err := client.Discover()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	if string(buf) != "msg" {
		t.Fatal("received wrong message", string(buf))
	}
}

func TestServerIpv4lo(t *testing.T) {
	in, err := Discover(net.FlagLoopback)
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}

	server := &Server{}
	server.Interface = in
	server.AddrVer = Ipv4
	server.Protocol = func(addr *net.UDPAddr, recv []byte) (msg []byte, err error) {
		if string(recv) != "request" {
			return nil, e.New("protocol error")
		}
		return []byte("msg"), nil
	}
	err = server.Do()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	defer server.Close()

	client := &Client{}
	client.Interface = in
	client.AddrVer = Ipv4
	client.Request = func(dst *net.UDPAddr) ([]byte, error) {
		return []byte("request"), nil
	}
	buf, err := client.Discover()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	if string(buf) != "msg" {
		t.Fatal("received wrong message")
	}
}

func TestServerIpv4bc(t *testing.T) {
	in, err := Discover(net.FlagBroadcast)
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}

	server := &Server{}
	server.Interface = in
	server.AddrVer = Ipv4
	server.Protocol = func(addr *net.UDPAddr, recv []byte) (msg []byte, err error) {
		if string(recv) != "request" {
			return nil, e.New("protocol error")
		}
		return []byte("msg"), nil
	}
	err = server.Do()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	defer server.Close()

	client := &Client{}
	client.Interface = in
	client.AddrVer = Ipv4
	client.Request = func(dst *net.UDPAddr) ([]byte, error) {
		return []byte("request"), nil
	}
	buf, err := client.Discover()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	if string(buf) != "msg" {
		t.Fatal("received wrong message", string(buf))
	}
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
	server.Protocol = func(addr *net.UDPAddr, recv []byte) (msg []byte, err error) {
		if string(recv) != "request" {
			return nil, e.New("protocol error")
		}
		return []byte("msg"), nil
	}
	err = server.Do()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	defer server.Close()

	client := &Client{}
	client.Interface = in
	client.AddrVer = Ipv4
	client.Request = func(dst *net.UDPAddr) ([]byte, error) {
		return []byte("request"), nil
	}
	buf, err := client.Discover()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	if string(buf) != "msg" {
		t.Fatal("received wrong message")
	}
}

func TestServerFail(t *testing.T) {
	server := &Server{}
	server.Interface = ":)"
	server.AddrVer = Any
	server.Protocol = func(addr *net.UDPAddr, recv []byte) (msg []byte, err error) {
		if string(recv) != "request" {
			return nil, e.New("protocol error")
		}
		return []byte("msg"), nil
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
	client.Request = func(dst *net.UDPAddr) ([]byte, error) {
		return []byte("request"), nil
	}
	_, err := client.Discover()
	if err != nil && !e.Equal(err, "none interface with this name") {
		t.Fatal(e.Trace(e.Forward(err)))
	}

	client = &Client{}
	client.AddrVer = Any
	client.Timeout = 1 * time.Second
	client.Deadline = 100 * time.Millisecond
	client.Request = func(dst *net.UDPAddr) ([]byte, error) {
		return []byte("request"), nil
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
	server.Protocol = func(addr *net.UDPAddr, recv []byte) (msg []byte, err error) {
		if string(recv) != "request" {
			return nil, e.New("protocol error")
		}
		return []byte("msg"), nil
	}
	err = server.Do()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	defer server.Close()

	client := &Client{}
	client.Interface = in
	client.AddrVer = Ipv4
	client.Request = func(dst *net.UDPAddr) ([]byte, error) {
		return []byte("blá"), nil
	}
	_, err = client.Discover()
	if err != nil && e.Find(err, "protocol fail") < 0 {
		t.Fatal(e.Trace(e.Forward(err)))
	}
}

// Example demonstrate discovery in work.
func Example() {
	server := &Server{}
	server.Protocol = func(addr *net.UDPAddr, recv []byte) (msg []byte, err error) {
		if string(recv) != "request" {
			return nil, errors.New("protocol error")
		}
		return []byte("msg"), nil
	}
	err = server.Do()
	if err != nil {
		fmt.Println(err)
	}
	defer server.Close()

	client := &Client{}
	client.Request = func(dst *net.UDPAddr) ([]byte, error) {
		return []byte("request"), nil
	}
	buf, err := client.Discover()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(buf))
	//Output:
	//msg
}
