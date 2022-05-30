package db

import (
	"xsyn-services/boiler"
	"xsyn-services/passport/passdb"
	"xsyn-services/types"

	"github.com/shopspring/decimal"
)

func SupInCents() (decimal.Decimal, error) {
	state, err := boiler.States().One(passdb.StdConn)
	return state.SupToUsd.Mul(decimal.NewFromInt(100)), err
}

// StateGet gets the latest state
func StateGet(isTestnetBlockchain bool) (*types.State, error) {
	state := &types.State{}
	q := `SELECT latest_eth_block, latest_bsc_block, eth_to_usd, bnb_to_usd, sup_to_usd FROM state`
	if isTestnetBlockchain {
		q = `SELECT latest_block_eth_testnet AS latest_eth_block, latest_block_bsc_testnet AS latest_bsc_block, eth_to_usd, bnb_to_usd, sup_to_usd FROM state`
	}

	err := passdb.StdConn.QueryRow(q).Scan(&state.LatestEthBlock, &state.LatestBscBlock, &state.ETHtoUSD, &state.BNBtoUSD, &state.SUPtoUSD)
	if err != nil {
		return nil, err
	}
	return state, nil
}

// UpdateLatestETHBlock updates the latest eth block checked
func UpdateLatestETHBlock(isTestnetBlockchain bool, latestBlock uint64) (*types.State, error) {
	state := &types.State{}
	q := `UPDATE state SET latest_eth_block = $1 RETURNING latest_eth_block, latest_bsc_block, eth_to_usd, bnb_to_usd, sup_to_usd`
	if isTestnetBlockchain {
		q = `UPDATE state SET latest_block_eth_testnet = $1 RETURNING latest_block_eth_testnet AS latest_eth_block, latest_block_bsc_testnet AS latest_bsc_block, eth_to_usd, bnb_to_usd, sup_to_usd`
	}

	err := passdb.StdConn.QueryRow(q, latestBlock).Scan(&state.LatestEthBlock, &state.LatestBscBlock, &state.ETHtoUSD, &state.BNBtoUSD, &state.SUPtoUSD)
	if err != nil {
		return nil, err
	}
	return state, nil
}

// UpdateLatestBSCBlock updates the latest eth block checked
func UpdateLatestBSCBlock(isTestnetBlockchain bool, latestBlock uint64) (*types.State, error) {
	state := &types.State{}
	q := `UPDATE state SET latest_bsc_block = $1 RETURNING latest_eth_block, latest_bsc_block, eth_to_usd, bnb_to_usd, sup_to_usd`
	if isTestnetBlockchain {
		q = `UPDATE state SET latest_block_bsc_testnet = $1 RETURNING latest_block_eth_testnet AS latest_eth_block, latest_block_bsc_testnet AS latest_bsc_block, eth_to_usd, bnb_to_usd, sup_to_usd`
	}
	err := passdb.StdConn.QueryRow(q, latestBlock).Scan(&state.LatestEthBlock, &state.LatestBscBlock, &state.ETHtoUSD, &state.BNBtoUSD, &state.SUPtoUSD)
	if err != nil {
		return nil, err
	}
	return state, nil
}

// UpdateExchangeRates updates the latest eth block checked
func UpdateExchangeRates(isTestnetBlockchain bool, exchangeRates *types.State) (*types.State, error) {
	state := &types.State{}
	q := `UPDATE state SET eth_to_usd = $1, bnb_to_usd = $2, sup_to_usd = $3 RETURNING latest_eth_block, latest_bsc_block, eth_to_usd, bnb_to_usd, sup_to_usd`
	if isTestnetBlockchain {
		q = `UPDATE state SET eth_to_usd = $1, bnb_to_usd = $2, sup_to_usd = $3 RETURNING latest_block_eth_testnet AS latest_eth_block, latest_block_bsc_testnet AS latest_bsc_block, eth_to_usd, bnb_to_usd, sup_to_usd`
	}
	err := passdb.StdConn.QueryRow(q, exchangeRates.ETHtoUSD, exchangeRates.BNBtoUSD, exchangeRates.SUPtoUSD).Scan(&state.LatestEthBlock, &state.LatestBscBlock, &state.ETHtoUSD, &state.BNBtoUSD, &state.SUPtoUSD)
	if err != nil {
		return nil, err
	}
	return state, nil
}
