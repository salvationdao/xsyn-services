package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"xsyn-services/passport/db"
	"xsyn-services/passport/helpers"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog/log"
)

type EarlyContributorReturn struct {
	Key    string `json:"key"`
	Value  bool   `json:"value"`
	Signed bool   `json:"has_signed"`
	Agreed bool   `json:"agreed"`
}

type EarlyContributorSign struct {
	Address   string `json:"address"`
	Signature string `json:"signature"`
}

func (api *API) CheckUserEarlyContributor(w http.ResponseWriter, r *http.Request) (int, error) {
	messageSignatureData := r.URL.Query()
	address, ok := messageSignatureData["address"]
	if !ok {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("url query param not given"), "Failed to provide address")
	}
	lowerAddress := strings.ToLower(address[0])
	isEarly, user, err := db.IsUserEarlyContributor(r.Context(), api.Conn, lowerAddress)
	if err != nil {
		found := pgxscan.NotFound(err)
		if !found {
			return http.StatusInternalServerError, terror.Error(err, "Failed to check early contributor")
		}
		earlyContributorReturn := EarlyContributorReturn{Key: "is-early", Value: isEarly, Signed: false, Agreed: false}
		if user.Agree.Valid {
			earlyContributorReturn.Agreed = user.Agree.Bool
		}
		if user.Message.Valid {
			if user.Message.String != "" {
				earlyContributorReturn.Signed = true
			}
		}

		return helpers.EncodeJSON(w, earlyContributorReturn)
	}
	earlyContributorReturn := EarlyContributorReturn{Key: "is-early", Value: isEarly, Signed: false, Agreed: false}
	if user.Agree.Valid {
		earlyContributorReturn.Agreed = user.Agree.Bool
	}
	if user.Message.Valid {
		if user.Message.String != "" {
			earlyContributorReturn.Signed = true
		}
	}

	return helpers.EncodeJSON(w, earlyContributorReturn)
}

func (api *API) EarlyContributorSignMessage(w http.ResponseWriter, r *http.Request) (int, error) {
	messageSignatureData := r.URL.Query()
	message, ok := messageSignatureData["message"]
	if !ok {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("message query param not given"), "Failed to provide message in URL param")
	}
	messageHex, ok := messageSignatureData["hex"]
	if !ok {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("hex message query param not given"), "Failed to provide hexxed message in URL param")
	}
	signature, ok := messageSignatureData["signature"]
	if !ok {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("signature query param not given"), "Failed to provide signature in URL param")
	}
	address, ok := messageSignatureData["address"]
	if !ok {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("address query param not given"), "Failed to provide address in URL param")
	}
	agree, ok := messageSignatureData["agree"]
	if !ok {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("signature agree query param not given"), "Failed to terms agreement in URL param")
	}

	verify, _, err := verifySignature(common.HexToAddress(address[0]), []byte(message[0]), []byte(signature[0]))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Error verifying signature")
	}
	if !verify {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("couldnt verify signature"), "Error verifying signature")
	}
	lowerAddress := strings.ToLower(address[0])
	agreed, err := strconv.ParseBool(agree[0])
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Error parsing agreement boolean")
	}

	err = db.UserSignMessage(r.Context(), api.Conn, lowerAddress, message[0], signature[0], messageHex[0], agreed)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	isEarly, user, err := db.IsUserEarlyContributor(r.Context(), api.Conn, lowerAddress)
	if err != nil {
		found := pgxscan.NotFound(err)
		if !found {
			return http.StatusInternalServerError, terror.Error(err, "Failed to check early contributor")
		}
		earlyContributorReturn := EarlyContributorReturn{Key: "is-early", Value: isEarly, Signed: false, Agreed: false}
		if user.Agree.Valid {
			earlyContributorReturn.Agreed = user.Agree.Bool
		}
		if user.Message.Valid {
			if user.Message.String != "" {
				earlyContributorReturn.Signed = true
			}
		}

		return helpers.EncodeJSON(w, earlyContributorReturn)
	}
	earlyContributorReturn := EarlyContributorReturn{Key: "is-early", Value: isEarly, Signed: false, Agreed: false}
	if user.Agree.Valid {
		earlyContributorReturn.Agreed = user.Agree.Bool
	}
	if user.Message.Valid {
		if user.Message.String != "" {
			earlyContributorReturn.Signed = true
		}
	}
	return helpers.EncodeJSON(w, earlyContributorReturn)
}

func prefixedHash(data []byte) []byte {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	return crypto.Keccak256([]byte(msg))
}

func verifySignature(address common.Address, message, messageSignature []byte) (bool, []byte, error) {
	// https://gist.github.com/dcb9/385631846097e1f59e3cba3b1d42f3ed#file-eth_sign_verify-go
	// https://stackoverflow.com/questions/49085737/geth-ecrecover-invalid-signature-recovery-id
	// log.Info().Msg(fmt.Sprintf("Veryfying %v", messageSignature[32]))
	decodedSig, err := hexutil.Decode(string(messageSignature))
	if err != nil {
		return false, nil, fmt.Errorf("sig to pub: %w", err)
	}
	if decodedSig[64] == 0 || decodedSig[64] == 1 {
		log.Info().Msg("signature good 0 or 1")
		//https://ethereum.stackexchange.com/questions/102190/signature-signed-by-go-code-but-it-cant-verify-on-solidity
		decodedSig[64] += 27
	} else if decodedSig[64] != 27 && decodedSig[64] != 28 {
		return false, nil, terror.Error(fmt.Errorf("decode sig invalid %v", decodedSig[64]))
	}
	decodedSig[64] -= 27

	prefixedMsg := prefixedHash(message)

	recoveredPublicKey, err := crypto.Ecrecover(prefixedMsg, decodedSig)
	if err != nil {
		return false, nil, terror.Error(err, "Failed to recover public key from signature")
	}
	secp256k1RecoveredPublicKey, err := crypto.UnmarshalPubkey(recoveredPublicKey)
	if err != nil {
		return false, nil, terror.Error(err, "Failed to unmarshal public key")
	}

	recoveredAddress := crypto.PubkeyToAddress(*secp256k1RecoveredPublicKey).Hex()

	matches := strings.EqualFold(common.HexToAddress(string(recoveredAddress)).Hex(), address.Hex())
	if !matches {
		return false, nil, nil
	}
	return true, prefixedMsg, nil
}
