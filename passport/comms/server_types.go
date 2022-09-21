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
}

type SpendSupsResp struct {
	TransactionID string `json:"transaction_id"`
}

type GetMechOwnerResp struct {
	Payload types.JSON
}

type UserGetReq struct {
	ApiKey string
	UserID types2.UserID `json:"userID"`
}

type UserBalanceGetReq struct {
	ApiKey string
	UserID uuid.UUID `json:"userID"`
}

type UserBalanceGetResp struct {
	Balance decimal.Decimal `json:"balance"`
}

type AssetOnChainStatusReq struct {
	AssetID string `json:"asset_id"`
}

type AssetOnChainStatusResp struct {
	OnChainStatus string `json:"on_chain_status"`
}

type AssetsOnChainStatusReq struct {
	AssetIDs []string `json:"asset_ids"`
}

type AssetsOnChainStatusResp struct {
	OnChainStatuses map[string]string `json:"on_chain_statuses"`
}

type UserFactionEnlistReq struct {
	ApiKey    string
	UserID    string `json:"userID"`
	FactionID string `json:"factionID"`
}

type UserFactionEnlistResp struct{}

type SyndicateCreateReq struct {
	ApiKey      string `json:"api_key"`
	SyndicateID string `json:"syndicate_id"`
	FoundedByID string `json:"founded_by_id"`
	Name        string `json:"name"`
}
type SyndicateCreateResp struct{}

type SyndicateNameCreateReq struct {
	ApiKey      string `json:"api_key"`
	SyndicateID string `json:"syndicate_id"`
	Name        string `json:"name"`
}
type SyndicateNameChangeResp struct{}

type SyndicateLiquidateReq struct {
	ApiKey        string   `json:"api_key"`
	SyndicateID   string   `json:"syndicate_id"`
	RemainUserIDs []string `json:"remain_user_ids"`
}
type SyndicateLiquidateResp struct{}

type GetCurrentSupPriceReq struct{}

type GetCurrentSupPriceResp struct {
	PriceUSD decimal.Decimal `json:"price_usd"`
}
