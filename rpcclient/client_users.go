package rpcclient

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofrs/uuid"
)

type PlayerRegisterReq struct {
	ID            uuid.UUID
	Username      string
	FactionID     uuid.UUID
	PublicAddress common.Address
}
type PlayerRegisterResp struct {
	ID uuid.UUID
}

func PlayerRegister(
	UserID uuid.UUID,
	Username string,
	FactionID uuid.UUID,
	PublicAddress common.Address,
) error {
	resp := &PlayerRegisterResp{}
	err := Client.Call("S.PlayerRegister", &PlayerRegisterReq{UserID, Username, FactionID, PublicAddress}, resp)
	if err != nil {
		return err
	}

	return nil
}
