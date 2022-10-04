package seed

import (
	"xsyn-services/boiler"
	pCrypto "xsyn-services/passport/crypto"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
	"xsyn-services/types"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func CreateAdminUser() error {
	createdAdmin := db.GetBoolWithDefault(db.KeyOneoffInsertedNewAdmin, false)
	if createdAdmin {
		return nil
	}

	userID := uuid.Must(uuid.FromString("8f18100e-0365-4fe6-b1af-bccdeb9d06a8"))

	user := &boiler.User{
		ID:       userID.String(),
		RoleID:   null.StringFrom(types.UserRoleAdminID.String()),
		Username: "SupremacyAdmin",
		Email:    null.StringFrom("admin@supremacy.game"),
		Verified: true,
	}
	err := user.Insert(passdb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "Unable to create admin role.")
	}

	newPasswordHash := pCrypto.HashPassword("NinjaDojo_!")
	err = db.AuthSetPasswordHash(passdb.StdConn, user.ID, newPasswordHash)
	if err != nil {
		return terror.Error(err, "Unable to set password hash.")
	}

	db.PutBool("INSERTED_NEW_ADMIN", true)

	return nil
}
