package seed

import (
	"context"
	"fmt"
	"passport"
	"passport/api"
	"passport/db"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/sale/dispersions"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
)

// EarlyContributors seeds private sale users with their dispersions.
// Relies on the exported variable dispersions.go from the
func (s *Seeder) EarlyContributors(ctx context.Context) error {
	dispersionMap := dispersions.All()
	i := 0
	for k, v := range dispersionMap {
		u := &passport.User{}
		u.Username = k.Hex()
		u.PublicAddress = null.NewString(k.Hex(), true)
		u.RoleID = passport.UserRoleMemberID
		err := db.UserCreate(ctx, s.Conn, u)
		if err != nil {
			return terror.Error(err)
		}
		for _, output := range v {
			txID := fmt.Sprintf("%s|%d", uuid.Must(uuid.NewV4()), time.Now().Nanosecond())
			amt := decimal.New(int64(output.Output), 18)
			_, err = api.CreateTransactionEntry(s.TxConn,
				txID,
				*amt.BigInt(),
				u.ID,
				passport.XsynSaleUserID,
				"Supremacy early contributor dispersion.",
				passport.TransactionReference(fmt.Sprintf("supremacy|early_contributor|%d", i)),
			)
			if err != nil {
				return terror.Error(err)
			}
			i++
		}
	}

	return nil
}
