package comms

import (
	"context"
	"passport"
	"passport/db"
	"passport/passdb"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gofrs/uuid"
	"github.com/volatiletech/null/v8"
)

type UserReq struct {
	ID uuid.UUID
}
type UserResp struct {
	ID            uuid.UUID
	Username      string
	FactionID     null.String
	PublicAddress common.Address
}

func (c *S) User(req UserReq, resp *UserResp) error {
	ctx := context.Background()
	u, err := db.UserGet(ctx, passdb.Conn, passport.UserID(req.ID))
	if err != nil {
		return err
	}
	resp.FactionID = null.NewString(u.FactionID.String(), u.FactionID == nil)
	resp.ID = uuid.UUID(u.ID)
	resp.PublicAddress = common.HexToAddress(u.PublicAddress.String)
	resp.Username = u.Username

	return nil
}
