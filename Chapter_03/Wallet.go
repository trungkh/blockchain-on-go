package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	//"crypto/sha1"
	//"encoding/asn1"
	//"encoding/hex"
	"log"

	"math/big"
)

type Wallet struct {
	publicKey  *ecdsa.PublicKey
	privateKey *ecdsa.PrivateKey
}

func (this *Wallet) init() {
	this.privateKey, this.publicKey = this.genKeyPair()
}

func (this *Wallet) genKeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
	curve := elliptic.P256()
	privKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	pubKey := &privKey.PublicKey

	return privKey, pubKey
}

func (this Wallet) bytesToPrivateKey(privateKey []byte)  *ecdsa.PrivateKey {
    k := new(big.Int)
    k.SetBytes(privateKey)

	curve := elliptic.P256()
    priv := new(ecdsa.PrivateKey)
    priv.D = k
    priv.PublicKey.Curve = curve
    priv.PublicKey.X, priv.PublicKey.Y = curve.ScalarBaseMult(k.Bytes())

    return priv
}

func (this Wallet) bytesToPublicKey(publicKey []byte) *ecdsa.PublicKey {
	x := new(big.Int)
	y := new(big.Int)
    keyLen := len(publicKey)
    x.SetBytes(publicKey[:(keyLen / 2)])
    y.SetBytes(publicKey[(keyLen / 2):])

    pubKey := new(ecdsa.PublicKey)
    pubKey.X = x
    pubKey.Y = y
    pubKey.Curve = elliptic.P256()
    return pubKey
}

func (this Wallet) getPublicKey() []byte {
	return append(this.publicKey.X.Bytes(), this.publicKey.Y.Bytes()...)
}

func (this Wallet) getPrivateKey() []byte {
	return this.privateKey.D.Bytes()
}

func (this *Wallet) sign(message []byte, privateKey []byte) []byte {
	priv := this.bytesToPrivateKey(privateKey)
	r, s, _ := ecdsa.Sign(rand.Reader, priv, message)
	return append(r.Bytes(), s.Bytes()...)
}

func (this *Wallet) verifySignature(message []byte, signature []byte, publicKey []byte) bool {
	r := new(big.Int)
	s := new(big.Int)
    sigLen := len(signature)
    r.SetBytes(signature[:(sigLen / 2)])
    s.SetBytes(signature[(sigLen / 2):])

	pubKey := this.bytesToPublicKey(publicKey)

	return ecdsa.Verify(pubKey, message, r, s)
}

/*func (this Wallet) testSigning() {
	privByte, _ := hex.DecodeString("9bda80432dbd72a7a20f9411fb9fb5c4cee2021ffe7d869f6199878606cadf45")
	log.Printf("Private Key: %x", privByte)
	sigByte := this.sign([]byte("hello"), privByte)
	log.Printf("Signature:   %x", sigByte)

	pubByte, _ := hex.DecodeString("4b83487732a84f3963bd20f61341a1a69fd9d5db6be47d0f9d92015baf8848b3beb0c447ed24b7e0b5adc310da9b6cc5f482c53bf04508f72dd7cd4818006906")
    ok := this.verifySignature([]byte("hello"), sigByte, pubByte)
	log.Println(ok)
}

func (this *Wallet) _sign(message []byte, privateKey []byte) []byte {
	digest := sha1.Sum(message)
	priv := this.bytesToPrivateKey(privateKey)

	var esig struct { R, S *big.Int }
	esig.R, esig.S, _ = ecdsa.Sign(rand.Reader, priv, digest[:])
	log.Printf("R: %x", esig.R)
	log.Printf("S: %x", esig.S)

	signature, _ := asn1.Marshal(esig)
	return signature
}

func (this *Wallet) _verifySignature(message []byte, signature []byte, publicKey []byte) bool {
    digest := sha1.Sum(message)
 
	var esig struct { R, S *big.Int }
	asn1.Unmarshal(signature, &esig)
	log.Printf("R: %x", esig.R)
	log.Printf("S: %x", esig.S)

	pubKey := this.bytesToPublicKey(publicKey)

	return ecdsa.Verify(pubKey, digest[:], esig.R, esig.S)
}*/