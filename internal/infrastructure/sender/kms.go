package sender

import (
	"context"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"math/big"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"golang.org/x/exp/slices"
)

var secp256k1N = crypto.S256().Params().N
var secp256k1HalfN = new(big.Int).Div(secp256k1N, big.NewInt(2))

type kmsSender struct {
	keyID   string
	pub     []byte
	address common.Address
	client  *kms.Client
}

func (s *kmsSender) Address() common.Address {
	return s.address
}

func (s *kmsSender) Sign(ctx context.Context, hash common.Hash) ([]byte, error) {
	signInput := &kms.SignInput{
		KeyId:            aws.String(s.keyID),
		SigningAlgorithm: types.SigningAlgorithmSpecEcdsaSha256,
		MessageType:      types.MessageTypeDigest,
		Message:          hash[:],
	}

	out, err := s.client.Sign(ctx, signInput)
	if err != nil {
		return nil, err
	}

	sig := new(ecdsaSigFields)
	_, err = asn1.Unmarshal(out.Signature, sig)
	if err != nil {
		return nil, err
	}

	sigR, sigS := sig.R.Bytes, sig.S.Bytes

	// Correct S, if necessary, so that it's in the lower half of the group.
	sigSNum := new(big.Int).SetBytes(sigS)
	if sigSNum.Cmp(secp256k1HalfN) > 0 {
		sigS = new(big.Int).Sub(secp256k1N, sigSNum).Bytes()
	}

	// Determine whether V ought to be 0 or 1.
	sigRS := append(fixLen(sigR), fixLen(sigS)...)
	sigRSV := append(sigRS, 0)

	recPub, err := crypto.Ecrecover(hash[:], sigRSV)
	if err != nil {
		return nil, err
	}

	if slices.Equal(recPub, s.pub) {
		return sigRSV, nil
	}

	sigRSV = append(sigRS, 1)
	recPub, err = crypto.Ecrecover(hash[:], sigRSV)
	if err != nil {
		return nil, err
	}

	if slices.Equal(recPub, s.pub) {
		return sigRSV, nil
	}

	return nil, fmt.Errorf("couldn't choose a working V from the returned R and S")
}

func fixLen(in []byte) []byte {
	outStart := 0
	inLen := len(in)
	inStart := 0

	if inLen > 32 {
		inStart = inLen - 32
	} else if inLen < 32 {
		outStart = 32 - inLen
	}

	out := make([]byte, common.HashLength)
	copy(out[outStart:], in[inStart:])
	return out
}

type publicKeyFields struct {
	Algorithm        pkix.AlgorithmIdentifier
	SubjectPublicKey asn1.BitString
}

type ecdsaSigFields struct {
	R asn1.RawValue
	S asn1.RawValue
}

func FromKMS(ctx context.Context, client *kms.Client, keyID string) (Sender, error) {
	pubResp, err := client.GetPublicKey(ctx, &kms.GetPublicKeyInput{KeyId: aws.String(keyID)})
	if err != nil {
		return nil, err
	}

	cert := new(publicKeyFields)

	_, err = asn1.Unmarshal(pubResp.PublicKey, cert)
	if err != nil {
		return nil, err
	}

	pub, err := crypto.UnmarshalPubkey(cert.SubjectPublicKey.Bytes)
	if err != nil {
		return nil, err
	}

	pubBytes := secp256k1.S256().Marshal(pub.X, pub.Y)

	addr := crypto.PubkeyToAddress(*pub)

	return &kmsSender{keyID: keyID, pub: pubBytes, address: addr, client: client}, nil
}
