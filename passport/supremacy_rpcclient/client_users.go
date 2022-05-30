package supremacy_rpcclient

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofrs/uuid"
	"xsyn-services/passport/passlog"
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
	err := SupremacyClient.Call("S.PlayerRegisterHandler", &PlayerRegisterReq{UserID, Username, FactionID, PublicAddress}, resp)
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to insert player to supremacy server")
		return err
	}

	return nil
}
