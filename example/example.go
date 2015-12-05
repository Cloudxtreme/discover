package main

import (
	"crypto/rand"
	"crypto/rsa"
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

	MasterKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}
	SlaveKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}

	Keys := discover.NewPubKeys()
	Keys.Put("slave", &SlaveKey.PublicKey)

	server := &discover.Server{}
	server.Name = "master"
	server.PrivateKey = MasterKey
	server.PubKeys = Keys
	server.Interface = in
	server.AddrVer = discover.Ipv4
	server.Port = "3333"
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
	client.ServerName = "master"
	client.ServerKey = &MasterKey.PublicKey
	client.Name = "slave"
	client.PrivateKey = SlaveKey
	client.Interface = in
	client.AddrVer = discover.Ipv4
	client.Port = "3333"
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
