package passport

import (
	"math/big"
	"time"
)

type TransactionStatus string

const (
	TransactionSuccess TransactionStatus = "success"
	TransactionFailed  TransactionStatus = "failed"
)

type Transaction struct {
	ID                   int64             `json:"id" db:"id"`
	Credit               UserID            `json:"credit" db:"credit"`
	Debit                UserID            `json:"debit" db:"debit"`
	Amount               BigInt            `json:"amount" db:"amount"`
	Status               TransactionStatus `json:"status" db:"status"`
	TransactionReference string            `json:"transactionReference" db:"transaction_reference"`
	Description          string            `json:"description" db:"description"`
	Reason               string            `json:"reason" db:"reason"`
	CreatedAt            time.Time         `json:"created_at" db:"created_at"`
}

type ChainConfirmations struct {
	Tx                 string     `json:"tx" db:"tx"`
	TxID               int64      `json:"txID" db:"tx_id"`
	Block              uint64     `json:"block" db:"block"`
	ChainID            uint64     `json:"chainID" db:"chain_id"`
	ConfirmedAt        *time.Time `json:"confirmedAt" db:"confirmed_at"`
	CreatedAt          time.Time  `json:"createdAt" db:"created_at"`
	ConfirmationAmount int        `json:"confirmationAmount" db:"confirmation_amount"`
	UserID             UserID     `json:"userID" db:"user_id"`
}

type TransactionReference string

type NewTransaction struct {
	To                   UserID               `json:"credit" db:"credit"`
	From                 UserID               `json:"debit" db:"debit"`
	Amount               big.Int              `json:"amount" db:"amount"`
	TransactionReference TransactionReference `json:"transactionReference" db:"transaction_reference"`
	Description          string               `json:"description" db:"description"`
	ResultChan           chan *TransactionResult
}

type TransactionResult struct {
	Transaction *Transaction
	Error       error
}
