package sender

import (
	"context"
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type privateKeySender struct {
	privateKey *ecdsa.PrivateKey
	address    common.Address
}

func (a *privateKeySender) Address() common.Address {
	return a.address
}

func (a *privateKeySender) Sign(_ context.Context, hash common.Hash) ([]byte, error) {
	return crypto.Sign(hash[:], a.privateKey)
}

func FromKey(hexPriv string) (Sender, error) {
	privBytes := common.FromHex(hexPriv)
	priv, err := crypto.ToECDSA(privBytes)
	if err != nil {
		return nil, err
	}

	address := crypto.PubkeyToAddress(priv.PublicKey)

	return &privateKeySender{privateKey: priv, address: address}, nil
}
