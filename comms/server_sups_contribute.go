package comms

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"passport"
	"passport/api"
	"passport/db"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/volatiletech/null/v8"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/shopspring/decimal"
)

type InsertTransactionsResp struct {
}
type InsertTransactionsReq struct {
	Transactions []*PendingTransaction
}

// PendingTransaction is an object representing the database table.
type PendingTransaction struct {
	ID                   string
	FromUserID           string
	ToUserID             string
	Amount               decimal.Decimal
	TransactionReference string
	Description          string
	Group                string
	Subgroup             string
	ProcessedAt          null.Time
	DeletedAt            null.Time
	UpdatedAt            time.Time
	CreatedAt            time.Time
}

func (c *S) InsertTransactions(req InsertTransactionsReq, resp *InsertTransactionsResp) error {
	for _, tx := range req.Transactions {
		_, _, _, err := c.UserCacheMap.Process(&passport.NewTransaction{
			From:                 passport.UserID(uuid.Must(uuid.FromString(tx.FromUserID))),
			To:                   passport.UserID(uuid.Must(uuid.FromString(tx.ToUserID))),
			TransactionReference: passport.TransactionReference(tx.TransactionReference),
			Amount:               tx.Amount,
			Description:          tx.Description,
			Group:                passport.TransactionGroup(tx.Group),
			SubGroup:             tx.Subgroup,
		})
		if err != nil {
			return fmt.Errorf("process tx in user cache map: %w", err)
		}
	}
	return nil
}

func (c *S) SupremacySpendSupsHandler(req SpendSupsReq, resp *SpendSupsResp) error {
	amt, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return err
	}

	if amt.LessThan(decimal.Zero) {
		return terror.Error(terror.ErrInvalidInput, "Sups amount can not be negative")
	}

	tx := &passport.NewTransaction{
		From:                 passport.UserID(req.FromUserID),
		To:                   passport.UserID(req.ToUserID),
		TransactionReference: req.TransactionReference,
		Description:          req.Description,
		Amount:               amt,
		Group:                req.Group,
		SubGroup:             req.SubGroup,
	}

	if req.NotSafe {
		tx.NotSafe = true
	}

	_, _, txID, err := c.UserCacheMap.Process(tx)
	if err != nil {
		return terror.Error(err, "failed to process sups")
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

func (c *S) ReleaseTransactionsHandler(req ReleaseTransactionsReq, resp *ReleaseTransactionsResp) error {
	c.Txs.TxMx.Lock()
	defer c.Txs.TxMx.Unlock()
	for _, txID := range req.TxIDs {
		for _, tx := range c.Txs.Txes {
			if txID != tx.ID {
				continue
			}
			_, _, _, err := c.UserCacheMap.Process(tx)
			if err != nil {
				c.Log.Err(err).Msg("failed to process user sups fund")
				continue
			}
		}
	}

	c.Txs.Txes = []*passport.NewTransaction{}

	return nil
}

func (c *S) supremacyFeed() {
	fund := big.NewInt(0)
	fund, ok := fund.SetString("500000000000000000", 10)
	if !ok {
		c.Log.Err(fmt.Errorf("setting string not ok on fund big int")).Msg("too many strings")
		return
	}

	tx := &passport.NewTransaction{
		From:                 passport.XsynTreasuryUserID,
		To:                   passport.SupremacySupPoolUserID,
		Amount:               decimal.NewFromBigInt(fund, 18),
		TransactionReference: passport.TransactionReference(fmt.Sprintf("treasury|ticker|%s", time.Now())),
	}

	// process user cache map
	_, _, _, err := c.UserCacheMap.Process(tx)
	if err != nil {
		c.Log.Err(err).Msg(err.Error())
		return
	}
}

func (c *S) TickerTickHandler(req TickerTickReq, resp *TickerTickResp) error {
	// make treasury send game server user moneys
	// Turn off the supremacy feed for now
	c.supremacyFeed()

	// sups guard
	// kick users off the list, if they don't have any sups\
	um := make(map[passport.UserID]passport.FactionID)
	userMap := make(map[int][]passport.UserID)
	for multiplier, userIDs := range req.UserMap {
		newList := []passport.UserID{}

		for _, userID := range userIDs {
			_, err := c.UserCacheMap.Get(userID.String())
			if err != nil {
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

	if len(um) == 0 {
		return nil
	}

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
	supPool := decimal.NewFromInt32(0)
	supPool = supPool.Add(supsForTick)
	supPool = supPool.Div(decimal.NewFromInt32(3))

	if supPool.LessThan(decimal.NewFromInt32(1)) {
		return nil
	}

	// distribute Red Mountain sups
	c.distributeFund(supPool.String(), int64(rmTotalPoint), rmTotalMap)

	// distribute Boston sups
	c.distributeFund(supPool.String(), int64(bTotalPoint), bTotalMap)

	// distribute Zaibatsu sups
	c.distributeFund(supPool.String(), int64(zTotalPoint), zTotalMap)

	return nil
}

func (c *S) distributeFund(fundstr string, totalPoints int64, userMap map[int][]passport.UserID) {
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
			cap := big.NewInt(1000000000000000000)
			cap.Mul(cap, big.NewInt(100))
			if usersSups.Cmp(cap) >= 0 {
				usersSups = cap
			}

			tx := &passport.NewTransaction{
				From:                 passport.SupremacySupPoolUserID,
				To:                   user,
				Amount:               decimal.NewFromBigInt(usersSups, 0),
				TransactionReference: passport.TransactionReference(fmt.Sprintf("supremacy|ticker|%s|%s", user, time.Now())),
				Description:          "Ticker earnings.",
				Group:                passport.TransactionGroupSupremacy,
				SubGroup:             "Ticker",
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

func (c *S) SupremacyGetSpoilOfWarHandler(req GetSpoilOfWarReq, resp *GetSpoilOfWarResp) error {
	// get current sup pool user sups
	supsPoolUser, err := c.UserCacheMap.Get(passport.SupremacySupPoolUserID.String())
	if err != nil {
		return terror.Error(err)
	}

	battleUser, err := c.UserCacheMap.Get(passport.SupremacyBattleUserID.String())
	if err != nil {
		return terror.Error(err)
	}

	result := decimal.NewFromInt32(0)
	result = result.Add(supsPoolUser)
	result = result.Add(battleUser)

	resp.Amount = result.String()
	return nil
}

func (c *S) UserSupsMultiplierSendHandler(req UserSupsMultiplierSendReq, resp *UserSupsMultiplierSendResp) error {
	for _, usm := range req.UserSupsMultiplierSends {
		// broadcast to specific hub client if session id is provided
		if usm.ToUserSessionID != nil && *usm.ToUserSessionID != "" {
			go c.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", api.HubKeyUserSupsMultiplierSubscribe, usm.ToUserID)), usm.SupsMultipliers, messagebus.BusSendFilterOption{
				SessionID: *usm.ToUserSessionID,
			})
			continue
		}

		// otherwise, broadcast to the target user
		go c.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", api.HubKeyUserSupsMultiplierSubscribe, usm.ToUserID)), usm.SupsMultipliers)
	}
	return nil
}

func (c *S) TransferBattleFundToSupPoolHandler(req TransferBattleFundToSupPoolReq, resp *TransferBattleFundToSupPoolResp) error {
	ctx := context.Background()
	// recalculate faction mvp user
	err := db.FactionMvpMaterialisedViewRefresh(ctx, c.Conn)
	if err != nil {
		return terror.Error(err, "Failed to refresh faction mvp list")
	}

	// generate new go routine to trickle sups
	c.TickerPoolCache.Lock()
	defer c.TickerPoolCache.Unlock()
	// get current battle user sups
	battleUser, err := c.UserCacheMap.Get(passport.SupremacyBattleUserID.String())
	if err != nil {
		return terror.Error(err, "failed to get battle user balance from db")
	}

	// calc trickling sups for current round
	supsForTrickle := decimal.NewFromInt32(0)
	supsForTrickle = supsForTrickle.Add(battleUser)

	// subtrack the sups that is trickling at the moment
	for _, tricklingSups := range c.TickerPoolCache.TricklingAmountMap {
		supsForTrickle = supsForTrickle.Sub(tricklingSups)
	}

	// transfer 10% of current spoil of war back to treasury
	// supsForTreasury := big.NewInt(0)
	// supsForTreasury.Add(supsForTreasury, supsForTrickle)
	// supsForTreasury.Div(supsForTreasury, big.NewInt(10))
	// if supsForTreasury.Cmp(big.NewInt(0)) <= 0 {
	// 	return nil
	// }
	// tx := &passport.NewTransaction{
	// 	From:                 passport.SupremacyBattleUserID,
	// 	To:                   passport.XsynTreasuryUserID,
	// 	Amount:               *supsForTreasury,
	// 	TransactionReference: passport.TransactionReference(fmt.Sprintf("supremacy|battle_sups_spend_transfer|%s", time.Now())),
	// }
	// _, _, _, err = c.UserCacheMap.Process(tx)
	// if err != nil {
	// 	return terror.Error(err, "Failed to transfer 10% spoil of war to treasury")
	// }
	// // reduce the sups for trickle from sups for treasury
	// supsForTrickle.Sub(supsForTrickle, supsForTreasury)

	// so here we want to trickle the battle pool out over 5 minutes, so we create a ticker that ticks every 5 seconds with a max ticks of 300 / 5
	ticksInFiveMinutes := 300 / 5
	supsPerTick := decimal.NewFromInt(0)
	supsPerTick = supsPerTick.Div(decimal.NewFromInt(int64(ticksInFiveMinutes)))

	// skip, if trickle amount is empty
	if supsPerTick.LessThanOrEqual(decimal.Zero) {
		return nil
	}

	// append the amount set to the list
	key := uuid.Must(uuid.NewV4()).String()
	c.TickerPoolCache.TricklingAmountMap[key] = decimal.NewFromInt(0)
	c.TickerPoolCache.TricklingAmountMap[key] = c.TickerPoolCache.TricklingAmountMap[key].Add(supsForTrickle)

	// start a new go routine for current round
	go c.newSupsTrickle(key, ticksInFiveMinutes, supsPerTick)

	return nil
}

// trickle factory
func (c *S) newSupsTrickle(key string, totalTick int, supsPerTick decimal.Decimal) {
	i := 0
	for {
		i++

		tx := &passport.NewTransaction{
			From:                 passport.SupremacyBattleUserID,
			To:                   passport.SupremacySupPoolUserID,
			Amount:               supsPerTick,
			TransactionReference: passport.TransactionReference(fmt.Sprintf("supremacy|battle_sups_spend_transfer|%s", time.Now())),
		}

		c.TickerPoolCache.Lock()
		// process user cache map
		_, _, _, err := c.UserCacheMap.Process(tx)
		if err != nil {
			c.Log.Err(err).Msg("insufficient fund")
			c.TickerPoolCache.Unlock()
			return
		}
		// if the routine is not finished
		if i < totalTick {
			// update current trickling amount
			c.TickerPoolCache.TricklingAmountMap[key] = c.TickerPoolCache.TricklingAmountMap[key].Sub(supsPerTick)

			time.Sleep(5 * time.Second)
			c.TickerPoolCache.Unlock()
			continue
		}
		c.TickerPoolCache.Unlock()

		// otherwise, delete the trickle amount from the map
		delete(c.TickerPoolCache.TricklingAmountMap, key)
		break
	}
}

func (c *S) TopSupsContributorHandler(req TopSupsContributorReq, resp *TopSupsContributorResp) error {
	ctx := context.Background()

	var err error

	// get top contribute users
	resp.TopSupsContributors, err = db.BattleArenaSupsTopContributors(ctx, c.Conn, req.StartTime, req.EndTime)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return terror.Error(err)
	}

	// get top contribute faction
	resp.TopSupsContributeFactions, err = db.BattleArenaSupsTopContributeFaction(ctx, c.Conn, req.StartTime, req.EndTime)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return terror.Error(err)
	}

	return nil
}
