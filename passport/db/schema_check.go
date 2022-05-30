package db

import (
	"fmt"
	"xsyn-services/boiler"
	"xsyn-services/passport/passdb"
)

func IsSchemaDirty() error {
	count, err := boiler.SchemaMigrations(
		boiler.SchemaMigrationWhere.Dirty.EQ(true),
	).Count(passdb.StdConn)
	if err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("db is dirty")
	}
	return nil
}
