package auth

import (
	"crypto/rsa"
	"fmt"
)

var RefreshSecret []byte
var PublicKey *rsa.PublicKey

func init() {
	// Load RSA Public Key
	publicKey, err := LoadPublicKey("token_public.pem")
	if err != nil {
		panic(fmt.Errorf("failed to load public key: %w", err))
	}
	PublicKey = publicKey
}
