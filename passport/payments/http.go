package payments

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passlog"

	"github.com/ethereum/go-ethereum/common"
)

const baseURL = "http://v3.supremacy-api.avantdata.com:3001"
const stagingURL = "http://v3-staging.supremacy-api.avantdata.com:3001"

type Path string

const SUPSWithdrawTxs Path = "sups_withdraw_txs"
const SUPSDepositTxs Path = "sups_deposit_txs"
const NFTOwnerPath Path = "nft_tokens"
const BNBPurchasePath Path = "bnb_txs"
const BUSDPurchasePath Path = "busd_txs"
const ETHPurchasePath Path = "eth_txs"
const USDCPurchasePath Path = "usdc_txs"
const MultiTokenTxs Path = "multi_token_txs"

func Ping() error {
	u := fmt.Sprintf("%s/ping", baseURL)
	resp, err := http.Get(u)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("non 200 response: " + strconv.Itoa(resp.StatusCode))
	}
	return nil
}

func getPurchaseRecords(path Path, latestBlock int, testnet bool) ([]*PurchaseRecord, int, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/%s", baseURL, path), nil)
	if err != nil {
		return nil, 0, err
	}
	q := req.URL.Query()
	q.Add("since_block", strconv.Itoa(latestBlock))
	if testnet {
		q.Add("is_testnet", "true")
	}
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("non 200 response for %s: %d", req.URL.String(), resp.StatusCode)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	result := []*PurchaseRecord{}
	err = json.Unmarshal(b, &result)
	if err != nil {
		return nil, 0, err
	}

	latest := latestPurchaseBlockFromRecords(latestBlock, result)

	return result, latest, nil
}

func getNFTOwnerRecords(path Path, collection *boiler.Collection, testnet bool) (map[int]*NFTOwnerStatus, error) {
	l := passlog.L.With().Str("svc", "avant_nft_ownership_update").Logger()
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/%s?contract_address=%s&is_testnet=%v&confirmations=3", baseURL, path, collection.MintContract.String, testnet), nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()

	l.Debug().Str("url", req.URL.String()).Msg("fetch NFT owners from Avant API")
	q.Add("contract_address", collection.MintContract.String)
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non 200 response for %s: %d", req.URL.String(), resp.StatusCode)
	}
	var records []*NFTOwnerRecord
	err = json.NewDecoder(resp.Body).Decode(&records)
	if err != nil {
		return nil, err
	}

	result := OwnerRecordToOwnerStatus(records, collection)

	return result, nil
}

func OwnerRecordToOwnerStatus(records []*NFTOwnerRecord, collection *boiler.Collection) map[int]*NFTOwnerStatus {
	result := map[int]*NFTOwnerStatus{}
	for _, record := range records {
		// Current owner owns it; or
		owner := common.HexToAddress(record.ToAddress)
		if owner.Hex() == common.HexToAddress(collection.StakeContract.String).Hex() || owner.Hex() == common.HexToAddress(collection.StakingContractOld.String).Hex() {
			// Address who sent it to the staking contract owns it
			owner = common.HexToAddress(record.FromAddress)
		}

		onChainStatus := db.STAKABLE
		// Current owner IS staking contract
		if common.HexToAddress(record.ToAddress).Hex() == common.HexToAddress(collection.StakeContract.String).Hex() {
			onChainStatus = db.UNSTAKABLE
		}
		// Current owner IS staking contract
		if common.HexToAddress(record.ToAddress).Hex() == common.HexToAddress(collection.StakingContractOld.String).Hex() {
			onChainStatus = db.UNSTAKABLEOLD
		}

		result[record.TokenID] = &NFTOwnerStatus{
			Collection:    common.HexToAddress(collection.MintContract.String),
			Owner:         owner,
			OnChainStatus: onChainStatus,
		}
	}
	return result
}

func getSUPTransferRecords(path Path, latestBlock int, testnet bool) ([]*SUPTransferRecord, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/%s", baseURL, path), nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("since_block", strconv.Itoa(latestBlock))
	if testnet {
		q.Add("is_testnet", "true")
	}
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non 200 response for %s: %d", req.URL.String(), resp.StatusCode)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	result := []*SUPTransferRecord{}
	err = json.Unmarshal(b, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func getNFT1155TransferRecords(path Path, latestBlock int, testnet bool, contractAddress string) ([]*NFT1155TransferRecord, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/%s", baseURL, path), nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("since_block", strconv.Itoa(latestBlock))
	if testnet {
		q.Add("is_testnet", "true")
	}

	q.Add("contract_address", contractAddress)
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non 200 response for %s: %d", req.URL.String(), resp.StatusCode)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	result := []*NFT1155TransferRecord{}
	err = json.Unmarshal(b, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func GetWithdraws(testnet bool) ([]*SUPTransferRecord, error) {
	latestWithdrawBlock := db.GetInt(db.KeyLatestWithdrawBlock)
	records, err := getSUPTransferRecords(SUPSWithdrawTxs, latestWithdrawBlock, testnet)
	if err != nil {
		return nil, fmt.Errorf("get withdraw txes: %w", err)
	}
	newLatestWithdrawBlock := latestSUPTransferBlockFromRecords(latestWithdrawBlock, records)
	db.PutInt(db.KeyLatestWithdrawBlock, newLatestWithdrawBlock)
	return records, nil
}

func GetDeposits(testnet bool) ([]*SUPTransferRecord, error) {
	latestDepositBlock := db.GetInt(db.KeyLatestDepositBlock)
	records, err := getSUPTransferRecords(SUPSDepositTxs, latestDepositBlock, testnet)
	if err != nil {
		return nil, err
	}
	db.PutInt(db.KeyLatestDepositBlock, latestSUPTransferBlockFromRecords(latestDepositBlock, records))
	return records, nil
}

func GetNFTOwnerRecords(testnet bool, collection *boiler.Collection) (map[int]*NFTOwnerStatus, error) {
	return getNFTOwnerRecords(NFTOwnerPath, collection, testnet)
}

func Get1155Deposits(testnet bool, contractAddress string) ([]*NFT1155TransferRecord, error) {
	latestDepositBlock := db.GetIntWithDefault(db.KeyLatest1155DepositBlock, 0)
	records, err := getNFT1155TransferRecords(MultiTokenTxs, latestDepositBlock, testnet, contractAddress)
	if err != nil {
		return nil, err
	}
	db.PutInt(db.KeyLatest1155DepositBlock, latestNFT1155TransferBlockFromRecords(latestDepositBlock, records))
	return records, nil
}

func Get1155Withdraws(testnet bool, contractAddress string) ([]*NFT1155TransferRecord, error) {
	latest1155Block := db.GetInt(db.KeyLatest1155WithdrawBlock)
	records, err := getNFT1155TransferRecords(MultiTokenTxs, latest1155Block, testnet, contractAddress)
	if err != nil {
		return nil, fmt.Errorf("get 1155 txes: %w", err)
	}
	newLatestWithdrawBlock := latestNFT1155TransferBlockFromRecords(latest1155Block, records)
	db.PutInt(db.KeyLatest1155WithdrawBlock, newLatestWithdrawBlock)
	return records, nil
}

func latestPurchaseBlockFromRecords(currentBlock int, records []*PurchaseRecord) int {
	latestBlock := currentBlock
	for _, record := range records {
		if record.BlockNumber > latestBlock {
			latestBlock = record.BlockNumber
		}
	}
	passlog.L.Debug().Int("latest_block", latestBlock).Msg("Get latest block for purchase records")
	return latestBlock
}

func latestSUPTransferBlockFromRecords(currentBlock int, records []*SUPTransferRecord) int {
	latestBlock := currentBlock
	for _, record := range records {
		if record.BlockNumber > latestBlock {
			latestBlock = record.BlockNumber
		}
	}
	passlog.L.Debug().Int("latest_block", latestBlock).Msg("Get latest block for sup transfer records")
	return latestBlock
}

func latestNFT1155TransferBlockFromRecords(currentBlock int, records []*NFT1155TransferRecord) int {
	latestBlock := currentBlock
	for _, record := range records {
		if record.BlockNumber > latestBlock {
			latestBlock = record.BlockNumber
		}
	}
	passlog.L.Debug().Int("latest_block", latestBlock).Msg("Get latest block for sup transfer records")
	return latestBlock
}
