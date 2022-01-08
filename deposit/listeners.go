package deposit

import "math/big"

type ERC20Transfer struct {
	Chain         string
	Confirmations int
	Contract      string
	TXID          string
	From          string
	To            string
	Value         *big.Int
}

type ERC20 interface {
	Chain() string
	Listen() chan ERC20Transfer
	Stop()
	Error() chan error
}
type ERC721Transfer struct {
	Chain         string
	Confirmations int
	Contract      string
	TXID          string
	From          string
	To            string
	TokenID       int64
}
type ERC721 interface {
	Chain() string
	Listen() chan ERC721Transfer
	Stop()
	Error() chan error
}
type ERC1155Transfer struct {
	Chain         string
	Confirmations int
	Contract      string
	TXID          string
	From          string
	To            string
	TokenID       int64
	Value         int64
}
type ERC1155 interface {
	Chain() string
	Listen() chan ERC1155Transfer
	Stop()
	Error() chan error
}
