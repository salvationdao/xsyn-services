package events

import (
	"fmt"
	"iter"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type EventIterator interface {
	CurrentBlock() uint64
}

type IteratorFactory struct {
	Client          *w3.Client
	StartBlock      uint64
	FinalBlock      *big.Int
	ChunkSize       uint64
	WithProgress    bool
	ContractAddress common.Address
	Event           *w3.Event
	Debug           bool
}

func NewIterator(
	eventSignature string,
	RPCURL string,
	startBlock uint64,
	endBlock *big.Int,
	chunkSize uint64,
	withProgress bool,
	contractAddress common.Address,
	debug bool,
) (iter.Seq2[*types.Log, error], error) {
	client, err := w3.Dial(RPCURL)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}
	if endBlock == nil {
		var finalblock *big.Int
		err = client.Call(
			eth.BlockNumber().Returns(&finalblock),
		)
		if err != nil {
			return nil, fmt.Errorf("call: %w", err)
		}
		endBlock = finalblock
	}

	e, err := w3.NewEvent(eventSignature)
	if err != nil {
		return nil, fmt.Errorf("new event: %w", err)
	}

	i := &IteratorFactory{
		Client:          client,
		StartBlock:      startBlock,
		FinalBlock:      endBlock,
		ChunkSize:       chunkSize,
		WithProgress:    withProgress,
		ContractAddress: contractAddress,
		Event:           e,
		Debug:           debug,
	}
	fn, err := newIter(i)
	if err != nil {
		return nil, fmt.Errorf("new iterator: %w", err)
	}
	return fn, nil
}

// newIter returns a function that iterates over the claimed events
func newIter(
	opts *IteratorFactory,
) (iter.Seq2[*types.Log, error], error) {

	var err error
	if opts.FinalBlock == nil {
		var finalblock *big.Int
		err = opts.Client.Call(
			eth.BlockNumber().Returns(&finalblock),
		)
		if err != nil {
			return nil, fmt.Errorf("call: %w", err)
		}
	}

	startBlock := opts.StartBlock
	fn := func(yield func(*types.Log, error) bool) {
		var logs []types.Log
		t := time.NewTicker(5 * time.Second)
		defer t.Stop()
		totalLogs := 0
		for {
			toBlock := startBlock + uint64(opts.ChunkSize) - 1
			toBlock = min(toBlock, opts.FinalBlock.Uint64())

			if opts.WithProgress {
				fmt.Printf("Processing block %d-%d/%d\n", startBlock, toBlock, opts.FinalBlock.Uint64())
			}
			f := ethereum.FilterQuery{
				BlockHash: nil,
				FromBlock: big.NewInt(int64(startBlock)),
				ToBlock:   big.NewInt(int64(toBlock)),
				Addresses: []common.Address{opts.ContractAddress},
				Topics:    [][]common.Hash{{opts.Event.Topic0}},
			}
			err = opts.Client.Call(
				eth.Logs(f).Returns(&logs),
			)
			if err != nil {
				yield(nil, fmt.Errorf("call: %w", err))
				return
			}
			if opts.Debug {
				fmt.Printf("Found %d logs in block range %d-%d\n", len(logs), startBlock, toBlock)
			}
			blockLogCount := orderedmap.New[uint64, int]()
			for _, log := range logs {
				value, _ := blockLogCount.Get(log.BlockNumber)
				blockLogCount.Set(log.BlockNumber, value+1)
				if !yield(&log, nil) {
					return
				}
				totalLogs++
			}
			// Print the number of logs per block in order
			if opts.Debug {
				for pair := blockLogCount.Oldest(); pair != nil; pair = pair.Next() {
					fmt.Printf("[Block %v] => %v logs\n", pair.Key, pair.Value)
				}
			}
			if toBlock == opts.FinalBlock.Uint64() {
				break
			}
			startBlock = toBlock + 1
		}
	}
	return fn, nil
}
