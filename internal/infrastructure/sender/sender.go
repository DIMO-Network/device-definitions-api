package sender

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

type Sender interface {
	Address() common.Address
	Sign(ctx context.Context, hash common.Hash) ([]byte, error)
}
