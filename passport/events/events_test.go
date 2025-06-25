package events_test

import (
	"math/big"
	"os"
	"testing"
	"xsyn-services/passport/events"
	"xsyn-services/passport/payments"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lmittmann/w3"
)

func TestIterETH(t *testing.T) {
	RPC_URL := os.Getenv("SUPREMACY_ETH_RPCURL")
	t.Logf("RPC URL: %s", RPC_URL)
	start := payments.SUPSETHCreationBlock

	chunk := 50
	SUPS := payments.Mainnet.SUPSToken
	eventSignature := "Transfer(address indexed from, address indexed to, uint256 value)"
	iter, err := events.NewIterator(
		eventSignature,
		RPC_URL,
		uint64(start),
		nil,
		uint64(chunk),
		true,
		SUPS,
		true,
	)
	if err != nil {
		t.Fatalf("error creating iterator: %v", err)
	}

	e, err := w3.NewEvent(eventSignature)
	if err != nil {
		t.Fatalf("error creating event: %v", err)
	}

	logCounter := 0
	for log, err := range iter {
		if err != nil {
			t.Fatalf("error iterating logs: %v", err)
		}
		if log == nil {
			t.Fatal("received nil log")
		}
		// t.Logf("Log: %v", log)
		var from common.Address
		var to common.Address
		var amount *big.Int

		err = e.DecodeArgs(log, &from, &to, &amount)
		if err != nil {
			t.Fatalf("error decoding log args: %v", err)
		}
		logCounter++

		// t.Logf("Log counter: %d: [%s] %s -> %s", logCounter, amount.String(), from.Hex(), to.Hex())
	}
	if logCounter == 0 {
		t.Fatal("no logs were processed")
	}
}

func TestIterBSC(t *testing.T) {
	RPC_URL := os.Getenv("SUPREMACY_BSC_RPCURL")
	t.Logf("RPC URL: %s", RPC_URL)
	start := payments.SUPSBSCCreationBlock

	chunk := 50
	SUPS := payments.BSC.SUPSToken
	eventSignature := "Transfer(address indexed from, address indexed to, uint256 value)"
	iter, err := events.NewIterator(
		eventSignature,
		RPC_URL,
		uint64(start),
		nil,
		uint64(chunk),
		true,
		SUPS,
		true,
	)
	if err != nil {
		t.Fatalf("error creating iterator: %v", err)
	}

	e, err := w3.NewEvent(eventSignature)
	if err != nil {
		t.Fatalf("error creating event: %v", err)
	}

	logCounter := 0
	for log, err := range iter {
		if err != nil {
			t.Fatalf("error iterating logs: %v", err)
		}
		if log == nil {
			t.Fatal("received nil log")
		}
		// t.Logf("Log: %v", log)
		var from common.Address
		var to common.Address
		var amount *big.Int

		err = e.DecodeArgs(log, &from, &to, &amount)
		if err != nil {
			t.Fatalf("error decoding log args: %v", err)
		}
		logCounter++

		// t.Logf("Log counter: %d: [%s] %s -> %s", logCounter, amount.String(), from.Hex(), to.Hex())
	}
	if logCounter == 0 {
		t.Fatal("no logs were processed")
	}
}
