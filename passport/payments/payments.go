package payments

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"xsyn-services/passport/db"
	"xsyn-services/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
)

type UserCacheMap interface {
	Transact(nt *types.NewTransaction) (decimal.Decimal, decimal.Decimal, string, error)
}

const SUPDecimals = 18

func CreateOrGetUser(ctx context.Context, conn *pgxpool.Pool, userAddr common.Address) (*types.User, error) {
	var user *types.User
	var err error
	user, err = db.UserByPublicAddress(ctx, conn, userAddr)
	if errors.Is(err, pgx.ErrNoRows) {
		user = &types.User{}
		user.Username = userAddr.Hex()
		user.PublicAddress = null.NewString(userAddr.Hex(), true)
		user.RoleID = types.UserRoleMemberID
		err := db.UserCreate(ctx, conn, user)
		if err != nil {
			return nil, terror.Error(err)
		}
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, terror.Error(err)
	}
	return user, nil
}

func ProcessValues(sups string, inputValue string, inputTokenDecimals int) (decimal.Decimal, decimal.Decimal, error) {
	outputAmt, err := decimal.NewFromString(sups)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	bigOutputAmt := outputAmt.Shift(1 * types.SUPSDecimals)
	inputAmt, err := decimal.NewFromString(inputValue)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	return inputAmt, bigOutputAmt, nil
}

func StoreRecord(ctx context.Context, fromUserID types.UserID, toUserID types.UserID, ucm UserCacheMap, record *PurchaseRecord) error {
	input, output, err := ProcessValues(record.Sups, record.ValueInt, record.ValueDecimals)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("purchased %s SUPS for %s [%s]", output.Shift(-1*types.SUPSDecimals).StringFixed(4), input.Shift(-1*int32(record.ValueDecimals)).StringFixed(4), strings.ToUpper(record.Symbol))
	trans := &types.NewTransaction{
		To:                   toUserID,
		From:                 fromUserID,
		Amount:               output,
		TransactionReference: types.TransactionReference(record.TxHash),
		Description:          msg,
		Group:                types.TransactionGroupStore,
	}

	_, _, _, err = ucm.Transact(trans)
	if err != nil {
		return fmt.Errorf("create tx entry for tx %s: %w", record.TxHash, err)
	}
	return nil
}

func BUSD() ([]*PurchaseRecord, error) {
	currentBlock := db.GetInt(db.KeyLatestBUSDBlock)
	records, latestBlock, err := getPurchaseRecords(BUSDPurchasePath, currentBlock, false)
	if err != nil {
		return nil, err
	}
	if latestBlock > currentBlock {
		db.PutInt(db.KeyLatestBUSDBlock, latestBlock)
	}
	return records, nil
}

func USDC() ([]*PurchaseRecord, error) {
	currentBlock := db.GetInt(db.KeyLatestUSDCBlock)
	records, latestBlock, err := getPurchaseRecords(USDCPurchasePath, currentBlock, false)
	if err != nil {
		return nil, err
	}

	if latestBlock > currentBlock {
		db.PutInt(db.KeyLatestUSDCBlock, latestBlock)
	}

	return records, nil
}

func ETH() ([]*PurchaseRecord, error) {
	currentBlock := db.GetInt(db.KeyLatestETHBlock)
	records, latestBlock, err := getPurchaseRecords(ETHPurchasePath, currentBlock, false)
	if err != nil {
		return nil, err
	}

	if latestBlock > currentBlock {
		db.PutInt(db.KeyLatestETHBlock, latestBlock)
	}
	return records, nil
}

func BNB() ([]*PurchaseRecord, error) {
	currentBlock := db.GetInt(db.KeyLatestBNBBlock)
	records, latestBlock, err := getPurchaseRecords(BNBPurchasePath, currentBlock, false)
	if err != nil {
		return nil, err
	}
	if latestBlock > currentBlock {
		db.PutInt(db.KeyLatestBNBBlock, latestBlock)
	}

	return records, nil
}
