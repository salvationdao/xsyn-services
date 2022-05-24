package api

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"math/rand"
	"net/http"
	"os"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/passport/payments"
)

func DevRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/give-mechs/{public_address}", WithError(WithDev(devGiveMechs)))


	return r
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


func devGiveMechs(w http.ResponseWriter, r *http.Request) (int, error) {
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

	return http.StatusOK, nil
}