package types

import (
	"github.com/ethereum/go-ethereum/common"
)

type Config struct {
	CookieSecure            bool
	CookieKey               string
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
	BotSecret               string
	EmailTemplatePath       string

	CaptchaSiteKey string
	CaptchaSecret  string
}

type BridgeParams struct {
	OperatorAddr          common.Address
	MoralisKey            string
	UsdcAddr              common.Address
	BusdAddr              common.Address
	SupAddrBSC            common.Address
	SupAddrETH            common.Address
	DepositAddr           common.Address
	SignerPrivateKey      string
	BscNodeAddr           string
	EthNodeAddr           string
	BSCChainID            int64
	ETHChainID            int64
	BSCRouterAddr         common.Address
	AchievementsSignerKey string
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
