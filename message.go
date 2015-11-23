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
	Data      []byte
	Signature []byte
}

func NewMsg(from, to string, fromkey *rsa.PrivateKey, tokey *rsa.PublicKey, data []byte) (*Msg, error) {
	hash := crypto.SHA256
	ciphertext, err := rsa.EncryptOAEP(hash.New(), rand.Reader, tokey, data, []byte(""))
	if err != nil {
		return nil, e.Push(err, "can't encrypt message")
	}
	pssh := hash.New()
	pssh.Write(data)
	hashed := pssh.Sum(nil)
	var opts rsa.PSSOptions
	opts.SaltLength = rsa.PSSSaltLengthAuto
	signature, err := rsa.SignPSS(rand.Reader, fromkey, hash, hashed, &opts)
	if err != nil {
		return nil, e.Push(err, "can't sign the message")
	}
	return &Msg{
		From:      from,
		To:        to,
		Data:      ciphertext,
		Signature: signature,
	}, nil
}

func (m *Msg) Message(fromkey *rsa.PublicKey, dstkey *rsa.PrivateKey) ([]byte, error) {
	hash := crypto.SHA256
	plainText, err := rsa.DecryptOAEP(hash.New(), rand.Reader, dstkey, m.Data, []byte(""))
	if err != nil {
		return nil, e.Push(err, "can't decrypt the message")
	}
	pssh := hash.New()
	pssh.Write(plainText)
	hashed := pssh.Sum(nil)
	var opts rsa.PSSOptions
	opts.SaltLength = rsa.PSSSaltLengthAuto
	err = rsa.VerifyPSS(fromkey, hash, hashed, m.Signature, &opts)
	if err != nil {
		return nil, e.Push(err, "can't verify the signature")
	}
	return plainText, nil
}
