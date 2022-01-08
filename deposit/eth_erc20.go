package deposit

import "github.com/ethereum/go-ethereum/common"

type ETHERC20Listener struct {
	ToWhitelist []common.Address
	ErrChan     chan error
	stop        chan bool
}

func NewETHERC20Listener(toWhitelist []common.Address) *ETHERC20Listener {
	l := &ETHERC20Listener{
		ErrChan:     make(chan error),
		stop:        make(chan bool),
		ToWhitelist: toWhitelist,
	}
	return l

}
func (l *ETHERC20Listener) Chain() string {
	return "mainnet"
}
func (l *ETHERC20Listener) Listen() chan ERC20Transfer {
	ch := make(chan ERC20Transfer)
	go func() {
		for {
			select {
			case <-l.stop:
				return
			}
		}
	}()
	return ch

}
func (l *ETHERC20Listener) Stop() {

}
func (l *ETHERC20Listener) Error() chan error {
	return nil
}
