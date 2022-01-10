package deposit

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// LogTransfer ..
type LogTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
}
