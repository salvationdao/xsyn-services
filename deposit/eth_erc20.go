package deposit

import (
	"context"
	"math/big"
	"passport/deposit/token"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
)

type ETHERC20Listener struct {
	ErrChan chan error
	stop    chan bool
}

func NewETHERC20Listener() *ETHERC20Listener {

	l := &ETHERC20Listener{
		ErrChan: make(chan error),
		stop:    make(chan bool),
	}
	return l

}
func (l *ETHERC20Listener) Chain() string {
	return "mainnet"
}
func (l *ETHERC20Listener) Listen() chan ERC20Transfer {
	output := make(chan ERC20Transfer)
	client, err := ethclient.Dial("wss://mainnet.infura.io/ws/v3/38ee3b4f0d5a4adfb02fe1ca64645e22")
	if err != nil {
		log.Err(err).Msg("dial")
	}
	go func(client *ethclient.Client, output chan ERC20Transfer) {
		// contractAddress := common.HexToAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48") // USDC
		query := ethereum.FilterQuery{
			// Addresses: []common.Address{
			// 	contractAddress,
			// },
		}
		ch := make(chan types.Log)
		sub, err := client.SubscribeFilterLogs(context.Background(), query, ch)
		if err != nil {
			log.Err(err).Msg("create subscription")
		}

		contractAbi, err := abi.JSON(strings.NewReader(string(token.TokenABI)))
		if err != nil {
			log.Err(err).Msg("parse ABI")
		}

		logTransferSig := []byte("Transfer(address,address,uint256)")
		logTransferSigHash := crypto.Keccak256Hash(logTransferSig)
		defer sub.Unsubscribe()
		for {
			select {
			case vLog := <-ch:
				if vLog.Topics[0].Hex() != logTransferSigHash.Hex() {
					log.Debug().Err(err).Msg("not a transfer")
					continue
				}
				var transferEvent LogTransfer
				result, err := contractAbi.Unpack("Transfer", vLog.Data)
				if err != nil {
					log.Debug().Err(err).Msg("unpack ABI")
					continue
				}
				if len(vLog.Topics) != 3 {
					log.Debug().Err(err).Msg("missing parameter")
					continue
				}
				transferEvent.From = common.HexToAddress(vLog.Topics[1].Hex())
				transferEvent.To = common.HexToAddress(vLog.Topics[2].Hex())
				if len(result) > 0 {
					val := result[0].(*big.Int)
					transferEvent.Value = val
				}
				// log.Info().Str("chain", "mainnet").Str("contract", DisplayAddress(vLog.Address.Hex())).Uint64("block", vLog.BlockNumber).Str("tx", DisplayAddress(vLog.TxHash.Hex())).Str("from", DisplayAddress(transferEvent.From.Hex())).Str("to", DisplayAddress(transferEvent.To.Hex())).Str("value", transferEvent.Value.String()).Msg("transfer")
				output <- ERC20Transfer{
					Chain:         l.Chain(),
					Confirmations: 0,
					Contract:      vLog.Address.Hex(),
					TXID:          vLog.TxHash.Hex(),
					From:          transferEvent.From.Hex(),
					To:            transferEvent.To.Hex(),
					Value:         transferEvent.Value,
				}
			case err := <-sub.Err():
				log.Err(err).Msg("subscription")
			}
		}
	}(client, output)
	return output

}
func (l *ETHERC20Listener) Stop() {

}
func (l *ETHERC20Listener) Error() chan error {
	return nil
}
