package rpcclient

import (
	"errors"
	"fmt"
	"net/rpc"
	"passport/passlog"
	"time"

	"github.com/jpillora/backoff"
	"github.com/ninja-software/terror/v2"
	"go.uber.org/atomic"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type C struct {
	addrs   []string
	clients []*rpc.Client
	inc     *atomic.Int32
}

var Client *C

func NewClient(addrs ...string) (*C, error) {
	clients, err := connect(addrs...)
	if err != nil {
		return nil, err
	}
	c := &C{addrs, clients, atomic.NewInt32(0)}
	return c, nil
}

func SetGlobalClient(c *C) {
	if Client != nil {
		passlog.L.Fatal().Msg("rpc client already initialised")
	}
	Client = c
}

func connect(addrs ...string) ([]*rpc.Client, error) {
	b := &backoff.Backoff{
		Min:    1 * time.Second,
		Max:    10 * time.Second,
		Factor: 2,
	}
	attempts := 0
	var clients []*rpc.Client

	for {
		attempts++
		passlog.L.Info().Int("attempt", attempts).Msg("connecting to gameserver")
		clients = []*rpc.Client{}
		for _, addr := range addrs {
			passlog.L.Info().Str("addr", addr).Msg("registering RPC client")
			client, err := rpc.Dial("tcp", addr)
			if err != nil {
				passlog.L.Err(err).Str("addr", addr).Msg("registering RPC client")
				time.Sleep(b.Duration())
				continue
			}
			clients = append(clients, client)
		}

		break
	}
	return clients, nil
}

func (c *C) GoCall(serviceMethod string, args interface{}, reply interface{}, callback func(error)) {
	go func() {
		err := c.Call(serviceMethod, args, reply)
		if callback != nil {
			callback(err)
		}
	}()
}

func (c *C) Call(serviceMethod string, args interface{}, reply interface{}) error {
	if c == nil || c.clients == nil || len(c.clients) <= 0 {
		return terror.Error(fmt.Errorf("rpc client not ready"), "The purchase system is currently not available. Please try again later.")
	}
	defer passlog.L.Debug().Str("fn", serviceMethod).Interface("args", args).Msg("rpc call")
	span := tracer.StartSpan("rpc.call", tracer.ResourceName(serviceMethod))
	defer span.Finish()
	c.inc.Add(1)
	i := c.inc.Load()
	if i >= int32(len(c.clients)-1) {
		c.inc.Store(0)
		i = 0
	}
	if len(c.clients) < int(i) {
		return fmt.Errorf("index out of range len = %d, index = %d", len(c.clients), int(i))
	}
	client := c.clients[i]
	err := client.Call(serviceMethod, args, reply)
	if err != nil && errors.Is(err, rpc.ErrShutdown) {
		newClients, err := connect(c.addrs...)
		if err != nil {
			time.Sleep(5 * time.Second)
			return c.Call(serviceMethod, args, reply)
		}
		c.clients = newClients
		return c.Call(serviceMethod, args, reply)
	}
	if err != nil {
		return err
	}
	return nil
}
