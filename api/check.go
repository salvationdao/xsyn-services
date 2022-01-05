package api

import (
	"context"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"passport/db"
)

var (
	ErrCheckDBQuery = fmt.Errorf("error: executing db query")
	ErrCheckDBDirty = fmt.Errorf("db is dirty")
)

// check checks server is working correctly
func check(ctx context.Context, conn db.Conn) error {
	count := 0
	err := db.IsSchemaDirty(ctx, conn, &count)
	if err != nil {
		return terror.Error(ErrCheckDBQuery)
	}
	if count > 0 {
		return terror.Error(ErrCheckDBDirty)
	}
	return nil
}
