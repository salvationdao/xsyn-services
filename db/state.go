package db

import (
	"passport"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
	"golang.org/x/net/context"
)

// StateGet gets the latest state
func StateGet(ctx context.Context, isTestnetBlockchain bool, conn Conn) (*passport.State, error) {
	state := &passport.State{}
	q := `SELECT latest_eth_block, latest_bsc_block, eth_to_usd, bnb_to_usd, sup_to_usd FROM state`
	if isTestnetBlockchain {
		q = `SELECT latest_block_eth_testnet AS latest_eth_block, latest_block_bsc_testnet AS latest_bsc_block, eth_to_usd, bnb_to_usd, sup_to_usd FROM state`
	}
	err := pgxscan.Get(ctx, conn, state, q)
	if err != nil {
		return nil, terror.Error(err)
	}
	return state, nil
}

// UpdateLatestETHBlock updates the latest eth block checked
func UpdateLatestETHBlock(ctx context.Context, isTestnetBlockchain bool, conn Conn, latestBlock uint64) (*passport.State, error) {
	state := &passport.State{}
	q := `UPDATE state SET latest_eth_block = $1 RETURNING latest_eth_block, latest_bsc_block, eth_to_usd, bnb_to_usd, sup_to_usd`
	if isTestnetBlockchain {
		q = `UPDATE state SET latest_block_eth_testnet = $1 RETURNING latest_block_eth_testnet AS latest_eth_block, latest_block_bsc_testnet AS latest_bsc_block, eth_to_usd, bnb_to_usd, sup_to_usd`
	}
	err := pgxscan.Get(ctx, conn, state, q, latestBlock)
	if err != nil {
		return nil, terror.Error(err)
	}
	return state, nil
}

// UpdateLatestBSCBlock updates the latest eth block checked
func UpdateLatestBSCBlock(ctx context.Context, isTestnetBlockchain bool, conn Conn, latestBlock uint64) (*passport.State, error) {
	state := &passport.State{}
	q := `UPDATE state SET latest_bsc_block = $1 RETURNING latest_eth_block, latest_bsc_block, eth_to_usd, bnb_to_usd, sup_to_usd`
	if isTestnetBlockchain {
		q = `UPDATE state SET latest_block_bsc_testnet = $1 RETURNING latest_block_eth_testnet AS latest_eth_block, latest_block_bsc_testnet AS latest_bsc_block, eth_to_usd, bnb_to_usd, sup_to_usd`
	}
	err := pgxscan.Get(ctx, conn, state, q, latestBlock)
	if err != nil {
		return nil, terror.Error(err)
	}
	return state, nil
}

// UpdateExchangeRates updates the latest eth block checked
func UpdateExchangeRates(ctx context.Context, isTestnetBlockchain bool, conn Conn, exchangeRates *passport.State) (*passport.State, error) {
	state := &passport.State{}
	q := `UPDATE state SET eth_to_usd = $1, bnb_to_usd = $2, sup_to_usd = $3 RETURNING latest_eth_block, latest_bsc_block, eth_to_usd, bnb_to_usd, sup_to_usd`
	if isTestnetBlockchain {
		q = `UPDATE state SET eth_to_usd = $1, bnb_to_usd = $2, sup_to_usd = $3 RETURNING latest_block_eth_testnet AS latest_eth_block, latest_block_bsc_testnet AS latest_bsc_block, eth_to_usd, bnb_to_usd, sup_to_usd`
	}
	err := pgxscan.Get(ctx, conn, state, q, exchangeRates.ETHtoUSD, exchangeRates.BNBtoUSD, exchangeRates.SUPtoUSD)
	if err != nil {
		return nil, terror.Error(err)
	}
	return state, nil
}
