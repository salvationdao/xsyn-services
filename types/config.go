package types

import "github.com/ethereum/go-ethereum/common"

type Config struct {
	CookieSecure            bool
	CookieKey               string
	EncryptTokens           bool
	EncryptTokensKey        string
	TokenExpirationDays     int
	PassportWebHostURL      string
	GameserverHostURL       string
	MetaMaskSignMessage     string // The message to see in the metamask signup flow, needs to match frontend
	Web3Params              *Web3Params
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

type Web3Params struct {
	BscChainID            int
	EthChainID            int
	MoralisKey            string
	SignerPrivateKey      string
	AchievementsSignerKey string
	SupAddrBSC            common.Address
	SupAddrETH            common.Address
	SupWithdrawalAddrBSC  common.Address
	SupWithdrawalAddrETH  common.Address
	PurchaseAddress       common.Address
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
