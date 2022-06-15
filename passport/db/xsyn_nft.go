package db

import (
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"xsyn-services/boiler"
	"xsyn-services/passport/passdb"
)

func Withdraw1155AssetWithPendingRollback(count, externalTokenID int, ownerID string) error {
	q := `WITH ass AS (
    UPDATE user_assets_1155 ua1 set count = count - $1
           WHERE ua1.owner_id = $2 AND ua1.external_token_id = $3 AND ua1.service_id is null
           RETURNING ua1.owner_id, ua1.id
	) INSERT INTO pending_1155_rollback(user_id, asset_id, count, refunded_at)
	SELECT ass.owner_id, ass.id, $1, NOW() + interval '10' MINUTE
	FROM ass;`

	_, err := boiler.NewQuery(qm.SQL(q, count, ownerID, externalTokenID)).Exec(passdb.StdConn)
	if err != nil {
		return err
	}
	return nil
}
