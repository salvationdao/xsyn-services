package payments

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"passport/db/boiler"
	"passport/passdb"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type NFTTransaction struct {
	TxHash          string `json:"tx_hash"`
	Time            int    `json:"time"`
	TokenID         int    `json:"token_id"`
	ToAddress       string `json:"to_address"`
	FromAddress     string `json:"from_address"`
	ContractAddress string `json:"contract_address"`
	BlockNumber     int    `json:"block_number"`
	IsVerified      bool   `json:"is_verified"`
	JSON            struct {
		Value            string      `json:"value"`
		Amount           string      `json:"amount"`
		Operator         interface{} `json:"operator"`
		TokenID          string      `json:"token_id"`
		Verified         int         `json:"verified"`
		LogIndex         int         `json:"log_index"`
		BlockHash        string      `json:"block_hash"`
		ToAddress        string      `json:"to_address"`
		BlockNumber      string      `json:"block_number"`
		FromAddress      string      `json:"from_address"`
		ContractType     string      `json:"contract_type"`
		TokenAddress     string      `json:"token_address"`
		BlockTimestamp   time.Time   `json:"block_timestamp"`
		TransactionHash  string      `json:"transaction_hash"`
		TransactionType  string      `json:"transaction_type"`
		TransactionIndex int         `json:"transaction_index"`
	} `json:"json"`
}

var latestNFTBlock = 0

func GetNFTTransactions(contractAddr common.Address) ([]*NFTTransaction, error) {
	u, err := url.Parse("http://104.238.152.254:3001/api/nft_txs")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("contract_address", contractAddr.Hex())
	q.Set("since_block", strconv.Itoa(latestNFTBlock))
	u.RawQuery = q.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	result := []*NFTTransaction{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	updateLatestNFTBlockFromRequest(result)
	return result, nil
}

func updateLatestNFTBlockFromRequest(txes []*NFTTransaction) {
	for _, tx := range txes {
		if tx.BlockNumber > latestNFTBlock {
			latestNFTBlock = tx.BlockNumber
		}
	}
}

func UpsertNFTTransactions(contractAddr common.Address, nftTxes []*NFTTransaction) (int, int, error) {
	tx, err := passdb.StdConn.Begin()
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()
	collectionExists, err := boiler.Collections(
		boiler.CollectionWhere.MintContract.EQ(null.StringFrom(contractAddr.Hex())),
	).Exists(tx)
	if err != nil {
		return 0, 0, fmt.Errorf("collection does not exist: %w", err)
	}
	if !collectionExists {
		return 0, 0, fmt.Errorf("collection does not exist: %s", contractAddr)
	}

	collection, err := boiler.Collections(
		boiler.CollectionWhere.MintContract.EQ(null.StringFrom(contractAddr.Hex())),
	).One(tx)
	if err != nil {
		return 0, 0, fmt.Errorf("get collection: %w", err)
	}

	skipped := 0
	success := 0
	for _, nfttx := range nftTxes {
		exists, err := boiler.ItemOnchainTransactions(boiler.ItemOnchainTransactionWhere.TXID.EQ(nfttx.TxHash)).Exists(tx)
		if err != nil {
			return skipped, success, err
		}
		if exists {
			skipped++
			continue
		}
		record := &boiler.ItemOnchainTransaction{
			CollectionID:    collection.ID,
			TXID:            common.HexToHash(nfttx.TxHash).Hex(),
			ExternalTokenID: nfttx.TokenID,
			ContractAddr:    contractAddr.Hex(),
			FromAddr:        common.HexToAddress(nfttx.FromAddress).Hex(),
			ToAddr:          common.HexToAddress(nfttx.ToAddress).Hex(),
			BlockNumber:     nfttx.BlockNumber,
			BlockTimestamp:  nfttx.JSON.BlockTimestamp,
		}
		err = record.Insert(tx, boil.Infer())
		if err != nil {
			return skipped, success, fmt.Errorf("insert onchain tx: %w", err)
		}
		success++
	}

	tx.Commit()
	return skipped, success, nil
}
