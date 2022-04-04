package types

import (
	"github.com/ethereum/go-ethereum/common"
)

type Config struct {
	CookieSecure            bool
	EncryptTokens           bool
	EncryptTokensKey        string
	TokenExpirationDays     int
	PassportWebHostURL      string
	GameserverHostURL       string
	MetaMaskSignMessage     string // The message to see in the metamask signup flow, needs to match frontend
	BridgeParams            *BridgeParams
	OnlyWalletConnect       bool
	WhitelistEndpoint       string
	InsecureSkipVerifyCheck bool
	AuthParams              *AuthParams
	WebhookParams           *WebhookParams
}

type BridgeParams struct {
	OperatorAddr     common.Address
	MoralisKey       string
	UsdcAddr         common.Address
	BusdAddr         common.Address
	SupAddr          common.Address
	DepositAddr      common.Address
	PurchaseAddr     common.Address
	WithdrawAddr     common.Address
	SignerPrivateKey string
	BscNodeAddr      string
	EthNodeAddr      string
	BSCChainID       int64
	ETHChainID       int64
	BSCRouterAddr    common.Address
}

type AuthParams struct {
	GameserverToken     string
	GoogleClientID      string
	TwitchClientID      string
	TwitchClientSecret  string
	TwitterAPIKey       string
	TwitterAPISecret    string
	DiscordClientID     string
	DiscordClientSecret string
}

type WebhookParams struct {
	GameserverHostUrl      string
	GameserverWebhookToken string
}