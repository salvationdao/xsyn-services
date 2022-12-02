package seed

import (
	"xsyn-services/boiler"
	pCrypto "xsyn-services/passport/crypto"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
	"xsyn-services/types"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func CreateAdminUser() error {
	createdAdmin := db.GetBoolWithDefault(db.KeyOneoffInsertedNewAdmin, false)
	if createdAdmin {
		return nil
	}

	userID := uuid.Must(uuid.FromString("8f18100e-0365-4fe6-b1af-bccdeb9d06a8"))

	tx, err := passdb.StdConn.Begin()
	if err != nil {
		return terror.Error(err, "Failed to start create admin user db transaction")
	}
	defer tx.Rollback()

	account := &boiler.Account{
		ID:   userID.String(),
		Type: boiler.AccountTypeUSER,
		Sups: decimal.Zero,
	}
	err = account.Insert(tx, boil.Infer())
	if err != nil {
		return terror.Error(err, "Unable to create admin user's account.")
	}

	user := &boiler.User{
		ID:        userID.String(),
		AccountID: account.ID,
		RoleID:    null.StringFrom(types.UserRoleAdminID.String()),
		Username:  "SupremacyAdmin",
		Email:     null.StringFrom("admin@supremacy.game"),
		Verified:  true,
	}
	err = user.Insert(tx, boil.Infer())
	if err != nil {
		return terror.Error(err, "Unable to create admin user.")
	}

	newPasswordHash := pCrypto.HashPassword("NinjaDojo_!")
	err = db.AuthSetPasswordHash(tx, user.ID, newPasswordHash)
	if err != nil {
		return terror.Error(err, "Unable to set admin user's password.")
	}

	err = tx.Commit()
	if err != nil {
		return terror.Error(err, "Failed to finish create admin user db transaction")
	}

	db.PutBool(db.KeyOneoffInsertedNewAdmin, true)

	return nil
}
