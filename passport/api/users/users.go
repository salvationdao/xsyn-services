package users

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"xsyn-services/boiler"
	"xsyn-services/passport/crypto"
	"xsyn-services/passport/db"
	"xsyn-services/passport/email"
	"xsyn-services/passport/helpers"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/passport/supremacy_rpcclient"
	"xsyn-services/types"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gofrs/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type Fingerprint struct {
	VisitorID  string  `json:"visitor_id"`
	OSCPU      string  `json:"os_cpu"`
	Platform   string  `json:"platform"`
	Timezone   string  `json:"timezone"`
	Confidence float32 `json:"confidence"`
	UserAgent  string  `json:"user_agent"`
}

type UserGetter struct {
	Log    *zerolog.Logger
	Mailer *email.Mailer
}

var factions map[string]*boiler.Faction
var rlock sync.RWMutex
var once sync.Once

func Faction(id string) *boiler.Faction {
	if id == "" {
		return nil
	}
	once.Do(func() {
		factions = map[string]*boiler.Faction{}
		factionsAll, err := boiler.Factions().All(passdb.StdConn)
		if err != nil {
			passlog.L.Fatal().Err(err).Msg("unable to load factions from database")
		}

		for _, f := range factionsAll {
			factions[f.ID] = f
		}
	})
	rlock.RLock()
	defer rlock.RUnlock()

	return factions[id]
}

func FacebookID(s string) (*boiler.User, error) {
	user, err := boiler.Users(boiler.UserWhere.FacebookID.EQ(null.StringFrom(s))).One(passdb.StdConn)
	if err != nil {
		return nil, err
	}

	if user.FactionID.Valid {
		user.L.LoadFaction(passdb.StdConn, true, user, nil)
	}

	return user, nil
}

func GoogleID(s string) (*boiler.User, error) {
	user, err := boiler.Users(boiler.UserWhere.GoogleID.EQ(null.StringFrom(s))).One(passdb.StdConn)
	if err != nil {
		return nil, err
	}

	if user.FactionID.Valid {
		user.L.LoadFaction(passdb.StdConn, true, user, nil)
	}

	return user, nil
}

func TwitchID(s string) (*boiler.User, error) {
	user, err := boiler.Users(boiler.UserWhere.TwitchID.EQ(null.StringFrom(s))).One(passdb.StdConn)
	if err != nil {
		return nil, err
	}

	if user.FactionID.Valid {
		user.L.LoadFaction(passdb.StdConn, true, user, nil)
	}

	return user, nil
}

func TwitterID(s string) (*boiler.User, error) {
	user, err := boiler.Users(boiler.UserWhere.TwitterID.EQ(null.StringFrom(s))).One(passdb.StdConn)
	if err != nil {
		return nil, err
	}

	if user.FactionID.Valid {
		user.L.LoadFaction(passdb.StdConn, true, user, nil)
	}

	return user, nil
}

func DiscordID(s string) (*boiler.User, error) {
	user, err := boiler.Users(boiler.UserWhere.DiscordID.EQ(null.StringFrom(s))).One(passdb.StdConn)
	if err != nil {
		return nil, err
	}

	if user.FactionID.Valid {
		user.L.LoadFaction(passdb.StdConn, true, user, nil)
	}

	return user, nil
}

// UserExists checks if the User row exists.
func UserExists(email string) (bool, error) {
	var exists bool
	sql := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL LIMIT 1)`

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, strings.ToLower(email))
	}
	row := passdb.StdConn.QueryRow(sql, email)

	err := row.Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("boiler: unable to check if users exists: %v", err)
	}

	return exists, nil
}

func UserCreator(firstName, lastName, username, email, facebookID, googleID, twitchID, twitterID, discordID, phNumber string, publicAddress common.Address, password string, other ...interface{}) (*types.User, error) {
	if password != "" {
		err := helpers.IsValidPassword(password)
		if err != nil {
			return nil, err
		}

	}

	throughOauth := true
	if facebookID == "" && googleID == "" && publicAddress.Hex() == "" && twitchID == "" && twitterID == "" && discordID == "" {
		if email == "" {
			return nil, terror.Error(fmt.Errorf("email empty"), "Email cannot be empty")
		}

		throughOauth = false

		err := helpers.IsValidPassword(password)
		if err != nil {
			return nil, err
		}

		emailNotAvailable, err := UserExists(email)
		if err != nil {
			return nil, terror.Error(err, "Something went wrong. Please try again.")
		}
		if emailNotAvailable {
			return nil, terror.Error(fmt.Errorf("user already exists"), "A user with that email already exists. Perhaps you'd like to login instead?")
		}
	}

	trimmedUsername := "noob-" + username
	bm := bluemonday.StrictPolicy()
	sanitizedUsername := bm.Sanitize(trimmedUsername)

	err := helpers.IsValidUsername(sanitizedUsername)
	if err != nil {
		return nil, err
	}

	usExists, _ := boiler.Users(boiler.UserWhere.Username.EQ(strings.ToLower(sanitizedUsername))).One(passdb.StdConn)

	n := 1
	for usExists != nil {
		sanitizedUsername = helpers.RandStringBytes(n) + sanitizedUsername
		n++
		usExists, _ = boiler.Users(boiler.UserWhere.Username.EQ(strings.ToLower(sanitizedUsername))).One(passdb.StdConn)
		if n > 10 {
			return nil, fmt.Errorf("unable to generate a unique username")
		}
	}
	hexPublicAddress := ""
	if publicAddress != common.HexToAddress("") {
		hexPublicAddress = publicAddress.Hex()
	}

	user := &boiler.User{
		FirstName:     null.StringFrom(firstName),
		LastName:      null.StringFrom(lastName),
		Username:      sanitizedUsername,
		FacebookID:    types.NewString(facebookID),
		GoogleID:      types.NewString(googleID),
		TwitchID:      types.NewString(twitchID),
		TwitterID:     types.NewString(twitterID),
		DiscordID:     types.NewString(discordID),
		Email:         types.NewString(email),
		PublicAddress: types.NewString(hexPublicAddress),
		RoleID:        types.NewString(types.UserRoleMemberID.String()),
		Verified:      throughOauth, // verify users directly if they go through Oauth
	}

	err = user.Insert(passdb.StdConn, boil.Infer())
	if err != nil {
		passlog.L.Error().Err(err).Msg("insert new user failed")
		return nil, terror.Error(err, "create new user failed")
	}

	_ = supremacy_rpcclient.PlayerRegister(
		uuid.Must(uuid.FromString(user.ID)), user.Username, uuid.Nil, publicAddress)

	if password != "" && email != "" {
		pw := &boiler.PasswordHash{
			UserID:       user.ID,
			PasswordHash: crypto.HashPassword(password),
		}

		err := pw.Insert(passdb.StdConn, boil.Infer())
		if err != nil {
			return nil, err
		}

		return &types.User{User: user}, nil
	}

	return &types.User{User: user}, nil
}

func FingerprintUpsert(fingerprint Fingerprint, userID string) error {
	// Attempt to find fingerprint or create one
	fingerprintExists, err := boiler.Fingerprints(boiler.FingerprintWhere.VisitorID.EQ(fingerprint.VisitorID)).Exists(passdb.StdConn)
	if err != nil {
		return err
	}

	if !fingerprintExists {
		fp := boiler.Fingerprint{
			VisitorID:  fingerprint.VisitorID,
			OsCPU:      null.StringFrom(fingerprint.OSCPU),
			Platform:   null.StringFrom(fingerprint.Platform),
			Timezone:   null.StringFrom(fingerprint.Timezone),
			Confidence: decimal.NewNullDecimal(decimal.NewFromFloat32(fingerprint.Confidence)),
			UserAgent:  null.StringFrom(fingerprint.UserAgent),
		}
		err = fp.Insert(passdb.StdConn, boil.Infer())
		if err != nil {
			return err
		}
	}

	f, err := boiler.Fingerprints(boiler.FingerprintWhere.VisitorID.EQ(fingerprint.VisitorID)).One(passdb.StdConn)
	if err != nil {
		return err
	}

	// Link fingerprint to user
	userFingerprintExists, err := boiler.UserFingerprints(boiler.UserFingerprintWhere.UserID.EQ(userID), boiler.UserFingerprintWhere.FingerprintID.EQ(f.ID)).Exists(passdb.StdConn)
	if err != nil {
		return err
	}
	if !userFingerprintExists {
		// User fingerprint does not exist; create one
		newUserFingerprint := boiler.UserFingerprint{
			UserID:        userID,
			FingerprintID: f.ID,
		}
		err = newUserFingerprint.Insert(passdb.StdConn, boil.Infer())
		if err != nil {
			return err
		}
	}

	return nil
}

func PublicAddress(s common.Address) (*types.User, error) {
	user, err := boiler.Users(
		boiler.UserWhere.PublicAddress.EQ(null.StringFrom(s.Hex())),
		qm.Load(qm.Rels(boiler.UserRels.Faction)),
	).One(passdb.StdConn)
	if err != nil {
		return nil, err
	}
	return types.UserFromBoil(user)
}

func UUID(id uuid.UUID) (*types.User, error) {
	return ID(id.String())
}

func ID(id string) (*types.User, error) {
	user, err := boiler.Users(
		boiler.UserWhere.ID.EQ(id),
		qm.Load(qm.Rels(boiler.UserRels.Faction)),
	).One(passdb.StdConn)
	if err != nil {
		return nil, err
	}

	return types.UserFromBoil(user)
}

func Token(id uuid.UUID) (*types.User, error) {
	tokenUUID := types.IssueTokenID(id)
	result, err := boiler.FindIssueToken(passdb.StdConn, tokenUUID.String())
	if err != nil {
		return nil, err
	}

	user, err := boiler.Users(
		boiler.UserWhere.ID.EQ(result.UserID),
		qm.Load(qm.Rels(boiler.UserRels.Faction)),
	).One(passdb.StdConn)
	if err != nil {
		return nil, err
	}

	return types.UserFromBoil(user)
}

func Email(email string) (*types.User, error) {
	user, err := boiler.Users(
		boiler.UserWhere.Email.EQ(null.StringFrom(strings.ToLower(email))),
		qm.Load(qm.Rels(boiler.UserRels.Faction)),
	).One(passdb.StdConn)
	if err != nil {
		return nil, err
	}

	return types.UserFromBoil(user)
}

func EmailPassword(email string, password string) (*types.User, error) {

	errMsg := "invalid email or password, please try again."

	user, err := boiler.Users(
		boiler.UserWhere.Email.EQ(null.StringFrom(strings.ToLower(email))),
		qm.Load(qm.Rels(boiler.UserRels.Faction)),
	).One(passdb.StdConn)

	if err != nil {
		return nil, fmt.Errorf(errMsg)
	}

	userPassword, err := db.HashByUserID(user.ID)

	if err != nil {
		return nil, err
	}

	err = crypto.ComparePassword(userPassword, password)

	if err != nil {
		return nil, fmt.Errorf(errMsg)
	}

	return types.UserFromBoil(user)
}

func Username(uname string) (*boiler.User, string, error) {
	user, err := boiler.Users(boiler.UserWhere.Username.EQ(strings.ToLower(uname))).One(passdb.StdConn)
	if err != nil {
		return nil, "", err
	}

	hash, err := boiler.FindPasswordHash(passdb.StdConn, user.ID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, "", err
		}
	}

	if user.FactionID.Valid {
		err = user.L.LoadFaction(passdb.StdConn, true, user, nil)
		return nil, "", err
	}
	return user, hash.PasswordHash, nil

}
