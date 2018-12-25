package main

import (
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
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