package comms

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ninja-software/terror/v2"
)

type UserReq struct {
	Addr common.Address
}
type UserResp struct{}

func (c *S) User(req UserReq, resp *UserResp) error {
	return terror.ErrNotImplemented
}
