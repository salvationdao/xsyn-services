package comms

import (
	"fmt"
	"math/big"
	"net"
	"net/rpc"
	"passport/api"
	"passport/db"
	"sync"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
	"github.com/sasha-s/go-deadlock"
)

// for sups trickle handler
type TickerPoolCache struct {
	deadlock.Mutex
	TricklingAmountMap map[string]*big.Int
}

type C struct {
	UserCacheMap *api.UserCacheMap
	MessageBus   *messagebus.MessageBus
	Txs          *api.Transactions
	Log          *zerolog.Logger
	Conn         db.Conn
	DistLock     deadlock.Mutex

	TickerPoolCache *TickerPoolCache // user for sup trickle process
	HubSessionIDMap *sync.Map
}

func New(
	userCacheMap *api.UserCacheMap,
	messageBus *messagebus.MessageBus,
	txs *api.Transactions,
	log *zerolog.Logger,
	conn *pgxpool.Pool,
	cm *sync.Map,
) *C {
	result := &C{
		UserCacheMap: userCacheMap,
		MessageBus:   messageBus,
		Txs:          txs,
		Log:          log,
		Conn:         conn,
		TickerPoolCache: &TickerPoolCache{
			deadlock.Mutex{},
			make(map[string]*big.Int),
		},
		HubSessionIDMap: cm,
	}

	// run a ticker to clear up the client map
	go func() {
		for {
			time.Sleep(10 * time.Second)
			now := time.Now()
			result.HubSessionIDMap.Range(func(key, value interface{}) bool {
				registerTime, ok := value.(time.Time)
				if !ok {
					result.HubSessionIDMap.Delete(key)
					return true
				}

				// clean up the session key, if it is not verified for 10 minutes
				if now.Sub(registerTime).Minutes() == 10 {
					result.HubSessionIDMap.Delete(key)
					return true
				}

				return true
			})
		}
	}()

	return result
}

func (c *C) listen(addrStr ...string) ([]net.Listener, error) {
	listeners := make([]net.Listener, len(addrStr))
	for i, a := range addrStr {
		c.Log.Info().Str("addr", a).Msg("registering RPC server")
		addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("0.0.0.0:%s", a))
		if err != nil {
			c.Log.Err(err).Str("addr", a).Msg("registering RPC server")
			return listeners, nil
		}

		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return listeners, err
		}

		listeners[i] = l
	}

	return listeners, nil
}

func Start(c *C) error {
	listeners, err := c.listen("10001", "10002", "10003", "10004", "10005", "10006")
	if err != nil {
		return err
	}
	for _, l := range listeners {
		s := rpc.NewServer()
		err = s.Register(c)
		if err != nil {
			return err
		}

		c.Log.Info().Str("addr", l.Addr().String()).Msg("starting up RPC server")
		go s.Accept(l)
	}

	return nil
}
