package passport

import (
	"github.com/ethereum/go-ethereum/common"
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
	OnlyWalletConnect   bool
}

type BridgeParams struct {
	MoralisKey        string
	UsdcAddr          common.Address
	BusdAddr          common.Address
	SupAddr           common.Address
	DepositAddr       common.Address
	PurchaseAddr      common.Address
	WithdrawAddr      common.Address
	SignerAddr        string
	EthNftAddr        common.Address
	EthNftStakingAddr common.Address
	BscNodeAddr       string
	EthNodeAddr       string
	BSCChainID        int64
	ETHChainID        int64
	BSCRouterAddr     common.Address
}
