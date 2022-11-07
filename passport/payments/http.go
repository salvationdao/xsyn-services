package payments

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passlog"

	"github.com/ethereum/go-ethereum/common"
)

const baseURL = "http://v3.supremacy-api.avantdata.com:3001"
const stagingURL = "http://v3-staging.supremacy-api.avantdata.com:3001"

type Path string

const SUPSWithdrawTxsBSC Path = "sups_withdraw_txs"
const SUPSWithdrawTxsETH Path = "sups_eth_withdraw_txs"
const SUPSDepositTxsBSC Path = "sups_deposit_txs"
const SUPSDepositTxsETH Path = "sups_eth_deposit_txs"
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
		db.PutBool(db.KeyEnableSyncNFTOwners, false)
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
			Collection:     common.HexToAddress(collection.MintContract.String),
			Owner:          owner,
			OnChainStatus:  onChainStatus,
			TxHash:         record.TxHash,
			BlockNumber:    record.BlockNumber,
			BlockTimestamp: time.Unix(int64(record.Time), 0),
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

func GetWithdraws(bscWithdrawalsEnabled, ethWithdrawalsEnabled, testnet bool) ([]*SUPTransferRecord, error) {
	records := []*SUPTransferRecord{}

	if bscWithdrawalsEnabled {
		latestWithdrawBlockBSC := db.GetInt(db.KeyLatestWithdrawBlockBSC)

		bscRecords, err := getSUPTransferRecords(SUPSWithdrawTxsBSC, latestWithdrawBlockBSC, testnet)
		if err != nil {
			return nil, fmt.Errorf("get withdraw txes: %w", err)
		}
		passlog.L.Debug().Int("bsc withdrawals", len(bscRecords)).Msg("getting bsc withdrawals")
		records = append(records, bscRecords...)
		db.PutInt(db.KeyLatestWithdrawBlockBSC, latestSUPTransferBlockFromRecords(latestWithdrawBlockBSC, bscRecords))
	}
	if ethWithdrawalsEnabled {
		latestWithdrawBlockETH := db.GetInt(db.KeyLatestWithdrawBlockETH)

		ethRecords, err := getSUPTransferRecords(SUPSWithdrawTxsETH, latestWithdrawBlockETH, testnet)
		if err != nil {
			return nil, fmt.Errorf("get withdraw txes: %w", err)
		}
		passlog.L.Debug().Int("eth withdrawals", len(ethRecords)).Msg("getting eth withdrawals")
		records = append(records, ethRecords...)
		db.PutInt(db.KeyLatestWithdrawBlockETH, latestSUPTransferBlockFromRecords(latestWithdrawBlockETH, ethRecords))
	}

	return records, nil
}

func GetDeposits(testnet bool) ([]*SUPTransferRecord, error) {
	records := []*SUPTransferRecord{}

	if db.GetBool(db.KeyEnableBscDeposits) {
		latestDepositBlockBSC := db.GetInt(db.KeyLatestDepositBlockBSC)
		bscRecords, err := getSUPTransferRecords(SUPSDepositTxsBSC, latestDepositBlockBSC, testnet)
		if err != nil {
			return nil, err
		}
		if len(bscRecords) > 0 {
			passlog.L.Debug().Int("bsc deposits", len(bscRecords)).Msg("getting bsc deposits")
		}
		records = append(records, bscRecords...)
		db.PutInt(db.KeyLatestDepositBlockBSC, latestSUPTransferBlockFromRecords(latestDepositBlockBSC, bscRecords))
	}

	if db.GetBool(db.KeyEnableEthDeposits) {
		latestDepositBlockETH := db.GetInt(db.KeyLatestDepositBlockETH)
		ethRecords, err := getSUPTransferRecords(SUPSDepositTxsETH, latestDepositBlockETH, testnet)
		if err != nil {
			return nil, err
		}
		if len(ethRecords) > 0 {
			passlog.L.Debug().Int("eth deposits", len(ethRecords)).Msg("getting eth deposits")
		}
		records = append(records, ethRecords...)
		db.PutInt(db.KeyLatestDepositBlockETH, latestSUPTransferBlockFromRecords(latestDepositBlockETH, ethRecords))
	}

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
