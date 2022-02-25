package api

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"passport"
	"passport/db"
	"time"

	"github.com/ninja-software/log_helpers"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-software/tickle"
	"github.com/ninja-syndicate/hub/ext/messagebus"
)

//////////////////
// Asset Repair //
//////////////////

type RepairQueue map[string]bool

func (api *API) InitialAssetRepairCenter() {
	api.fastAssetRepairCenter = make(chan func(RepairQueue))
	fastRepairQueue := make(RepairQueue)
	go func() {
		for fn := range api.fastAssetRepairCenter {
			fn(fastRepairQueue)
		}
	}()
	api.startRepairTicker(RepairTypeFast)

	api.standardAssetRepairCenter = make(chan func(RepairQueue))
	standerRepairQueue := make(RepairQueue)
	go func() {
		for fn := range api.standardAssetRepairCenter {
			fn(standerRepairQueue)
		}
	}()
	api.startRepairTicker(RepairTypeStandard)
}

type RepairType string

const (
	RepairTypeFast     RepairType = "FAST"
	RepairTypeStandard RepairType = "STANDARD"
)

func (api *API) RegisterRepairCenter(rt RepairType, assetHash string) {
	switch rt {
	case RepairTypeFast:
		select {
		case api.fastAssetRepairCenter <- func(rq RepairQueue) {
			rq[assetHash] = true
		}:

		case <-time.After(10 * time.Second):
			api.Log.Err(errors.New("timeout on channel send exceeded"))
			panic("Fast repair")
		}

	case RepairTypeStandard:
		select {
		case api.standardAssetRepairCenter <- func(rq RepairQueue) {
			rq[assetHash] = true
		}:

		case <-time.After(10 * time.Second):
			api.Log.Err(errors.New("timeout on channel send exceeded"))
			panic("standard repair")
		}

	}
}

func (api *API) startRepairTicker(rt RepairType) {
	tickSecond := 0
	TraceTitle := ""
	var repairCenter chan func(RepairQueue)
	switch rt {
	case RepairTypeFast:
		tickSecond = 18 // repair from 0 to 100 take 30 minutes
		TraceTitle = "Fast Repair Center"
		repairCenter = api.fastAssetRepairCenter
	case RepairTypeStandard:
		tickSecond = 864 // repair from 0 to 100 take 24 hours
		TraceTitle = "Standard Repair Center"
		repairCenter = api.standardAssetRepairCenter
	}

	// build tickle
	assetRepairCenter := tickle.New(TraceTitle, float64(tickSecond), func() (int, error) {
		errChan := make(chan error)
		select {
		case repairCenter <- func(rq RepairQueue) {
			if len(rq) == 0 {
				errChan <- nil
				return
			}

			assetHashes := []string{}
			for tokenID := range rq {
				assetHashes = append(assetHashes, tokenID)
			}

			nfts, err := db.XsynAssetDurabilityBulkIncrement(context.Background(), api.Conn, assetHashes)
			if err != nil {
				errChan <- err
				return
			}

			// remove war machine which is completely repaired
			for _, nft := range nfts {
				if nft.Durability == 100 {
					delete(rq, nft.Hash)
				}
			}
			errChan <- nil
		}:

		case <-time.After(10 * time.Second):
			api.Log.Err(errors.New("timeout on channel send exceeded"))
			panic("Asset Repair Center")
		}

		err := <-errChan
		if err != nil {
			return http.StatusInternalServerError, terror.Error(err)
		}
		return http.StatusOK, nil
	})
	assetRepairCenter.Log = log_helpers.NamedLogger(api.Log, TraceTitle)
	assetRepairCenter.DisableLogging = true

	assetRepairCenter.Start()
}

//////////////////////////////
// War Machine Contract Map //
//////////////////////////////

const MaximumContractReward = "10000000000000000000" // 10 sups
const MinimumContractReward = "1000000000000000000"  // 1 sup

type WarMachineContract struct {
	CurrentReward big.Int
}

func (api *API) InitialiseFactionWarMachineContract(factionID passport.FactionID) {
	// get min price
	minPrice := big.NewInt(0)
	minPrice, ok := minPrice.SetString(MinimumContractReward, 10)
	if !ok {
		api.Log.Err(fmt.Errorf("failed to parse 1000000000000000000 to big int"))
		return
	}

	// set initial reward
	warMachineContract := &WarMachineContract{
		CurrentReward: *big.NewInt(0),
	}
	warMachineContract.CurrentReward.Add(&warMachineContract.CurrentReward, minPrice)

	go func() {
		for fn := range api.factionWarMachineContractMap[factionID] {
			fn(warMachineContract)
		}
	}()
}

// recalculateContractReward
func (api *API) recalculateContractReward(ctx context.Context, factionID passport.FactionID, queueNumber int) {
	minPrice := big.NewInt(0)
	minPrice, ok := minPrice.SetString(MinimumContractReward, 10)
	if !ok {
		api.Log.Err(fmt.Errorf("failed to parse 1000000000000000000 to big int"))
		return
	}

	maxPrice := big.NewInt(0)
	maxPrice, ok = maxPrice.SetString(MaximumContractReward, 10)
	if !ok {
		api.Log.Err(fmt.Errorf("failed to parse 10000000000000000000 to big int"))
		return
	}

	if _, ok := api.factionWarMachineContractMap[factionID]; ok {
		select {
		case api.factionWarMachineContractMap[factionID] <- func(wmc *WarMachineContract) {
			// reduce reward price when greater than 10
			if queueNumber >= 10 {
				wmc.CurrentReward.Mul(&wmc.CurrentReward, big.NewInt(99))
				wmc.CurrentReward.Div(&wmc.CurrentReward, big.NewInt(100))

				if wmc.CurrentReward.Cmp(minPrice) < 0 {
					wmc.CurrentReward = *big.NewInt(0)
					wmc.CurrentReward.Add(&wmc.CurrentReward, minPrice)
				}
			} else {
				wmc.CurrentReward.Mul(&wmc.CurrentReward, big.NewInt(101))
				wmc.CurrentReward.Div(&wmc.CurrentReward, big.NewInt(100))

				if wmc.CurrentReward.Cmp(maxPrice) > 0 {
					wmc.CurrentReward = *big.NewInt(0)
					wmc.CurrentReward.Add(&wmc.CurrentReward, maxPrice)
				}
			}

			// broadcast the latest reward
			go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyAssetQueueContractReward, factionID)), wmc.CurrentReward.String())
		}:

		case <-time.After(10 * time.Second):
			api.Log.Err(errors.New("timeout on channel send exceeded"))
			panic("recalculate Contract Reward")
		}
	}
}
