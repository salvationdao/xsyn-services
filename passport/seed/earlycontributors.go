package seed

import (
	"context"
	"fmt"
	"xsyn-services/passport/db"
	"xsyn-services/types"

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
		u := &types.User{}
		u.Username = k.Hex()
		u.PublicAddress = null.NewString(k.Hex(), true)
		u.RoleID = types.UserRoleMemberID
		err := db.UserCreateNoRPC(ctx, s.Conn, u)
		if err != nil {
			return terror.Error(err)
		}
		for _, output := range v {
			amt := decimal.New(int64(output.Output), 18)

			nt := &types.NewTransaction{
				Amount:               amt,
				From:                 types.XsynSaleUserID,
				To:                   u.ID,
				Description:          "Supremacy early contributor dispersion.",
				TransactionReference: types.TransactionReference(fmt.Sprintf("Supremacy early contributor dispersion #%04d", i)),
			}

			q := `INSERT INTO transactions(id, description, transaction_reference, amount, credit, debit)
        				VALUES((SELECT count(*) from transactions), $1, $2, $3, $4, $5);`

			_, err = s.Conn.Exec(ctx, q, nt.Description, nt.TransactionReference, nt.Amount.String(), nt.To, nt.From)
			if err != nil {
				return terror.Error(err)
			}
			i++
		}
	}

	return nil
}
