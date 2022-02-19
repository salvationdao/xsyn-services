package passport

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

type Config struct {
	CookieSecure        bool
	EncryptTokens       bool
	EncryptTokensKey    string
	TokenExpirationDays int
	PassportWebHostURL  string
	GameserverHostURL   string
	MetaMaskSignMessage string // The message to see in the metamask signup flow, needs to match frontend
	BridgeParams        *BridgeParams
}

type BridgeParams struct {
	UsdcAddr          common.Address
	BusdAddr          common.Address
	WethAddr          common.Address
	BscTestWethAddr   common.Address
	WbnbAddr          common.Address
	SupAddr           common.Address
	DepositAddr       common.Address
	PurchaseAddr      common.Address
	WithdrawAddr      common.Address
	RedemptionAddr    common.Address
	SignerAddr        string
	EthNftAddr        common.Address
	EthNftStakingAddr common.Address
	BscNodeAddr       string
	EthNodeAddr       string
	BSCChainID        int64
	ETHChainID        int64
	RouterAddr        common.Address
	ExchangeRates     *ExchangeRates
}

type ExchangeRates struct {
	USDtoETH decimal.Decimal
	USDtoBNB decimal.Decimal
	SUPtoUSD decimal.Decimal
}
