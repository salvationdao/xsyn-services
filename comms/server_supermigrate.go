package comms

import (
	"encoding/json"
	"passport/db/boiler"
	"passport/passdb"
)

func (s *S) SuperMigrate(req GetAllMechsReq, resp *GetAll) error {
	assets, err := boiler.XsynAssets().All(passdb.StdConn)
	if err != nil {
		return err
	}
	metadata, err := boiler.XsynMetadata().All(passdb.StdConn)
	if err != nil {
		return err
	}
	store, err := boiler.XsynStores().All(passdb.StdConn)
	if err != nil {
		return err
	}
	users, err := boiler.Users().All(passdb.StdConn)
	if err != nil {
		return err
	}

	factions, err := boiler.Factions().All(passdb.StdConn)
	if err != nil {
		return err
	}

	b1, err := json.Marshal(assets)
	if err != nil {
		return err
	}
	b2, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	b3, err := json.Marshal(store)
	if err != nil {
		return err
	}
	b4, err := json.Marshal(users)
	if err != nil {
		return err
	}
	b5, err := json.Marshal(factions)
	if err != nil {
		return err
	}
	resp.AssetPayload = b1
	resp.MetadataPayload = b2
	resp.StorePayload = b3
	resp.UserPayload = b4
	resp.FactionPayload = b5
	return nil
}
