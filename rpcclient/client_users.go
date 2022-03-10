package rpcclient

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gofrs/uuid"
	null "github.com/volatiletech/null/v8"
)

type PlayerRegisterReq struct {
	ID            uuid.UUID
	Username      string
	FactionID     uuid.UUID
	PublicAddress common.Address
}
type PlayerRegisterResp struct {
	ID            string
	FactionID     null.String
	Username      null.String
	PublicAddress null.String
	IsAi          bool
	DeletedAt     null.Time
	UpdatedAt     time.Time
	CreatedAt     time.Time
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
