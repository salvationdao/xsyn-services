package db

import (
	"errors"
	"strconv"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type KVKey string

const KeyCannyApiKey KVKey = "api_key_canny"
const KeyLatestWithdrawBlockBSC KVKey = "latest_withdraw_block"
const KeyLatestWithdrawBlockETH KVKey = "latest_withdraw_block_eth"
const KeyLatestDepositBlockBSC KVKey = "latest_deposit_block"
const KeyLatestDepositBlockETH KVKey = "latest_deposit_block_eth"
const KeyLatest1155WithdrawBlock KVKey = "latest_1155_withdraw_block"
const KeyLatest1155DepositBlock KVKey = "latest_1155_deposit_block"
const KeyLatestBUSDBlock KVKey = "latest_busd_block"
const KeyLatestUSDCBlock KVKey = "latest_usdc_block"
const KeyLatestETHBlock KVKey = "latest_eth_block"
const KeyLatestBNBBlock KVKey = "latest_bnb_block"
const KeyEnableWithdrawRollback KVKey = "enable_withdraw_rollback"
const KeyAvantFailureCount KVKey = "avant_failure_count"
const KeyAvantSuccessCount KVKey = "avant_success_count"

const KeyEnableSyncPayments KVKey = "enable_sync_payments"
const KeyEnableSyncDeposits KVKey = "enable_sync_deposits"
const KeyEnableSyncNFTOwners KVKey = "enable_sync_nft_owners"
const KeyEnableSyncWithdraw KVKey = "enable_sync_withdraw"
const KeyEnableSync1155 KVKey = "enable_sync_1155"
const KeyEnableSyncSale KVKey = "enable_sync_sale"

const KeySupsToUSD KVKey = "sups_to_usd"
const KeyBNBToUSD KVKey = "bnb_to_usd"
const KeyEthToUSD KVKey = "eth_to_usd"

const KeyPurchaseSupsFloorPrice KVKey = "purchase_sups_floor_price"
const KeyPurchaseSupsMarketPriceMultiplier KVKey = "purchase_sups_market_price_multiplier"
const KeyEnablePassportExchangeRate KVKey = "enable_passport_exchange_rate"
const KeyEnablePassportExchangeRateAfterETHBlock KVKey = "passport_exchange_rate_after_eth_block"
const KeyEnablePassportExchangeRateAfterBSCBlock KVKey = "passport_exchange_rate_after_bsc_block"

const KeySUPSPurchaseContract KVKey = "contract_purchase_address"
const KeySUPSWithdrawContractBSC KVKey = "contract_withdraw_address"
const KeySUPSWithdrawContractETH KVKey = "contract_withdraw_address_eth"

const KeySyndicateRegisterFee KVKey = "syndicate_create_fee"
const KeySyndicateRegisterFeeCut KVKey = "syndicate_create_fee_cut"

const KeyEnableEthDeposits = "enable_eth_deposits"
const KeyEnableEthWithdraws = "enable_eth_withdraws"
const KeyEnableBscDeposits = "enable_bsc_deposits"
const KeyEnableBscWithdraws = "enable_bsc_withdraws"

func get(key KVKey) string {
	exists, err := boiler.KVS(boiler.KVWhere.Key.EQ(string(key))).Exists(passdb.StdConn)
	if err != nil {
		passlog.L.Err(err).Str("key", string(key)).Msg("could not check kv exists")
		return ""
	}
	if !exists {
		passlog.L.Err(errors.New("kv does not exist")).Str("key", string(key)).Msg("kv does not exist")
		return ""
	}
	kv, err := boiler.KVS(boiler.KVWhere.Key.EQ(string(key))).One(passdb.StdConn)
	if err != nil {
		passlog.L.Err(err).Str("key", string(key)).Msg("could not get kv")
		return ""
	}
	return kv.Value
}

func put(key KVKey, value string) {
	kv := boiler.KV{
		Key:   string(key),
		Value: value,
	}
	err := kv.Upsert(passdb.StdConn, true, []string{boiler.KVColumns.Key}, boil.Whitelist(boiler.KVColumns.Value), boil.Infer())
	if err != nil {
		passlog.L.Err(err).Msg("could not put kv")
		return
	}
}

func GetStr(key KVKey) string {
	return get(key)

}
func GetStrWithDefault(key KVKey, defaultValue string) string {
	vStr := get(key)
	if vStr == "" {
		PutStr(key, defaultValue)
		return defaultValue
	}

	return GetStr(key)
}
func PutStr(key KVKey, value string) {
	put(key, value)
}
func GetBool(key KVKey) bool {
	v := get(key)
	b, err := strconv.ParseBool(v)
	if err != nil {
		passlog.L.Err(err).Str("key", string(key)).Str("val", v).Msg("could not parse boolean")
		return false
	}
	return b
}

func GetBoolWithDefault(key KVKey, defaultValue bool) bool {
	vStr := get(key)
	if vStr == "" {
		PutBool(key, defaultValue)
		return defaultValue
	}

	return GetBool(key)
}
func PutBool(key KVKey, value bool) {
	put(key, strconv.FormatBool(value))
}

func GetInt(key KVKey) int {
	vStr := get(key)
	v, err := strconv.Atoi(vStr)
	if err != nil {
		passlog.L.Err(err).Str("key", string(key)).Str("val", vStr).Msg("could not parse int")
		return 0
	}
	return v
}

func GetIntWithDefault(key KVKey, defaultValue int) int {
	vStr := get(key)
	if vStr == "" {
		PutInt(key, defaultValue)
		return defaultValue
	}

	return GetInt(key)
}

func PutInt(key KVKey, value int) {
	put(key, strconv.Itoa(value))
}

func GetDecimal(key KVKey) decimal.Decimal {
	vStr := get(key)
	v, err := decimal.NewFromString(vStr)
	if err != nil {
		passlog.L.Err(err).Str("key", string(key)).Str("val", vStr).Msg("could not parse decimal")
		return decimal.Zero
	}
	return v
}
func GetDecimalWithDefault(key KVKey, defaultValue decimal.Decimal) decimal.Decimal {
	vStr := get(key)

	if vStr == "" {
		PutDecimal(key, defaultValue)
		return defaultValue
	}
	return GetDecimal(key)
}

func PutDecimal(key KVKey, value decimal.Decimal) {
	put(key, value.String())
}
func GetTime(key KVKey) time.Time {
	vStr := get(key)
	t, err := time.Parse(time.RFC3339, vStr)
	if err != nil {
		passlog.L.Err(err).Str("key", string(key)).Str("val", vStr).Msg("could not parse time")
		return time.Time{}
	}
	return t
}
func GetTimeWithDefault(key KVKey, defaultValue time.Time) time.Time {
	vStr := get(key)
	if vStr == "" {
		PutTime(key, defaultValue)
		return defaultValue
	}

	return GetTime(key)
}
func PutTime(key KVKey, value time.Time) {
	put(key, value.Format(time.RFC3339))
}
