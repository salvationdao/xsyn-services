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

	ExchangeRates *ExchangeRates
}

type ExchangeRates struct {
	USDtoETH decimal.Decimal
	USDtoBNB decimal.Decimal
	SUPtoUSD decimal.Decimal
}

// USDCToSUPS decimal.Decimal // ETH Chain
// SUPToUSD   decimal.Decimal // Use BUSD against SUPS (inverse)
// BUSDToSUPS decimal.Decimal // BSC Chain

// WETHToSUPS decimal.Decimal // ETH Chain (and BSC)

// WBNBToSUPS decimal.Decimal // BSC Chain
// Grab eth price in busd
// Grab bnb price in busd
// Grab sups price in busd
