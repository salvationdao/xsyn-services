package passport

import (
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
