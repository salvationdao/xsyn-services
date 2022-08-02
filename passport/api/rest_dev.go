package api

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/nft1155"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/passport/payments"
	"xsyn-services/passport/supremacy_rpcclient"
	"xsyn-services/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/friendsofgo/errors"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type Dev struct {
	userCacheMap *Transactor
	R            *chi.Mux
}

func DevRoutes(userCacheMap *Transactor) *Dev {
	dev := &Dev{
		R:            chi.NewRouter(),
		userCacheMap: userCacheMap,
	}

	dev.R.Get("/give-mechs/{public_address}", WithError(WithDev(dev.devGiveMechs)))

	return dev
}

// WithDev checks that dev key is in the header and environment is development.
func WithDev(next func(w http.ResponseWriter, r *http.Request) (int, error)) func(w http.ResponseWriter, r *http.Request) (int, error) {
	fn := func(w http.ResponseWriter, r *http.Request) (int, error) {
		if os.Getenv("PASSPORT_ENVIRONMENT") != "development" {
			passlog.L.Warn().Err(terror.ErrUnauthorised).Str("os.Getenv(\"PASSPORT_ENVIRONMENT\")", os.Getenv("PASSPORT_ENVIRONMENT")).Msg("dev endpoint attempted in non dev environment")
			return http.StatusUnauthorized, terror.Error(terror.ErrUnauthorised, "Unauthorized.")
		}
		devPass := r.Header.Get("X-Authorization")
		if devPass != "NinjaDojo_!" {
			passlog.L.Warn().Err(terror.ErrUnauthorised).Str("devPass", devPass).Msg("unauthed attempted at dev rest end point")
			return http.StatusUnauthorized, terror.Error(terror.ErrUnauthorised, "Unauthorized.")
		}

		return next(w, r)
	}
	return fn
}

func (d *Dev) devGiveMechs(w http.ResponseWriter, r *http.Request) (int, error) {
	publicAddress := common.HexToAddress(chi.URLParam(r, "public_address"))
	user, err := payments.CreateOrGetUser(publicAddress)
	if err != nil {
		return http.StatusBadRequest, err
	}
	// get 3 random templates
	storeItems, err := boiler.StoreItems().All(passdb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	for i := range storeItems {
		j := rand.Intn(i + 1)
		storeItems[i], storeItems[j] = storeItems[j], storeItems[i]
	}

	for i, si := range storeItems {
		if i < 3 {
			_, err = db.PurchasedItemRegister(uuid.Must(uuid.FromString(si.ID)), uuid.Must(uuid.FromString(user.ID)))
			if err != nil {
				return http.StatusInternalServerError, err
			}
		} else {
			break
		}
	}

	// give account some suppies
	tx := &types.NewTransaction{
		Credit:               user.ID,
		Debit:                types.XsynSaleUserID.String(),
		Amount:               decimal.New(10000, 18),
		TransactionReference: types.TransactionReference(fmt.Sprintf("DEV SEED SUPS - %v", time.Now().UnixNano())),
		Description:          "Dev Seed Sups",
		Group:                "SEED",
		ServiceID:            types.XsynSaleUserID,
	}

	_, err = d.userCacheMap.Transact(tx)
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to give dev sups")
		return 0, err
	}

	collections, err := boiler.Collections(
		boiler.CollectionWhere.ContractType.EQ(null.StringFrom("EIP-1155")),
	).All(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to get collections")
		return http.StatusInternalServerError, err
	}

	// give account some keycardies
	for _, col := range collections {
		for _, tid := range col.ExternalTokenIds {
			asset, err := nft1155.CreateOrGet1155Asset(int(tid), user, col.Slug)
			if err != nil {
				passlog.L.Error().Err(err).Msg("failed to CreateOrGet1155Asset")
				return http.StatusInternalServerError, terror.Error(err, "Failed to create or get 1155 asset")
			}
			asset.Count += 5

			_, err = asset.Update(passdb.StdConn, boil.Whitelist(boiler.UserAssets1155Columns.Count))
			if err != nil {
				passlog.L.Error().Err(err).Msg("failed to update asset")
				return http.StatusInternalServerError, terror.Error(err, "Failed to update assets count")
			}
			err = supremacy_rpcclient.KeycardsTransferToSupremacy(types.UserAsset1155FromBoiler(asset), 0, 5)
			if err != nil {
				passlog.L.Error().Err(err).Msg("failed to transfer to supremacy")
				return http.StatusInternalServerError, terror.Error(err, "Failed to update assets count on supremacy")
			}
			offXsynAsset, err := boiler.UserAssets1155S(
				boiler.UserAssets1155Where.ExternalTokenID.EQ(int(tid)),
				boiler.UserAssets1155Where.OwnerID.EQ(user.ID),
				boiler.UserAssets1155Where.ServiceID.EQ(null.StringFrom(types.SupremacyGameUserID.String())),
			).One(passdb.StdConn)
			if errors.Is(err, sql.ErrNoRows) {
				offXsynAsset = &boiler.UserAssets1155{
					OwnerID:         asset.OwnerID,
					CollectionID:    asset.CollectionID,
					ExternalTokenID: asset.ExternalTokenID,
					Label:           asset.Label,
					Description:     asset.Description,
					Count:           5,
					ImageURL:        asset.ImageURL,
					AnimationURL:    asset.AnimationURL,
					KeycardGroup:    asset.KeycardGroup,
					Attributes:      asset.Attributes,
					ServiceID:       null.StringFrom(types.SupremacyGameUserID.String()),
				}
				if err := offXsynAsset.Insert(passdb.StdConn, boil.Infer()); err != nil {
					passlog.L.Error().Err(err).Msg("failed insert new keycards")
					return http.StatusInternalServerError, terror.Error(err, "Failed to update assets count on supremacy")
				}
			}
		}
	}

	return http.StatusOK, nil
}
