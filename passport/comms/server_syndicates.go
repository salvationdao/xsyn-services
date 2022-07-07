package comms

import (
	"fmt"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/api/users"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/types"

	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/shopspring/decimal"

	"github.com/gofrs/uuid"

	"github.com/ninja-software/terror/v2"
)

// SyndicateCreateHandler request an ownership transfer of an asset
func (s *S) SyndicateCreateHandler(req SyndicateCreateReq, resp *SyndicateCreateResp) error {
	serviceID, err := IsServerClient(req.ApiKey)
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to get service id - AssetTransferOwnershipHandler")
		return err
	}
	user, err := users.UUID(uuid.FromStringOrNil(req.FoundedByID))
	if err != nil {
		return err
	}

	if !user.FactionID.Valid {
		return fmt.Errorf("user does not have faction")
	}

	isLocked := user.CheckUserIsLocked("account")
	if isLocked {
		return terror.Error(fmt.Errorf("user: %s attempting to purchase on Supremacy while locked", user.ID), "This account is locked, contact support to unlock.")
	}

	tx, err := passdb.StdConn.Begin()
	if err != nil {
		passlog.L.Error().Err(err).Msg("Failed to begin db transaction")
		return terror.Error(err, "Failed to register syndicate in Xsyn")
	}

	defer tx.Rollback()

	// create an account for the syndicate
	account := boiler.Account{
		Type: boiler.AccountTypeSYNDICATE,
		Sups: decimal.Zero,
	}

	err = account.Insert(tx, boil.Infer())
	if err != nil {
		passlog.L.Error().Err(err).Interface("account", account).Msg("Failed to create syndicate account.")
		return terror.Error(err, "Failed to create syndicate account")
	}

	// create syndicate
	syndicate := boiler.Syndicate{
		ID:          req.ID,
		FoundedByID: req.FoundedByID,
		FactionID:   user.FactionID.String,
		Name:        req.Name,
		AccountID:   account.ID,
	}

	err = syndicate.Insert(tx, boil.Infer())
	if err != nil {
		passlog.L.Error().Err(err).Interface("syndicate", syndicate).Msg("Failed to insert syndicate into db")
		return terror.Error(err, "Failed to register syndicate in Xsyn")
	}

	nt := &types.NewTransaction{
		Debit:                req.FoundedByID,
		Credit:               types.XsynTreasuryUserID.String(), // TODO: check where does the fund goes?
		TransactionReference: types.TransactionReference(fmt.Sprintf("syndicate_create|SUPREMACY|%s|%d", req.ID, time.Now().UnixNano())),
		Description:          "Start a new syndicate",
		Amount:               decimal.New(500, 18), // TODO: calculate how much is 500 usd worth of sups
		Group:                "Syndicate",
		SubGroup:             "syndicate create",
		ServiceID:            types.UserID(uuid.FromStringOrNil(serviceID)),
	}

	_, err = s.UserCacheMap.Transact(nt)
	if err != nil {
		return terror.Error(err, err.Error())
	}

	err = tx.Commit()
	if err != nil {
		passlog.L.Error().Err(err).Msg("Failed to commit db transaction")
		return terror.Error(err, "Failed to register syndicate in Xsyn")
	}

	return nil
}
