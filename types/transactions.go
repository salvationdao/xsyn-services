package types

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/volatiletech/null/v8"

	"github.com/shopspring/decimal"
)

type TransactionStatus string

const (
	TransactionSuccess TransactionStatus = "success"
	TransactionFailed  TransactionStatus = "failed"
)
const SUPSSymbol = "SUPS"
const ETHSymbol = "ETH"
const BNBSymbol = "BNB"
const BUSDSymbol = "BUSD"
const USDCSymbol = "USDC"

const ETHDecimals = 18
const BNBDecimals = 18
const SUPSDecimals = 18

func (c *User) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(b, c)
}

type Transaction struct {
	ID                   string            `json:"id" db:"id"`
	RelatedTransactionID null.String       `json:"related_transaction_id" db:"related_transaction_id"`
	ServiceID            UserID            `json:"service_id" db:"service_id"`
	Credit               UserID            `json:"credit" db:"credit"`
	Debit                UserID            `json:"debit" db:"debit"`
	Amount               decimal.Decimal   `json:"amount" db:"amount"`
	Status               TransactionStatus `json:"status" db:"status"`
	TransactionReference string            `json:"transaction_reference" db:"transaction_reference"`
	Description          string            `json:"description" db:"description"`
	Reason               *string           `json:"reason" db:"reason"`
	CreatedAt            time.Time         `json:"created_at" db:"created_at"`
	Group                TransactionGroup  `json:"group" db:"group"`
	SubGroup             string            `json:"sub_group" db:"sub_group"`

	// Inner joined fields4b4
	To   User `json:"to"`
	From User `json:"from"`
}

type ChainConfirmations struct {
	Tx                 string     `json:"tx" db:"tx"`
	TxID               string     `json:"tx_id" db:"tx_id"`
	Block              uint64     `json:"block" db:"block"`
	ChainID            uint64     `json:"chain_id" db:"chain_id"`
	ConfirmedAt        *time.Time `json:"confirmed_at" db:"confirmed_at"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	ConfirmationAmount int        `json:"confirmation_amount" db:"confirmation_amount"`
	UserID             UserID     `json:"user_id" db:"user_id"`
}

type TransactionReference string

type NewTransaction struct {
	ID                   string               `json:"id" db:"id"`
	RelatedTransactionID null.String          `json:"related_transaction_id" db:"related_transaction_id"`
	ServiceID            UserID               `json:"service_id" db:"service_id"`
	To                   UserID               `json:"credit" db:"credit"`
	From                 UserID               `json:"debit" db:"debit"`
	Amount               decimal.Decimal      `json:"amount" db:"amount"`
	TransactionReference TransactionReference `json:"transaction_reference" db:"transaction_reference"`
	Description          string               `json:"description" db:"description"`
	Group                TransactionGroup     `json:"group" db:"group"`
	SubGroup             string               `json:"sub_group" db:"sub_group"`
	NotSafe              bool                 `json:"not_safe" db:"-"`
	Processed            bool                 `json:"processed" db:"-"`
	CreatedAt            time.Time            `json:"created_at" db:"created_at"`
}

type TransactionResult struct {
	Transaction *Transaction
	Error       error
}

type TransactionGroup string

const (
	TransactionGroupStore           TransactionGroup = "STORE"
	TransactionGroupDeposit         TransactionGroup = "DEPOSIT"
	TransactionGroupWithdrawal      TransactionGroup = "WITHDRAWAL"
	TransactionGroupBattle          TransactionGroup = "BATTLE"
	TransactionGroupSupremacy       TransactionGroup = "SUPREMACY"
	TransactionGroupAssetManagement TransactionGroup = "ASSET MANAGEMENT"
	TransactionGroupTesting         TransactionGroup = "TESTING"
)
