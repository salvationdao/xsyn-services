package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/api/users"
	"xsyn-services/passport/db"
	"xsyn-services/passport/helpers"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/passport/payments"
	"xsyn-services/types"

	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi/v5"
	"github.com/ninja-software/sale/dispersions"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/supremacy-bridge/bridge"
)

type MaxWithdrawResponse struct {
	MaxWithdraw    string `json:"max_withdraw"`
	TotalWithdrawn string `json:"total_withdrawn"`
	Unlimited      bool   `json:"unlimited"`
}

type CheckWithdrawResponse struct {
	CanWithdraw bool `json:"can_withdraw"`
}

func (api *API) GetMaxWithdrawAmount(w http.ResponseWriter, r *http.Request) (int, error) {
	address := chi.URLParam(r, "address")
	if address == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing address"), "Missing address.")
	}
	user, err := boiler.Users(
		boiler.UserWhere.PublicAddress.EQ(null.StringFrom(address)),
	).One(passdb.StdConn)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return http.StatusInternalServerError, terror.Error(err, "Failed to find users info")
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return http.StatusInternalServerError, terror.Error(err, "Failed to find users info")
	}

	state, err := boiler.States().One(passdb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get state")
	}

	toAddress := common.HexToAddress(address)
	amountCanRefund, infinite, err := dispersions.MaxWithdrawBefore(toAddress, time.Now(), state.WithdrawStartAt, state.CliffEndAt, state.DripStartAt)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get max withdraw amount")
	}

	// extraWithdraw := db.ExtraWithdraw(passport.UserID(uuid.Must(uuid.FromString(user.ID))))
	// passlog.L.Debug().Str("amount", extraWithdraw.Shift(-18).StringFixed(4)).Str("user_id", user.ID).Msg("extra refund")
	// amountCanRefund = amountCanRefund.Add(extraWithdraw)

	maxWithdrawResponse := MaxWithdrawResponse{MaxWithdraw: "0", Unlimited: infinite}
	if infinite {
		return helpers.EncodeJSON(w, maxWithdrawResponse)
	}
	maxWithdrawResponse.MaxWithdraw = amountCanRefund.BigInt().String()
	//amountCanRefund = decimal.NewFromBigInt(amountCanRefund.BigInt(), -18)

	refunds, err := boiler.PendingRefunds(
		qm.Where("user_id = ? AND is_refunded = false", user.ID),
	).All(passdb.StdConn)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		maxWithdrawResponse.MaxWithdraw = amountCanRefund.BigInt().String()
		return helpers.EncodeJSON(w, maxWithdrawResponse)
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Info().Msg(err.Error())
		return http.StatusInternalServerError, terror.Error(err, "Failed to find users refund details")
	}
	amt := decimal.Zero
	for _, refund := range refunds {
		amt = amt.Add(refund.AmountSups)
	}

	maxWithdrawResponse.TotalWithdrawn = amt.BigInt().String()

	return helpers.EncodeJSON(w, maxWithdrawResponse)
}

type HoldingResp struct {
	Amount string `json:"amount"`
}

func (api *API) HoldingSups(w http.ResponseWriter, r *http.Request) (int, error) {
	address := common.HexToAddress(chi.URLParam(r, "user_address"))
	u, err := boiler.Users(boiler.UserWhere.PublicAddress.EQ(null.StringFrom(address.String()))).One(passdb.StdConn)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		passlog.L.Error().Str("user_address", address.Hex()).Err(err).Msg("failed to find user by public address")
		return http.StatusBadRequest, err
	}

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		err = json.NewEncoder(w).Encode(&HoldingResp{Amount: decimal.Zero.String()})
		if err != nil {
			return http.StatusInternalServerError, err
		}
		return http.StatusOK, nil
	}

	exists, err := boiler.PendingRefunds(
		boiler.PendingRefundWhere.UserID.EQ(u.ID),
		boiler.PendingRefundWhere.IsRefunded.EQ(false),      // Only those not refunded by avant scraper yet
		boiler.PendingRefundWhere.RefundedAt.GT(time.Now()), // Only those with unexpired signatures
	).Exists(passdb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return http.StatusInternalServerError, err
	}

	if !exists {
		err = json.NewEncoder(w).Encode(&HoldingResp{Amount: decimal.Zero.String()})
		if err != nil {
			return http.StatusInternalServerError, err
		}
		return http.StatusOK, nil
	}

	records, err := boiler.PendingRefunds(
		boiler.PendingRefundWhere.UserID.EQ(u.ID),
		boiler.PendingRefundWhere.IsRefunded.EQ(false),      // Only those not refunded by avant scraper yet
		boiler.PendingRefundWhere.RefundedAt.GT(time.Now()), // Only those with unexpired signatures
		boiler.PendingRefundWhere.RefundCanceledAt.IsNull(),
	).All(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Str("pending_refunds", u.ID).Err(err).Msg("failed to find pending refunds")
		return http.StatusInternalServerError, err
	}
	total := decimal.Zero
	for _, record := range records {
		total = total.Add(record.AmountSups)
	}
	err = json.NewEncoder(w).Encode(&HoldingResp{Amount: total.String()})
	if err != nil {
		passlog.L.Error().Str("json_encode", err.Error()).Err(err).Msg("failed to encode json")
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

// WithdrawSups
// Flow to withdraw sups
// get nonce from withdraw contract
// send nonce, amount and user wallet addr to server,
// server validates they have enough sups
// server generates a sig and returns it
// submit that sig to withdraw contract withdrawSups func
// listen on backend for update
func (api *API) WithdrawSups(w http.ResponseWriter, r *http.Request) (int, error) {
	// todo: check passed in chain is valid
	// todo: check withdrawals are enabled for passed in chain

	address := chi.URLParam(r, "address")
	if address == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing address"), "Missing address.")
	}

	nonce := chi.URLParam(r, "nonce")
	if nonce == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing nonce"), "Missing nonce.")
	}

	amount := chi.URLParam(r, "amount")
	if amount == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing amount"), "Missing amount.")
	}

	chain := chi.URLParam(r, "chain")
	if chain == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing chain"), "Missing chain id.")
	}
	chainInt, err := strconv.Atoi(chain)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Invalid chain id.")
	}

	if chainInt == api.Web3Params.EthChainID {
		withdrawsEnabled := db.GetBool(db.KeyEnableBscWithdraws)

		if !withdrawsEnabled {
			return http.StatusServiceUnavailable, terror.Error(fmt.Errorf("bsc withdraws disabled"), "Withdraws on BSC are currently disabled while we migrate to Ethereum.")
		}
	} else if chainInt == api.Web3Params.BscChainID {
		withdrawsEnabled := db.GetBool(db.KeyEnableBscWithdraws)

		if !withdrawsEnabled {
			return http.StatusServiceUnavailable, terror.Error(fmt.Errorf("eth withdraws disabled"), "Withdraws on ETH are currently disabled while we migrate to Ethereum.")
		}
	} else {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("chain %d is invalid", chainInt), "Invalid chain id.")
	}

	toAddress := common.HexToAddress(address)

	//checks block_withdraw table and returns if user's connected wallet is found
	blockedExists, err := boiler.BlockWithdraws(
		boiler.BlockWithdrawWhere.PublicAddress.EQ(toAddress.String()),
		boiler.BlockWithdrawWhere.BlockNFTWithdraws.GTE(time.Now()),
	).Exists(passdb.StdConn)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not run checks on public address.")
	}

	if blockedExists {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("user is blocked from withdrawing sups"), "The address connected to this account may not withdraw SUPS.")
	}

	amountBigInt := new(big.Int)
	_, ok := amountBigInt.SetString(amount, 10)
	if !ok {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("failed to parse amount to big int"), "Invalid amount.")
	}

	nonceBigInt := new(big.Int)
	_, ok = nonceBigInt.SetString(nonce, 10)
	if !ok {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("failed to parse nonce to big int"), "Invalid nonce.")
	}

	// check balance
	user, err := users.PublicAddress(common.HexToAddress(address))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to find user with this wallet address.")
	}

	isLocked := user.CheckUserIsLocked("withdrawals")
	if isLocked {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("user: %s, attempting to withdraw while account is locked.", user.ID), "Withdrawals is locked, contact support to unlock.")
	}

	userAccount, err := db.UserBalance(user.ID)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not find SUPS balance")
	}

	if userAccount.Sups.LessThan(decimal.NewFromBigInt(amountBigInt, 0)) {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("user has insufficient funds: %s, %s", userAccount.Sups.String(), amountBigInt), "Insufficient funds.")
	}

	//  sign it
	expiry := time.Now().Add(5 * time.Minute)
	signer := bridge.NewSigner(api.Web3Params.SignerPrivateKey)
	_, messageSig, err := signer.GenerateSignatureWithExpiry(toAddress, amountBigInt, nonceBigInt, big.NewInt(expiry.Unix()))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to create withdraw signature, please try again or contact support.")
	}
	state, err := boiler.States().One(passdb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get state")
	}

	amountCanRefund, infinite, err := dispersions.MaxWithdrawBefore(toAddress, time.Now(), state.WithdrawStartAt, state.CliffEndAt, state.DripStartAt)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get max withdraw amount")
	}

	if !infinite {
		// extraWithdraw := db.ExtraWithdraw(user.ID)
		// passlog.L.Debug().Str("amount", extraWithdraw.Shift(-18).StringFixed(4)).Str("user_id", user.).Msg("extra refund")
		// amountCanRefund = amountCanRefund.Add(extraWithdraw)

		amountCanRefund = decimal.NewFromBigInt(amountCanRefund.BigInt(), -18)
		amt := decimal.Zero
		refunds, err := boiler.PendingRefunds(
			qm.Where("user_id = ? AND is_refunded = false", user.ID),
		).All(passdb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return http.StatusInternalServerError, terror.Error(err, "Failed to find users refund details")
		}
		if err != nil && errors.Is(err, sql.ErrNoRows) {
			if !decimal.NewFromBigInt(amountBigInt, -18).LessThanOrEqual(amountCanRefund) {
				return http.StatusInternalServerError, terror.Error(err, "Failed to withdraw amount")
			}
		} else {
			for _, refund := range refunds {
				amt = amt.Add(decimal.NewFromBigInt(refund.AmountSups.BigInt(), -18))
			}
			amt = amt.Add(decimal.NewFromBigInt(amountBigInt, -18))
			log.Info().Msg(fmt.Sprintf("%v %v %v", amt, amountCanRefund, amountBigInt))
			if !amt.LessThanOrEqual(amountCanRefund) {
				return http.StatusInternalServerError, terror.Error(errors.New("total withdrawn amount exceeds allowable"), "Failed to withdraw amount")
			}
		}
	}

	refundID, err := payments.InsertPendingRefund(api.userCacheMap, types.UserIDFromString(user.ID), decimal.NewFromBigInt(amountBigInt, 0), expiry)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to create withdraw signature, please try again or contact support.")
	}

	err = json.NewEncoder(w).Encode(struct {
		MessageSignature string `json:"messageSignature"`
		Expiry           int64  `json:"expiry"`
		RefundID         string `json:"refundID"`
	}{
		MessageSignature: hexutil.Encode(messageSig),
		Expiry:           expiry.Unix(),
		RefundID:         refundID,
	})
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

type CheckCanWithdrawResp struct {
	WithdrawalsEnabled        bool   `json:"withdrawals_enabled"`
	WithdrawalChain           int    `json:"withdrawal_chain"`
	WithdrawalContractAddress string `json:"withdrawal_contract_address"`
	TokenContractAddress      string `json:"token_contract_address"`
}

func (api *API) CheckCanWithdraw(w http.ResponseWriter, r *http.Request) (int, error) {
	withdrawsEnabledBSC := db.GetBool(db.KeyEnableBscWithdraws)
	withdrawsEnabledETH := db.GetBool(db.KeyEnableEthWithdraws)

	resp := &CheckCanWithdrawResp{}

	if withdrawsEnabledBSC {
		resp.WithdrawalsEnabled = true
		resp.WithdrawalChain = api.Web3Params.BscChainID
		resp.WithdrawalContractAddress = api.Web3Params.SupWithdrawalAddrBSC.Hex()
		resp.TokenContractAddress = api.Web3Params.SupAddrBSC.Hex()
	}
	// we only want one withdrawal open at the same time, ETH takes precedent
	if withdrawsEnabledETH {
		resp.WithdrawalsEnabled = true
		resp.WithdrawalChain = api.Web3Params.EthChainID
		resp.WithdrawalContractAddress = api.Web3Params.SupWithdrawalAddrETH.Hex()
		resp.TokenContractAddress = api.Web3Params.SupAddrETH.Hex()
	}

	return helpers.EncodeJSON(w, resp)
}

type CheckCanDepositResp struct {
	DepositsEnabledETH    bool   `json:"deposits_enabled_eth"`
	SupContractAddressETH string `json:"sup_contract_address_eth"`
	DepositsEnabledBSC    bool   `json:"deposits_enabled_bsc"`
	SupContractAddressBSC string `json:"sup_contract_address_bsc"`
}

func (api *API) CheckCanDeposit(w http.ResponseWriter, r *http.Request) (int, error) {
	return helpers.EncodeJSON(w, &CheckCanDepositResp{
		DepositsEnabledETH:    db.GetBool(db.KeyEnableEthDeposits),
		SupContractAddressETH: api.Web3Params.SupAddrETH.Hex(),
		DepositsEnabledBSC:    db.GetBool(db.KeyEnableBscDeposits),
		SupContractAddressBSC: api.Web3Params.SupAddrBSC.Hex(),
	})
}
