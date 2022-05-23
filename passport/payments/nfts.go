package payments

import (
	"database/sql"
	"errors"
	"fmt"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gofrs/uuid"
)

var MainnetStaking = common.HexToAddress("0x6476dB7cFfeeBf7Cc47Ed8D4996d1D60608AAf95")
var TestnetStaking = common.HexToAddress("0x0497e0F8FC07DaaAf2BC1da1eace3F5E60d008b8")

type NFTOwnerStatus struct {
	Collection common.Address
	Owner      common.Address
	Stakable   bool
	Unstakable bool
}

func getNFTContract(collectionSlug string, isTestnet bool) (common.Address, error) {
	switch collectionSlug {
	case "supremacy-genesis":
		if isTestnet {
			return common.HexToAddress("0xEEfaF47acaa803176F1711c1cE783e790E4E750D"), nil
		}
		return common.HexToAddress("0x651D4424F34e6e918D8e4D2Da4dF3DEbDAe83D0C"), nil
	case "supremacy-limited-release":
		if isTestnet {
			return common.HexToAddress("0x440e2CcE53F5a61Bb997ac67D8D45a2898daD27b"), nil
		}
		return common.HexToAddress("0xCA949036Ad7cb19C53b54BdD7b358cD12Cf0b810"), nil
	default:
		passlog.L.Warn().Str("svc", "collection slug doesn't exist").Msg(fmt.Sprintf("Collection slug '%s' doesn't exist", collectionSlug))
		return common.HexToAddress("0"), fmt.Errorf("collection slug does not exist")
	}

}

func UpdateOwners(nftStatuses map[int]*NFTOwnerStatus, isTestnet bool, collectionSlug string) (int, int, error) {
	l := passlog.L.With().Str("svc", "avant_nft_ownership_update").Logger()
	NFTAddr, err := getNFTContract(collectionSlug, isTestnet)
	if err != nil {
		return 0, 0, err
	}

	updated := 0
	skipped := 0
	l.Debug().Int("records", len(nftStatuses)).Msg("processing new owners for NFT")
	for tokenID, nftStatus := range nftStatuses {
		l.Debug().Int("token_id", tokenID).Str("collection", nftStatus.Collection.Hex()).Str("owner", nftStatus.Owner.Hex()).Bool("stakable", nftStatus.Stakable).Bool("unstakable", nftStatus.Unstakable).Msg("processing new owner for NFT")
		purchasedItem, err := db.PurchasedItemByMintContractAndTokenID(NFTAddr, tokenID)
		if err != nil && errors.Is(err, sql.ErrNoRows) {
			l.Debug().Str("collection_addr", NFTAddr.Hex()).Int("external_token_id", tokenID).Msg("item not found")
			skipped++
			continue
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return 0, 0, fmt.Errorf("get purchased item: %w", err)
		}
		onChainOwner, err := CreateOrGetUser(nftStatus.Owner)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return 0, 0, fmt.Errorf("get or create onchain user: %w", err)
		}

		offChainOwner, err := boiler.FindUser(passdb.StdConn, purchasedItem.OwnerID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return 0, 0, fmt.Errorf("get offchain user: %w", err)
		}
		offChainAddr := common.HexToAddress(offChainOwner.PublicAddress.String)
		onChainAddr := common.HexToAddress(onChainOwner.PublicAddress.String)
		l.Debug().Str("off_chain_user", offChainAddr.Hex()).Str("on_chain_user", onChainAddr.Hex()).Bool("matches", offChainAddr.Hex() != onChainAddr.Hex()).Msg("check if nft owners match")
		if offChainAddr.Hex() != onChainAddr.Hex() {
			itemID := uuid.Must(uuid.FromString(purchasedItem.ID))
			newOffchainOwnerID := uuid.FromStringOrNil(onChainOwner.ID)
			l.Debug().Str("new_owner", newOffchainOwnerID.String()).Str("item_id", itemID.String()).Msg("setting new nft owner")
			// TODO: Vinnie fix
			//_, err = db.PurchasedItemSetOwner(itemID, newOffchainOwnerID)
			//if err != nil && !errors.Is(err, sql.ErrNoRows) {
			//	return 0, 0, fmt.Errorf("set new nft owner: %w", err)
			//}
			updated++
		}

		l.Debug().Str("off_chain_stakable", purchasedItem.OnChainStatus).Bool("on_chain_stakable", nftStatus.Stakable).Msg("check if nft stakable state matches")
		if nftStatus.Stakable && purchasedItem.OnChainStatus != string(db.STAKABLE) {
			itemID := uuid.Must(uuid.FromString(purchasedItem.ID))

			err = db.PurchasedItemSetOnChainStatus(itemID, db.STAKABLE)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return 0, 0, fmt.Errorf("set new nft status: %w", err)
			}

			updated++
		}
		l.Debug().Str("off_chain_unstakable", purchasedItem.OnChainStatus).Bool("on_chain_unstakable", nftStatus.Unstakable).Msg("check if nft unstakable state matches")
		if nftStatus.Unstakable && purchasedItem.OnChainStatus != string(db.UNSTAKABLE) {
			itemID := uuid.Must(uuid.FromString(purchasedItem.ID))

			err = db.PurchasedItemSetOnChainStatus(itemID, db.UNSTAKABLE)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return 0, 0, fmt.Errorf("set new nft status: %w", err)
			}
			updated++
		}

	}

	return updated, skipped, nil
}
