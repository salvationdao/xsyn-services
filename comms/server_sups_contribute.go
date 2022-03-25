package comms

import (
	"fmt"
	"passport"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
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
		_, _, _, err := c.UserCacheMap.Transact(&passport.NewTransaction{
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

func (s *S) SupremacySpendSupsHandler(req SpendSupsReq, resp *SpendSupsResp) error {
	err := IsSupremacyClient(req.ApiKey)
	if err != nil {
		return terror.Error(err)
	}
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

	_, _, txID, err := s.UserCacheMap.Transact(tx)
	if err != nil {
		return terror.Error(err, "failed to process sups")
	}

	tx.ID = txID

	//// for refund
	//s.Txs.TxMx.Lock()
	//s.Txs.Txes = append(s.Txs.Txes, &passport.NewTransaction{
	//	ID:                   txID,
	//	From:                 tx.To,
	//	To:                   tx.From,
	//	Amount:               tx.Amount,
	//	TransactionReference: passport.TransactionReference(fmt.Sprintf("refund|sups vote|%s", txID)),
	//})
	//s.Txs.TxMx.Unlock()

	resp.TXID = txID
	return nil
}
