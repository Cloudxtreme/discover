// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

/*Discover package contains methods to announce and discover
resources on a broadcast or multicast capable network.

See the exemple:

package main

import (
 	"errors"
	"log"
	"net"

	"github.com/fcavani/discover"
)

func main() {
	in, err := discover.Discover(net.FlagMulticast)
	if err != nil {
		log.Fatal(err)
	}

	server := &discover.Server{}
	server.Interface = in
	server.AddrVer = discover.Ipv4
	server.Protocol = func(addr *net.UDPAddr, recv []byte) (msg []byte, err error) {
		if string(recv) != "request" {
			return nil, errors.New("protocol error")
		}
		return []byte("msg"), nil
	}
	err = server.Do()
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	client := &discover.Client{}
	client.Interface = in
	client.AddrVer = Ipv4
	client.Request = func(dst *net.UDPAddr) ([]byte, error) {
		return []byte("request"), nil
	}
	buf, err := client.Discover()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(buf))
}
*/
package discover
