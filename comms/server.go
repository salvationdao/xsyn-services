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

type S struct {
	UserCacheMap *api.UserCacheMap
	MessageBus   *messagebus.MessageBus
	Txs          *api.Transactions
	Log          *zerolog.Logger
	Conn         db.Conn
	DistLock     deadlock.Mutex

	TickerPoolCache *TickerPoolCache // user for sup trickle process
	HubSessionIDMap *sync.Map
}

func NewServer(
	userCacheMap *api.UserCacheMap,
	messageBus *messagebus.MessageBus,
	txs *api.Transactions,
	log *zerolog.Logger,
	conn *pgxpool.Pool,
	cm *sync.Map,
) *S {
	result := &S{
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

func (s *S) listen(addrStr ...string) ([]net.Listener, error) {
	listeners := make([]net.Listener, len(addrStr))
	for i, a := range addrStr {
		s.Log.Info().Str("addr", a).Msg("registering RPC server")
		addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("0.0.0.0:%s", a))
		if err != nil {
			s.Log.Err(err).Str("addr", a).Msg("registering RPC server")
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

func StartServer(s *S) error {
	listeners, err := s.listen("10000",
		"10001",
		"10002",
		"10003",
		"10004",
		"10005",
		"10006",
		"10007",
		"10008",
		"10009",
		"10010",
		"10011",
		"10012",
		"10013",
		"10014",
		"10015",
		"10016",
		"10017",
		"10018",
		"10019",
		"10020",
		"10021",
		"10022",
		"10023",
		"10024",
		"10025",
		"10026",
		"10027",
		"10028",
		"10029",
		"10030",
		"10031",
		"10032",
		"10033",
		"10034")
	if err != nil {
		return err
	}
	for _, l := range listeners {
		srv := rpc.NewServer()
		err = srv.Register(s)
		if err != nil {
			return err
		}

		s.Log.Info().Str("addr", l.Addr().String()).Msg("starting up RPC server")
		go srv.Accept(l)
	}

	return nil
}

// Ping to make sure it works and healthy
func (s *S) Ping(req bool, resp *string) error {
	*resp = "PONG from PASSPORT"
	return nil
}
