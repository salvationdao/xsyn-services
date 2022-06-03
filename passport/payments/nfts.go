package payments

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"xsyn-services/boiler"
	"xsyn-services/passport/asset"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"

	"github.com/ethereum/go-ethereum/common"
)

type NFTOwnerStatus struct {
	Collection    common.Address
	Owner         common.Address
	OnChainStatus db.OnChainStatus
}

func UpdateOwners(nftStatuses map[int]*NFTOwnerStatus, collection *boiler.Collection) (int, int, error) {
	l := passlog.L.With().Str("svc", "avant_nft_ownership_update").Logger()

	updated := 0
	skipped := 0

	l.Debug().Int("records", len(nftStatuses)).Msg("processing new owners for NFT")

	for tokenID, nftStatus := range nftStatuses {
		l.Debug().
			Int("token_id", tokenID).
			Str("collection", nftStatus.Collection.Hex()).
			Str("owner", nftStatus.Owner.Hex()).
			Str("on_chain_status", string(nftStatus.OnChainStatus)).
			Msg("processing new owner for NFT")

		userAsset, err := boiler.UserAssets(
			boiler.UserAssetWhere.CollectionID.EQ(collection.ID),
			boiler.UserAssetWhere.TokenID.EQ(int64(tokenID)),
		).One(passdb.StdConn)
		if err != nil && errors.Is(err, sql.ErrNoRows) {
			l.Debug().Str("collection_addr", collection.MintContract.String).Int("external_token_id", tokenID).Msg("item not found")
			skipped++
			continue
		} else if err != nil {
			return 0, 0, fmt.Errorf("get purchased item: %w", err)
		}

		// on chain user may not exist in our db
		onChainOwner, err := CreateOrGetUser(nftStatus.Owner)
		if err != nil {
			return 0, 0, fmt.Errorf("get or create onchain user: %w", err)
		}

		// of chain user has to exist, it is the current owner
		offChainOwner, err := boiler.FindUser(passdb.StdConn, userAsset.OwnerID)
		if err != nil {
			return 0, 0, fmt.Errorf("get offchain user: %w", err)
		}

		offChainAddr := common.HexToAddress(offChainOwner.PublicAddress.String)
		onChainAddr := common.HexToAddress(onChainOwner.PublicAddress.String)

		l.Debug().
			Str("off_chain_user", offChainAddr.Hex()).
			Str("on_chain_user", onChainAddr.Hex()).
			Bool("matches", offChainAddr.Hex() == onChainAddr.Hex()).
			Msg("check if nft owners match")

		updatedBool := false

		// if the owner is different, transfer asset to new owner
		if offChainAddr.Hex() != onChainAddr.Hex() {
			l.Debug().
				Str("new_owner", onChainOwner.ID).
				Str("old_owner", offChainOwner.ID).
				Str("item_id", userAsset.ID).
				Msg("setting new nft owner")

			userAsset, _, err = asset.TransferAsset(userAsset.Hash, offChainOwner.ID, onChainOwner.ID, "", null.String{})
			if err != nil {
				passlog.L.Error().Err(err).
					Str("userAsset.Hash", userAsset.Hash).
					Str("offChainOwner.ID", offChainOwner.ID).
					Str("onChainOwner.ID", onChainOwner.ID).
					Msg("failed to transfer asset - UpdateOwners")
				return 0, 0, fmt.Errorf("set new nft owner: %w", err)
			}
			updatedBool = true
		}

		if string(nftStatus.OnChainStatus) != userAsset.OnChainStatus {
			userAsset.OnChainStatus = string(nftStatus.OnChainStatus)
			_, err = userAsset.Update(passdb.StdConn, boil.Infer())
			if err != nil {
				return 0, 0, err
			}
			updatedBool = true
		}
		if updatedBool {
			updated++
		}
	}

	return updated, skipped, nil
}
