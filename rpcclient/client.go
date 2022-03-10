package rpcclient

import (
	"fmt"
	"log"
	"net/rpc"
	"passport/passlog"
	"sync"

	"github.com/ninja-software/terror/v2"
	"go.uber.org/atomic"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// XrpcClient is a basic RPC client with retry function and also support multiple addresses for increase through-put
type XrpcClient struct {
	Addrs   []string       // list of rpc addresses available to use
	clients []*rpc.Client  // holds rpc clients, same len/pos as the Addrs
	counter *atomic.Uint64 // counter for cycling address/clients
	mutex   *sync.Mutex    // lock and unlocks clients slice editing
}

var Client *XrpcClient

func SetGlobalClient(c *XrpcClient) {
	if Client != nil {
		passlog.L.Fatal().Msg("rpc client already initialised")
	}
	Client = c
}

// GoCall consider deprecate this function
// func (c *XrpcClient) GoCall(serviceMethod string, args interface{}, reply interface{}, callback func(error)) {
// 	go func() {
// 		err := c.Call(serviceMethod, args, reply)
// 		if callback != nil {
// 			callback(err)
// 		}
// 	}()
// }

// Call calls RPC server and retry, also initialise if it is the first time
func (c *XrpcClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	span := tracer.StartSpan("rpc.Call", tracer.ResourceName(serviceMethod))
	defer span.Finish()

	// used for the first time, initialise
	if c == nil {
		passlog.L.Debug().Msg("comms.Call init first time")
		if len(c.Addrs) <= 0 {
			log.Fatal("no rpc address set")
		}
		c.mutex.Lock()
		for i := 0; i < len(c.Addrs); i++ {
			c.clients = append(c.clients, &rpc.Client{})
		}
		c.mutex.Unlock()
	}

	passlog.L.Debug().Str("fn", serviceMethod).Interface("args", args).Msg("rpc call")

	// count up, and use the next client/address
	c.counter.Add(1)
	counter := c.counter.Load()
	i := int(counter) % len(c.Addrs)
	client := c.clients[i]

	var err error
	var retryCall uint
	for {
		if client == nil {
			// keep redialing until rpc server comes back online
			client, err = dial(-1, c.Addrs[i])
			if err != nil {
				return terror.Error(err)
			}
			c.mutex.Lock()
			c.clients[i] = client
			c.mutex.Unlock()
		}

		err = client.Call(serviceMethod, args, reply)
		if err == nil {
			// done
			break
		}

		// clean up before retry
		if client != nil {
			// close first
			client.Close()
		}
		client = nil

		retryCall++
		if retryCall > 6 {
			return terror.Error(fmt.Errorf("call retry exceeded 6 times"))
		}
	}

	return nil
}

// dial is primitive rpc dialer, short and simple
// maxRetry -1 == unlimited
func dial(maxRetry int, addrAndPort string) (client *rpc.Client, err error) {
	retry := 0
	err = fmt.Errorf("x")

	for err != nil {
		// rpc have own timeout probably 1~1.4 sec?
		client, err = rpc.Dial("tcp", addrAndPort)
		if err == nil {
			break
		}
		passlog.L.Debug().Err(err).Str("fn", "comms.dial").Msgf("err: dial fail, retrying... %d", retry)

		// unlimited retry
		if maxRetry < 0 {
			continue
		}

		retry++
		// limited retry
		if retry > maxRetry {
			return nil, terror.Error(fmt.Errorf("rpc dial failed after %d retries", maxRetry))
		}
	}

	return client, nil
}
