package comms

import (
	"fmt"
	"net"
	"net/rpc"
	"strconv"
	"sync"
	"time"
	"xsyn-services/passport/api"
	"xsyn-services/types"

	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
	"github.com/sasha-s/go-deadlock"
	"github.com/shopspring/decimal"
)

type TickerPoolCache struct {
	deadlock.Mutex
	TricklingAmountMap map[string]decimal.Decimal
}

type S struct {
	StartPort    int
	EndPort      int
	UserCacheMap *api.Transactor
	MessageBus   *messagebus.MessageBus
	SMS          types.SMS
	//Txs          *api.Transactions
	Log                 *zerolog.Logger
	DistLock            deadlock.Mutex
	TickerPoolCache     *TickerPoolCache // user for sup trickle process
	HubSessionIDMap     *sync.Map
	TokenExpirationDays int
	TokenEncryptionKey  []byte
}

func NewServer(
	userCacheMap *api.Transactor,
	//txs *api.Transactions,
	log *zerolog.Logger,
	cm *sync.Map,
	sms types.SMS,
	config *types.Config,
) *S {
	result := &S{
		UserCacheMap: userCacheMap,
		//Txs:          txs,
		Log:                 log,
		TokenExpirationDays: config.TokenExpirationDays,
		TokenEncryptionKey:  []byte(config.EncryptTokensKey),
		TickerPoolCache: &TickerPoolCache{
			deadlock.Mutex{},
			make(map[string]decimal.Decimal),
		},
		HubSessionIDMap: cm,
		SMS:             sms,
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

func (s *S) Start(startPort, numPorts int) error {
	listenPorts := make([]string, numPorts)
	for i := 0; i < numPorts; i++ {
		listenPorts[i] = strconv.Itoa(startPort + i)
	}

	listeners, err := s.listen(listenPorts...)
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
