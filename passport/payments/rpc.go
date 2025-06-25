package payments

import (
	"fmt"
	"math/big"
	"xsyn-services/boiler"
	"xsyn-services/passport/events"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lmittmann/w3"
	"github.com/shopspring/decimal"
)

type Addresses struct {
	SUPSToken          common.Address
	WithdrawalContract common.Address
}

var Mainnet Addresses = Addresses{
	SUPSToken:          common.HexToAddress("0xCF39360b26a7E54f6c456E69640671Fc5e774FA2"),
	WithdrawalContract: common.HexToAddress("0x02EC1F7071152a9c50720B158d7b833744BE6Bf7"),
}
var BSC Addresses = Addresses{
	SUPSToken:          common.HexToAddress("0xc99cFaA8f5D9BD9050182f29b83cc9888C5846C4"),
	WithdrawalContract: common.HexToAddress("0x6476dB7cFfeeBf7Cc47Ed8D4996d1D60608AAf95"),
}

type ChainID int

var BSCChainID ChainID = 56
var MainnetChainID ChainID = 1
var GoerliChainID ChainID = 5

var SUPSETHCreationBlock = 15879854
var SUPSBSCCreationBlock = 15443081

type NFTer interface {
	GetNFT1155TransferRecords(fromBlock int, contractAddress string) ([]*NFT1155TransferRecord, error)
	GetNFTOwnerRecords(collection *boiler.Collection) (map[int]*NFTOwnerStatus, error)
}
type Pricer interface {
	FetchPrice(symbol string, passportExchangeRateEnabled bool) (decimal.Decimal, error)
}
type SUPSer interface {
	GetSUPTransferRecords(fromBlock int) ([]*SUPTransferRecord, error)
}
type Purchaser interface {
	GetPurchaseRecords(fromBlock int) ([]*PurchaseRecord, int, error)
}
type Client struct {
	mainnetRpcURL string
	bscRPCURL     string
	chunkSize     int
}

func NewClient(mainnetRpcURL string, bscRpcURL string, chunksize int) *Client {
	return &Client{
		mainnetRpcURL: mainnetRpcURL,
		bscRPCURL:     bscRpcURL,
		chunkSize:     chunksize, // Default chunk size, can be adjusted as needed
	}
}

func (s *Client) FetchPrice(symbol string, passportExchangeRateEnabled bool) (decimal.Decimal, error) {
	return decimal.Zero, nil
}

func (s *Client) GetNFT1155TransferRecords(path Path, latestBlock int, testnet bool, contractAddress string) ([]*NFT1155TransferRecord, error) {
	return nil, nil
}

var eventSignature = "Transfer(address indexed from, address indexed to, uint256 value)"
var transferEvent = w3.MustNewEvent(eventSignature)

func (s *Client) GetSUPTransferRecords(fromBlock int, chainID ChainID) ([]*SUPTransferRecord, error) {
	if chainID != MainnetChainID && chainID != BSCChainID {
		return nil, fmt.Errorf("unsupported chain ID: %d", chainID)
	}
	contractAddress := Mainnet.SUPSToken
	if chainID == BSCChainID {
		contractAddress = BSC.SUPSToken
	}
	rpcURL := s.mainnetRpcURL
	if chainID == BSCChainID {
		rpcURL = s.bscRPCURL
	}
	iter, err := events.NewIterator(
		eventSignature,
		rpcURL,
		uint64(fromBlock),
		nil, // fetch until the latest block
		uint64(s.chunkSize),
		true,
		contractAddress,
		true,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating iterator: %w", err)
	}

	result := []*SUPTransferRecord{}
	logCounter := 0
	for log, err := range iter {
		if err != nil {
			return nil, fmt.Errorf("error iterating logs: %w", err)
		}
		if log == nil {
			return nil, fmt.Errorf("received nil log")
		}
		var from common.Address
		var to common.Address
		var amount *big.Int

		err = transferEvent.DecodeArgs(log, &from, &to, &amount)
		if err != nil {
			return nil, fmt.Errorf("error decoding log args: %w", err)
		}

		logCounter++
		fmt.Printf("Transfer %05d: [%s] %s -> %s\n", logCounter, amount.String(), from.Hex(), to.Hex())
		record := &SUPTransferRecord{
			TxHash:      log.TxHash.Hex(),
			FromAddress: from.Hex(),
			ToAddress:   to.Hex(),
			ValueInt:    amount.String(),
			BlockNumber: int(log.BlockNumber),
		}
		result = append(result, record)
	}

	return result, nil
}

func (s *Client) GetNFTOwnerRecords(path Path, collection *boiler.Collection, testnet bool) (map[int]*NFTOwnerStatus, error) {
	return nil, nil
}

func (s *Client) GetPurchaseRecords(path Path, latestBlock int, testnet bool) ([]*PurchaseRecord, int, error) {
	return nil, 0, nil
}

func (s *Client) Ping() error {
	return nil
}
