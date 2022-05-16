package passdb

import (
	"database/sql"
	"fmt"
)

var StdConn *sql.DB

func New(stdConn *sql.DB) error {
	if StdConn != nil {
		return fmt.Errorf("db already initialised")
	}
	StdConn = stdConn
	return nil
}
