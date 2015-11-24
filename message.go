// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package discover

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"

	"github.com/fcavani/e"
)

type Msg struct {
	From      string
	To        string
	Data      [][]byte
	Signature [][]byte
	Err       error
}

func NewMsg(from, to string, fromkey *rsa.PrivateKey, tokey *rsa.PublicKey, data []byte) (*Msg, error) {
	hash := crypto.SHA256
	max := tokey.N.BitLen()/8 - 2*hash.Size() - 2
	num := len(data) / max
	if len(data)%max > 0 {
		num++
	}
	msg := &Msg{
		From:      from,
		To:        to,
		Data:      make([][]byte, num),
		Signature: make([][]byte, num),
	}
	j := 0
	for i := 0; i < len(data); i += max {
		end := i + max
		if end > len(data) {
			end = len(data)
		}

		ciphertex, err := rsa.EncryptOAEP(hash.New(), rand.Reader, tokey, data[i:end], []byte(""))
		if err != nil {
			return nil, e.Push(err, "can't encrypt message")
		}
		pssh := hash.New()
		pssh.Write(data[i:end])
		hashed := pssh.Sum(nil)
		var opts rsa.PSSOptions
		opts.SaltLength = rsa.PSSSaltLengthAuto
		signature, err := rsa.SignPSS(rand.Reader, fromkey, hash, hashed, &opts)
		if err != nil {
			return nil, e.Push(err, "can't sign the message")
		}
		msg.Data[j] = ciphertex
		msg.Signature[j] = signature
		j++
	}

	return msg, nil
}

func (m *Msg) Message(fromkey *rsa.PublicKey, dstkey *rsa.PrivateKey) (data []byte, err error) {
	hash := crypto.SHA256
	data = make([]byte, 0)
	for i, d := range m.Data {
		plainText, err := rsa.DecryptOAEP(hash.New(), rand.Reader, dstkey, d, []byte(""))
		if err != nil {
			return nil, e.Push(err, "can't decrypt the message")
		}
		pssh := hash.New()
		pssh.Write(plainText)
		hashed := pssh.Sum(nil)
		var opts rsa.PSSOptions
		opts.SaltLength = rsa.PSSSaltLengthAuto
		err = rsa.VerifyPSS(fromkey, hash, hashed, m.Signature[i], &opts)
		if err != nil {
			return nil, e.Push(err, "can't verify the signature")
		}
		data = append(data, plainText...)
	}
	return data, nil
}
