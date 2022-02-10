package db

import (
	"passport"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
	"golang.org/x/net/context"
)

// StateGet gets the latest state
func StateGet(ctx context.Context, conn Conn) (*passport.State, error) {
	state := &passport.State{}
	q := `SELECT * FROM state`
	err := pgxscan.Get(ctx, conn, state, q)
	if err != nil {
		return nil, terror.Error(err)
	}
	return state, nil
}

// UpdateLatestETHBlock updates the latest eth block checked
func UpdateLatestETHBlock(ctx context.Context, conn Conn, latestBlock uint64) (*passport.State, error) {
	state := &passport.State{}
	q := `UPDATE state SET latest_eth_block = $1 RETURNING latest_eth_block, latest_bsc_block`
	err := pgxscan.Get(ctx, conn, state, q, latestBlock)
	if err != nil {
		return nil, terror.Error(err)
	}
	return state, nil
}

// UpdateLatestBSCBlock updates the latest eth block checked
func UpdateLatestBSCBlock(ctx context.Context, conn Conn, latestBlock uint64) (*passport.State, error) {
	state := &passport.State{}
	q := `UPDATE state SET latest_bsc_block = $1 RETURNING latest_eth_block, latest_bsc_block`
	err := pgxscan.Get(ctx, conn, state, q, latestBlock)
	if err != nil {
		return nil, terror.Error(err)
	}
	return state, nil
}
