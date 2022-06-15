package api

import (
	"fmt"
	"xsyn-services/passport/db"

	"github.com/ninja-software/terror/v2"
)

var (
	ErrCheckDBQuery = fmt.Errorf("error: executing db query")
	ErrCheckDBDirty = fmt.Errorf("db is dirty")
)

// check checks server is working correctly
func check() error {
	err := db.IsSchemaDirty()
	if err != nil {
		return terror.Error(ErrCheckDBQuery)
	}
	return nil
}
