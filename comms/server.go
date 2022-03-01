package comms

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/rpc"
	"passport"
	"passport/api"
	"passport/db"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
)

type C struct {
	UserCacheMap *api.UserCacheMap
	MessageBus   *messagebus.MessageBus
	Txs          *api.Transactions
	Log          *zerolog.Logger
	Conn         db.Conn
	DistLock     sync.Mutex
}

type SpendSupsReq struct {
	Amount               string                        `json:"amount"`
	FromUserID           passport.UserID               `json:"userID"`
	TransactionReference passport.TransactionReference `json:"transactionReference"`
	GroupID              passport.TransactionGroup     `json:"groupID"`
}
type SpendSupsResp struct {
	TXID string `json:"txid"`
}

func New(
	userCacheMap *api.UserCacheMap,
	messageBus *messagebus.MessageBus,
	txs *api.Transactions,
	log *zerolog.Logger,
	conn *pgxpool.Pool,
) *C {
	result := &C{
		UserCacheMap: userCacheMap,
		MessageBus:   messageBus,
		Txs:          txs,
		Log:          log,
		Conn:         conn,
	}
	return result
}

func (c *C) listen(addrStr ...string) ([]net.Listener, error) {
	listeners := make([]net.Listener, len(addrStr))
	for i, a := range addrStr {
		c.Log.Info().Str("addr", a).Msg("registering RPC server")
		addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("0.0.0.0:%s", a))
		if err != nil {
			c.Log.Err(err).Str("addr", a).Msg("registering RPC server")
			return listeners, nil
		}

		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return listeners, err
		}

		listeners[i] = l
	}

	return listeners, nil
}

func Start(c *C) error {
	listeners, err := c.listen("10001", "10002", "10003", "10004", "10005", "10006")
	if err != nil {
		return err
	}
	for _, l := range listeners {
		s := rpc.NewServer()
		err = s.Register(c)
		if err != nil {
			return err
		}

		c.Log.Info().Str("addr", l.Addr().String()).Msg("starting up RPC server")
		go s.Accept(l)
	}

	return nil
}

func (c *C) SupremacySpendSupsHandler(req SpendSupsReq, resp *SpendSupsResp) error {
	ctx := context.Background()
	amt, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return err
	}
	if amt.LessThan(decimal.Zero) {
		return terror.Error(terror.ErrInvalidInput, "Sups amount can not be negative")
	}

	tx := &passport.NewTransaction{
		From:                 req.FromUserID,
		To:                   passport.SupremacyGameUserID,
		TransactionReference: req.TransactionReference,
		Amount:               *amt.BigInt(),
	}

	if req.GroupID != "" {
		tx.To = passport.SupremacyBattleUserID
		tx.GroupID = req.GroupID
	}

	nfb, ntb, txID, err := c.UserCacheMap.Process(tx)
	if err != nil {
		return terror.Error(err, "failed to process sups")
	}

	if !tx.From.IsSystemUser() {
		go c.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", api.HubKeyUserSupsSubscribe, tx.From)), nfb.String())
	}

	if !tx.To.IsSystemUser() {
		go c.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", api.HubKeyUserSupsSubscribe, tx.To)), ntb.String())
	}

	tx.ID = txID

	// for refund
	c.Txs.TxMx.Lock()
	c.Txs.Txes = append(c.Txs.Txes, &passport.NewTransaction{
		ID:                   txID,
		From:                 tx.To,
		To:                   tx.From,
		Amount:               tx.Amount,
		TransactionReference: passport.TransactionReference(fmt.Sprintf("refund|sups vote|%s", txID)),
	})
	c.Txs.TxMx.Unlock()

	resp.TXID = txID
	return nil
}

type ReleaseTransactionsReq struct {
	TxIDs []string `json:"txIDs"`
}
type ReleaseTransactionsResp struct{}

func (c *C) ReleaseTransactionsHandler(req ReleaseTransactionsReq, resp *ReleaseTransactionsResp) error {
	ctx := context.Background()
	c.Txs.TxMx.Lock()
	defer c.Txs.TxMx.Unlock()
	for _, txID := range req.TxIDs {
		for _, tx := range c.Txs.Txes {
			if txID != tx.ID {
				continue
			}
			nfb, ntb, _, err := c.UserCacheMap.Process(tx)
			if err != nil {
				c.Log.Err(err).Msg("failed to process user sups fund")
				continue
			}

			if !tx.From.IsSystemUser() {
				go c.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", api.HubKeyUserSupsSubscribe, tx.From)), nfb.String())
			}

			if !tx.To.IsSystemUser() {
				go c.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", api.HubKeyUserSupsSubscribe, tx.To)), ntb.String())
			}
		}
	}

	c.Txs.Txes = []*passport.NewTransaction{}

	return nil
}

type TickerTickReq struct {
	UserMap map[int][]passport.UserID `json:"userMap"`
}
type TickerTickResp struct{}

func (c *C) supremacyFeed() {
	fund := big.NewInt(0)
	fund, ok := fund.SetString("100000000000000000000", 10)
	if !ok {
		c.Log.Err(errors.New("setting string not ok on fund big int")).Msg("too many strings")
		return
	}

	tx := &passport.NewTransaction{
		From:                 passport.XsynTreasuryUserID,
		To:                   passport.SupremacySupPoolUserID,
		Amount:               *fund,
		TransactionReference: passport.TransactionReference(fmt.Sprintf("treasury|ticker|%s", time.Now())),
		GroupID:              passport.TransactionGroupBattleStream,
	}

	// process user cache map
	_, _, _, err := c.UserCacheMap.Process(tx)
	if err != nil {
		c.Log.Err(err).Msg(err.Error())
		return
	}
}

func (c *C) TickerTickHandler(req TickerTickReq, resp *TickerTickResp) error {
	// make treasury send game server user moneys
	c.supremacyFeed()

	// sups guard
	// kick users off the list, if they don't have any sups\
	um := make(map[passport.UserID]passport.FactionID)
	userMap := make(map[int][]passport.UserID)
	for multiplier, userIDs := range req.UserMap {
		newList := []passport.UserID{}

		for _, userID := range userIDs {
			amount, err := c.UserCacheMap.Get(userID.String())
			if err != nil || amount.BitLen() == 0 {
				// kick user out
				continue
			}
			um[userID] = passport.FactionID(uuid.Nil)
			newList = append(newList, userID)
		}

		if len(newList) > 0 {
			userMap[multiplier] = newList
		}
	}
	//  to avoid working in floats, a 100% multiplier is 100 points, a 25% is 25 points
	// This will give us what we need to divide the pool by and then times by to give the user the correct share of the pool

	err := db.GetFactionIDByUsers(context.Background(), c.Conn, um)
	if err != nil {
		return terror.Error(err)
	}

	// rebuild the sups disritbute system
	rmTotalPoint := 0
	rmTotalMap := make(map[int][]passport.UserID)
	bTotalPoint := 0
	bTotalMap := make(map[int][]passport.UserID)
	zTotalPoint := 0
	zTotalMap := make(map[int][]passport.UserID)

	// loop once to get total point count
	for multiplier, users := range userMap {
		for _, userID := range users {
			switch um[userID] {
			case passport.RedMountainFactionID:
				rmTotalPoint += multiplier

				// check user list
				if _, ok := rmTotalMap[multiplier]; !ok {
					rmTotalMap[multiplier] = []passport.UserID{}
				}
				rmTotalMap[multiplier] = append(rmTotalMap[multiplier], userID)

			case passport.BostonCyberneticsFactionID:
				bTotalPoint += multiplier

				// check user list
				if _, ok := bTotalMap[multiplier]; !ok {
					bTotalMap[multiplier] = []passport.UserID{}
				}
				bTotalMap[multiplier] = append(bTotalMap[multiplier], userID)

			case passport.ZaibatsuFactionID:
				zTotalPoint += multiplier

				// check set up separate user rate
				if _, ok := zTotalMap[multiplier]; !ok {
					zTotalMap[multiplier] = []passport.UserID{}
				}
				zTotalMap[multiplier] = append(zTotalMap[multiplier], userID)
			}
		}

	}

	// we take the whole balance of supremacy sup pool and give it to the users watching
	// amounts depend on their multiplier
	// the supremacy sup pool user gets sups trickled into it from the last battle and 4 every 5 seconds
	c.DistLock.Lock()
	defer c.DistLock.Unlock()
	supsForTick, err := c.UserCacheMap.Get(passport.SupremacySupPoolUserID.String())
	if err != nil {
		return terror.Error(err)
	}
	supPool := big.NewInt(0)
	supPool.Add(supPool, &supsForTick)
	supPool.Div(supPool, big.NewInt(3))

	// distribute Red Mountain sups
	c.DistrubuteFund(supPool.String(), int64(rmTotalPoint), rmTotalMap)

	// distribute Boston sups
	c.DistrubuteFund(supPool.String(), int64(bTotalPoint), bTotalMap)

	// distribute Zaibatsu sups
	c.DistrubuteFund(supPool.String(), int64(zTotalPoint), zTotalMap)

	return nil
}

func (c *C) DistrubuteFund(fundstr string, totalPoints int64, userMap map[int][]passport.UserID) {
	copiedFund := big.NewInt(0)
	copiedFund, ok := copiedFund.SetString(fundstr, 10)
	if !ok {
		c.Log.Err(fmt.Errorf("NOT work " + fundstr)).Msg(fundstr)
		return
	}

	if totalPoints <= 0 {
		return
	}

	totalPointsBigInt := big.NewInt(int64(totalPoints))

	// var transactions []*passport.NewTransaction
	onePointWorth := big.NewInt(0)
	onePointWorth = onePointWorth.Div(copiedFund, totalPointsBigInt)

	// loop again to create all transactions
	for multiplier, users := range userMap {
		for _, user := range users {
			usersSups := big.NewInt(0)
			usersSups = usersSups.Mul(onePointWorth, big.NewInt(int64(multiplier)))

			// if greater than 2 sups get 2 sups
			if usersSups.Cmp(big.NewInt(2000000000000000000)) >= 0 {
				usersSups = big.NewInt(2000000000000000000)
			}

			tx := &passport.NewTransaction{
				From:                 passport.SupremacySupPoolUserID,
				To:                   user,
				Amount:               *usersSups,
				TransactionReference: passport.TransactionReference(fmt.Sprintf("supremacy|ticker|%s|%s", user, time.Now())),
				GroupID:              passport.TransactionGroupBattleStream,
				Description:          "Watch to earn",
			}

			_, _, _, err := c.UserCacheMap.Process(tx)
			if err != nil {
				c.Log.Err(err).Msg("failed to process user fund")
				return
			}

			copiedFund = copiedFund.Sub(copiedFund, usersSups)
		}
	}
}

type GetSpoilOfWarReq struct{}
type GetSpoilOfWarResp struct {
	Amount string
}

func (c *C) SupremacyGetSpoilOfWarHandler(req GetSpoilOfWarReq, resp *GetSpoilOfWarResp) error {
	// get current sup pool user sups
	supsPoolUser, err := c.UserCacheMap.Get(passport.SupremacySupPoolUserID.String())
	if err != nil {
		return terror.Error(err)
	}

	battleUser, err := c.UserCacheMap.Get(passport.SupremacyBattleUserID.String())
	if err != nil {
		return terror.Error(err)
	}

	result := big.NewInt(0)
	result.Add(result, &supsPoolUser)
	result.Add(result, &battleUser)

	resp.Amount = result.String()
	return nil
}

type UserSupsMultiplierSendReq struct {
	UserSupsMultiplierSends []*UserSupsMultiplierSend `json:"userSupsMultiplierSends"`
}

type UserSupsMultiplierSend struct {
	ToUserID        passport.UserID   `json:"toUserID"`
	ToUserSessionID *hub.SessionID    `json:"toUserSessionID,omitempty"`
	SupsMultipliers []*SupsMultiplier `json:"supsMultiplier"`
}

type SupsMultiplier struct {
	Key       string    `json:"key"`
	Value     int       `json:"value"`
	ExpiredAt time.Time `json:"expiredAt"`
}

type UserSupsMultiplierSendResp struct{}

func (c *C) UserSupsMultiplierSendHandler(req UserSupsMultiplierSendReq, resp *UserSupsMultiplierSendResp) error {
	ctx := context.Background()
	for _, usm := range req.UserSupsMultiplierSends {
		// broadcast to specific hub client if session id is provided
		if usm.ToUserSessionID != nil && *usm.ToUserSessionID != "" {
			go c.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", api.HubKeyUserSupsMultiplierSubscribe, usm.ToUserID)), usm.SupsMultipliers, messagebus.BusSendFilterOption{
				SessionID: *usm.ToUserSessionID,
			})
			continue
		}

		// otherwise, broadcast to the target user
		go c.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", api.HubKeyUserSupsMultiplierSubscribe, usm.ToUserID)), usm.SupsMultipliers)
	}
	return nil
}
