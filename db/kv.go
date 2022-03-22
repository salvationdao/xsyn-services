package db

import (
	"errors"
	"passport/db/boiler"
	"passport/passdb"
	"passport/passlog"
	"strconv"
	"time"

	"github.com/volatiletech/sqlboiler/v4/boil"
)

const KeyLatestWithdrawBlock = "latest_withdraw_block"
const KeyLatestDepositBlock = "latest_deposit_block"
const KeyLatestBUSDBlock = "latest_busd_block"
const KeyLatestUSDCBlock = "latest_usdc_block"
const KeyLatestETHBlock = "latest_eth_block"
const KeyLatestBNBBlock = "latest_bnb_block"
const KeyEnableWithdrawRollback = "enable_withdraw_rollback"

func get(key string) string {
	exists, err := boiler.KVS(boiler.KVWhere.Key.EQ(key)).Exists(passdb.StdConn)
	if err != nil {
		passlog.L.Err(err).Str("key", key).Msg("could not check kv exists")
		return ""
	}
	if !exists {
		passlog.L.Err(errors.New("kv does not exist")).Str("key", key).Msg("kv does not exist")
		return ""
	}
	kv, err := boiler.KVS(boiler.KVWhere.Key.EQ(key)).One(passdb.StdConn)
	if err != nil {
		passlog.L.Err(err).Str("key", key).Msg("could not get kv")
		return ""
	}
	return kv.Value
}

func put(key, value string) {
	kv := boiler.KV{
		Key:   key,
		Value: value,
	}
	err := kv.Upsert(passdb.StdConn, true, []string{boiler.KVColumns.Key}, boil.Whitelist(boiler.KVColumns.Value), boil.Infer())
	if err != nil {
		passlog.L.Err(err).Msg("could not put kv")
		return
	}
}

func GetStr(key string) string {
	return get(key)

}
func GetStrWithDefault(key string, defaultValue string) string {
	vStr := get(key)
	if vStr == "" {
		PutStr(key, defaultValue)
		return defaultValue
	}

	return GetStr(key)
}
func PutStr(key, value string) {
	put(key, value)
}
func GetBool(key string) bool {
	v := get(key)
	b, err := strconv.ParseBool(v)
	if err != nil {
		passlog.L.Err(err).Str("key", key).Str("val", v).Msg("could not parse boolean")
		return false
	}
	return b
}

func GetBoolWithDefault(key string, defaultValue bool) bool {
	vStr := get(key)
	if vStr == "" {
		PutBool(key, defaultValue)
		return defaultValue
	}

	return GetBool(key)
}
func PutBool(key string, value bool) {
	put(key, strconv.FormatBool(value))
}

func GetInt(key string) int {
	vStr := get(key)
	v, err := strconv.Atoi(vStr)
	if err != nil {
		passlog.L.Err(err).Str("key", key).Str("val", vStr).Msg("could not parse int")
		return 0
	}
	return v
}

func GetIntWithDefault(key string, defaultValue int) int {
	vStr := get(key)
	if vStr == "" {
		PutInt(key, defaultValue)
		return defaultValue
	}

	return GetInt(key)
}

func PutInt(key string, value int) {
	put(key, strconv.Itoa(value))
}

func GetTime(key string) time.Time {
	vStr := get(key)
	t, err := time.Parse(time.RFC3339, vStr)
	if err != nil {
		passlog.L.Err(err).Str("key", key).Str("val", vStr).Msg("could not parse time")
		return time.Time{}
	}
	return t
}
func GetTimeWithDefault(key string, defaultValue time.Time) time.Time {
	vStr := get(key)
	if vStr == "" {
		PutTime(key, defaultValue)
		return defaultValue
	}

	return GetTime(key)
}
func PutTime(key string, value time.Time) {
	put(key, value.Format(time.RFC3339))
}
