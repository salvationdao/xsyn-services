package passdb

import (
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
)

var Conn *pgxpool.Pool
var StdConn *sql.DB

func New(conn *pgxpool.Pool, stdConn *sql.DB) error {
	if Conn != nil {
		return fmt.Errorf("db already initialised")
	}
	if StdConn != nil {
		return fmt.Errorf("db already initialised")
	}
	StdConn = stdConn
	Conn = conn
	return nil
}
