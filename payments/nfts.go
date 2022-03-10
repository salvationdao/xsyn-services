package payments

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"passport/db"
	"passport/db/boiler"
	"passport/passdb"
	"passport/passlog"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gofrs/uuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type NFTOwner struct {
	TxHash string `json:"tx_hash"`
	Time   int    `json:"time"`
	JSON   struct {
		To                string `json:"to"`
		Gas               string `json:"gas"`
		From              string `json:"from"`
		Hash              string `json:"hash"`
		Input             string `json:"input"`
		Nonce             string `json:"nonce"`
		GasUsed           string `json:"gasUsed"`
		TokenID           string `json:"tokenID"`
		GasPrice          string `json:"gasPrice"`
		BlockHash         string `json:"blockHash"`
		TimeStamp         string `json:"timeStamp"`
		TokenName         string `json:"tokenName"`
		BlockNumber       string `json:"blockNumber"`
		TokenSymbol       string `json:"tokenSymbol"`
		TokenDecimal      string `json:"tokenDecimal"`
		Confirmations     string `json:"confirmations"`
		ContractAddress   string `json:"contractAddress"`
		TransactionIndex  string `json:"transactionIndex"`
		CumulativeGasUsed string `json:"cumulativeGasUsed"`
	} `json:"json"`
	Chain           int    `json:"chain"`
	BlockNumber     int    `json:"block_number"`
	Confirmations   int    `json:"confirmations"`
	ContractAddress string `json:"contract_address"`
	TokenID         int    `json:"token_id"`
	OwnerAddress    string `json:"owner_address"`
}

var MainnetNFT = common.HexToAddress("0x651D4424F34e6e918D8e4D2Da4dF3DEbDAe83D0C")
var MainnetStaking = common.HexToAddress("0x6476dB7cFfeeBf7Cc47Ed8D4996d1D60608AAf95")
var TestnetNFT = common.HexToAddress("0xEEfaF47acaa803176F1711c1cE783e790E4E750D")
var TestnetStaking = common.HexToAddress("0x0497e0F8FC07DaaAf2BC1da1eace3F5E60d008b8")

type NFTOwnerStatus struct {
	Collection common.Address
	Owner      common.Address
	Stakable   bool
	Unstakable bool
}

func AllNFTOwners(isTestnet bool) (map[int]*NFTOwnerStatus, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/%s", baseURL, "nft_tokens"), nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()

	NFTAddr := MainnetNFT
	if isTestnet {
		NFTAddr = TestnetNFT
	}
	q.Add("contract_address", NFTAddr.Hex())
	q.Add("is_testnet", fmt.Sprintf("%v", isTestnet))
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non 200 response for %s: %d", req.URL.String(), resp.StatusCode)
	}
	records := []*NFTOwner{}
	err = json.NewDecoder(resp.Body).Decode(&records)
	if err != nil {
		return nil, err
	}

	StakingAddr := MainnetStaking
	if isTestnet {
		StakingAddr = TestnetStaking
	}

	result := map[int]*NFTOwnerStatus{}
	for _, record := range records {
		// Current owner owns it
		owner := common.HexToAddress(record.OwnerAddress)
		if record.OwnerAddress == StakingAddr.Hex() {
			// Address who sent it to the staking contract owns it
			owner = common.HexToAddress(record.JSON.From)
		}

		result[record.TokenID] = &NFTOwnerStatus{
			Collection: NFTAddr,
			Owner:      owner,
			Stakable:   record.OwnerAddress != StakingAddr.Hex(),
			Unstakable: record.OwnerAddress == StakingAddr.Hex(),
		}
	}

	return result, nil
}

func UpdateOwners(nftStatuses map[int]*NFTOwnerStatus, isTestnet bool) error {
	NFTAddr := MainnetNFT
	if isTestnet {
		NFTAddr = TestnetNFT
	}

	updated := 0
	for tokenID, nftStatus := range nftStatuses {
		purchasedItem, err := db.PurchasedItemByMintContractAndTokenID(NFTAddr, tokenID)
		if err != nil && errors.Is(err, sql.ErrNoRows) {
			passlog.L.Debug().Str("collection_addr", NFTAddr.Hex()).Int("external_token_id", tokenID).Msg("item not found")
			continue
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("get purchased item: %w", err)
		}

		if nftStatus.Stakable {
			u, err := CreateOrGetUser(context.Background(), passdb.Conn, nftStatus.Owner)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("get user: %w", err)
			}

			itemID := uuid.Must(uuid.FromString(purchasedItem.ID))
			userID := uuid.UUID(u.ID)

			err = db.PurchasedItemSetOnChainStatus(itemID, db.STAKABLE)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("set new nft status: %w", err)
			}
			_, err = db.PurchasedItemSetOwner(itemID, userID)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("set new nft owner: %w", err)
			}
			updated++
			continue
		}

		if nftStatus.Unstakable {
			u, err := CreateOrGetUser(context.Background(), passdb.Conn, nftStatus.Owner)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("get user: %w", err)
			}

			itemID := uuid.Must(uuid.FromString(purchasedItem.ID))
			userID := uuid.UUID(u.ID)

			err = db.PurchasedItemSetOnChainStatus(itemID, db.UNSTAKABLE)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("set new nft status: %w", err)
			}
			_, err = db.PurchasedItemSetOwner(itemID, userID)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("set new nft owner: %w", err)
			}
			updated++
			continue
		}
	}

	passlog.L.Info().Int("updated", updated).Msg("synced nft ownerships")

	return nil
}

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
		TimeStamp        string      `json:"timeStamp"`
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

		unixTime, err := strconv.Atoi(nfttx.JSON.TimeStamp)
		if err != nil {
			return skipped, success, err
		}

		t := time.Unix(int64(unixTime), 0)

		record := &boiler.ItemOnchainTransaction{
			CollectionID:    collection.ID,
			TXID:            common.HexToHash(nfttx.TxHash).Hex(),
			ExternalTokenID: nfttx.TokenID,
			ContractAddr:    contractAddr.Hex(),
			FromAddr:        common.HexToAddress(nfttx.FromAddress).Hex(),
			ToAddr:          common.HexToAddress(nfttx.ToAddress).Hex(),
			BlockNumber:     nfttx.BlockNumber,
			BlockTimestamp:  t,
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
