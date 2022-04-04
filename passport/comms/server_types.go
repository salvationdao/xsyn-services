package comms

import (
	types2 "xsyn-services/types"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type RefundTransactionReq struct {
	ApiKey        string
	TransactionID string `json:"transaction_id"`
}

type RefundTransactionResp struct {
	TransactionID string `json:"transaction_id"`
}

type SpendSupsReq struct {
	ApiKey               string
	Amount               string                      `json:"amount"`
	FromUserID           uuid.UUID                   `json:"from_user_id"`
	ToUserID             uuid.UUID                   `json:"to_user_id"`
	TransactionReference types2.TransactionReference `json:"transaction_reference"`
	Group                types2.TransactionGroup     `json:"group,omitempty"`
	SubGroup             string                      `json:"sub_group"`   //TODO: send battle id
	Description          string                      `json:"description"` //TODO: send descritpion

	NotSafe bool `json:"not_safe"`
}

type SpendSupsResp struct {
	TransactionID string `json:"transaction_id"`
}

type GetMechOwnerResp struct {
	Payload types.JSON
}
type GetAllMechsReq struct {
	ApiKey string
}

type GetAll struct {
	AssetPayload    types.JSON
	MetadataPayload types.JSON
	StorePayload    types.JSON
	UserPayload     types.JSON
	FactionPayload  types.JSON
}

type UserGetReq struct {
	ApiKey string
	UserID types2.UserID `json:"userID"`
}

type UserGetResp struct {
	User *types2.User `json:"user"`
}

type UserBalanceGetReq struct {
	ApiKey string
	UserID uuid.UUID `json:"userID"`
}

type UserBalanceGetResp struct {
	Balance decimal.Decimal `json:"balance"`
}
