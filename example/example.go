package main

import (
	"log"
	"net"

	"github.com/fcavani/discover"
	"github.com/fcavani/e"
)

func main() {
	in, err := discover.Discover(net.FlagMulticast)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Interface:", in)

	server := &discover.Server{}
	server.Interface = in
	server.AddrVer = discover.Ipv4
	server.Protocol = func(addr *net.UDPAddr, req *discover.Request) (resp *discover.Response, err error) {
		if string(req.Data) != "request" {
			return nil, e.New("protocol error")
		}
		return &discover.Response{
			Data: []byte("msg"),
		}, nil
	}
	err = server.Do()
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	client := &discover.Client{}
	client.Interface = in
	client.AddrVer = discover.Ipv4
	client.Request = func(dst *net.UDPAddr) (*discover.Request, error) {
		return &discover.Request{
			Data: []byte("request"),
		}, nil
	}
	resp, err := client.Discover()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(resp)
}
