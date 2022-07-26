package helpers

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ninja-software/terror/v2"
)

func GenerateNewWallet() (string, string, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return "", "", terror.Error(err)
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateHex := hexutil.Encode(privateKeyBytes)[2:]

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", "", terror.Error(fmt.Errorf("could not create a public key"), "Could not create a public key for user.")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

	return address, privateHex, nil
}
