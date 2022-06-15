package payments

// PriceResp is a record for price feed
// GET /eth_price
// GET /bnb_price
type PriceResp struct {
	Time int    `json:"time"`
	Usd  string `json:"usd"`
}

// PurchaseRecord is a record for a purchase with stablecoins or native tokens for SUPS
// GET /eth_txs
// GET /bnb_txs
// GET /usdc_txs
// GET /busd_txs
type PurchaseRecord struct {
	Chain           int    `json:"chain"`
	BlockNumber     int    `json:"block_number"`
	Confirmations   int    `json:"confirmations"`
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	ContractAddress string `json:"contract_address"`
	ValueInt        string `json:"value_int"`
	ValueDecimals   int    `json:"value_decimals"`
	Symbol          string `json:"symbol"`
	UsdRate         string `json:"usd_rate"`
	Sups            string `json:"sups"`
	TxHash          string `json:"tx_hash"`
}

// NFTOwnerRecord is a record for current owners for NFTs
// GET /nft_tokens
type NFTOwnerRecord struct {
	TxHash          string `json:"tx_hash"`
	LogIndex        int    `json:"log_index"`
	Time            int    `json:"time"`
	Chain           int    `json:"chain"`
	BlockNumber     int    `json:"block_number"`
	Confirmations   int    `json:"confirmations"`
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	ContractAddress string `json:"contract_address"`
	TokenID         int    `json:"token_id"`
}

// SUPTransferRecord used to grab records of SUPS going in and out of the system
// GET /sups_deposit_txs
// GET /sups_withdraw_txs
type SUPTransferRecord struct {
	TxHash          string `json:"tx_hash"`
	LogIndex        int    `json:"log_index"`
	Time            int    `json:"time"`
	Chain           int    `json:"chain"`
	BlockNumber     int    `json:"block_number"`
	Confirmations   int    `json:"confirmations"`
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	ContractAddress string `json:"contract_address"`
	ValueInt        string `json:"value_int"`
	ValueDecimals   int    `json:"value_decimals"`
}

type NFT1155TransferRecord struct {
	TxHash          string `json:"tx_hash"`
	LogIndex        int    `json:"log_index"`
	Time            int    `json:"time"`
	Chain           int    `json:"chain"`
	BlockNumber     int    `json:"block_number"`
	Confirmations   int    `json:"confirmations"`
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	ContractAddress string `json:"contract_address"`
	ValueInt        string `json:"value_int"`
	ValueDecimals   int    `json:"value_decimals"`
	TokenID         int    `json:"token_id"`
}
