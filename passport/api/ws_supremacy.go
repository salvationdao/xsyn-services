package api

import (
	"math/big"
	"xsyn-services/types"

	"github.com/ninja-software/log_helpers"
	"github.com/sasha-s/go-deadlock"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"
)

type TickerPoolCache struct {
	outerMx            deadlock.Mutex
	nextAccessMx       deadlock.Mutex
	dataMx             deadlock.Mutex
	TricklingAmountMap map[string]*big.Int
}

// SupremacyControllerWS holds handlers for supremacy and the supremacy held transactions
type SupremacyControllerWS struct {
	Conn            *pgxpool.Pool
	Log             *zerolog.Logger
	API             *API
	TickerPoolCache *TickerPoolCache

	Txs *Transactions
}

type Transactions struct {
	Txes []*types.NewTransaction
	TxMx deadlock.Mutex
}

// NewSupremacyController creates the supremacy hub
func NewSupremacyController(log *zerolog.Logger, api *API) *SupremacyControllerWS {
	supremacyHub := &SupremacyControllerWS{
		Log: log_helpers.NamedLogger(log, "supremacy"),
		API: api,
		TickerPoolCache: &TickerPoolCache{
			outerMx:            deadlock.Mutex{},
			nextAccessMx:       deadlock.Mutex{},
			dataMx:             deadlock.Mutex{},
			TricklingAmountMap: make(map[string]*big.Int),
		},
		Txs: &Transactions{
			Txes: []*types.NewTransaction{},
		},
	}

	return supremacyHub
}
