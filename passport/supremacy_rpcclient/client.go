package supremacy_rpcclient

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net/rpc"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"xsyn-services/passport/passlog"

	"github.com/ninja-software/terror/v2"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// SupremacyXrpcClient is a basic RPC client with retry function and also support multiple addresses for increase through-put
type SupremacyXrpcClient struct {
	Addrs   []string      // list of rpc addresses available to use
	clients []*rpc.Client // holds rpc clients, same len/pos as the Addrs
	counter uint64        // counter for cycling address/clients
	mutex   sync.Mutex    // lock and unlocks clients slice editing
}

var SupremacyClient *SupremacyXrpcClient

func SetGlobalClient(c *SupremacyXrpcClient) {
	if SupremacyClient != nil {
		passlog.L.Fatal().Msg("rpc client already initialised")
	}
	SupremacyClient = c
}

// Call calls RPC server and retry, also initialise if it is the first time
func (c *SupremacyXrpcClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	span := tracer.StartSpan("rpc_client", tracer.ResourceName(serviceMethod))
	defer span.Finish()

	// used for the first time, initialise
	if c.clients == nil {
		passlog.L.Debug().Msg("comms.Call init first time")
		if len(c.Addrs) <= 0 {
			log.Fatal("no rpc address set")
		}

		c.mutex.Lock()
		for i := 0; i < len(c.Addrs); i++ {
			c.clients = append(c.clients, nil)
		}
		c.mutex.Unlock()
	}

	passlog.L.Trace().Str("fn", serviceMethod).Interface("args", args).Msg("rpc call")

	// count up, and use the next client/address
	atomic.AddUint64(&c.counter, 1)
	i := int(c.counter) % len(c.Addrs)
	client := c.clients[i]

	var err error = nil
	var retryCall uint
	for {
		if client == nil {
			// keep redialing until rpc server comes back online
			client, err = dial(5, c.Addrs[i])
			if err != nil {
				return err
			}
			c.mutex.Lock()
			c.clients[i] = client
			c.mutex.Unlock()
		}

		err = client.Call(serviceMethod, args, reply)
		if err == nil {
			// Successful RPC call, break the retry loop
			break
		}

		if !errors.Is(err, rpc.ErrShutdown) {
			if !strings.Contains(err.Error(), "asset is equipped to another object") {
				passlog.L.Error().Err(err).Msg("RPC call has failed.")
			}
			return err
		}

		passlog.L.Error().Err(err).Msg("RPC call has failed due to error.")

		// clean up before retry
		if client != nil {
			// close first
			client.Close()
		}
		client = nil

		retryCall++
		if retryCall > 6 {
			return terror.Error(fmt.Errorf("call %s retry exceeded 6 times", serviceMethod))
		}
	}

	return err
}

// dial is primitive rpc dialer, short and simple
// maxRetry -1 == unlimited
func dial(maxRetry int, addrAndPort string) (client *rpc.Client, err error) {
	retry := 0
	err = fmt.Errorf("x")
	sleepTime := time.Millisecond * 1000 // 1 second
	sleepTimeMax := time.Second * 10     // 10 seconds
	sleepTimeScale := 1.01               // power scale, takes 12 retries to reach max 10 sec sleep, which is total of ~40 seconds later

	for err != nil {
		client, err = rpc.Dial("tcp", addrAndPort)
		if err == nil {
			break
		}
		passlog.L.Debug().Err(err).Str("fn", "comms.dial").Msgf("err: dial fail, retrying... %d", retry)

		// increase timeout each time upto max
		time.Sleep(sleepTime)
		sleepTime = time.Duration(math.Pow(float64(sleepTime), sleepTimeScale))
		if sleepTime > sleepTimeMax {
			sleepTime = sleepTimeMax
		}

		retry++
		// unlimited retry
		if maxRetry < 0 {
			continue
		}

		// limited retry
		if retry > maxRetry {
			return nil, terror.Error(fmt.Errorf("rpc dial failed after %d retries", retry))
		}
	}

	return client, nil
}
