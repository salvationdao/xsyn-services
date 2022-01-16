package passport

import (
	"time"

	"github.com/gofrs/uuid"
)

type TransactionStatus string

const (
	TransactionPending TransactionStatus = "pending"
	TransactionSuccess TransactionStatus = "success"
	TransactionFailed  TransactionStatus = "failed"
)

type Transaction struct {
	ID                   uuid.UUID         `json:"id" db:"id"`
	FromID               UserID            `json:"fromId" db:"from_id"`
	ToID                 UserID            `json:"toId" db:"to_id"`
	Amount               BigInt            `json:"amount" db:"amount"`
	Status               TransactionStatus `json:"status" db:"status"`
	TransactionReference string            `json:"transactionReference" db:"transaction_reference"`
	CreatedAt            time.Time         `json:"created_at" db:"created_at"`
}
