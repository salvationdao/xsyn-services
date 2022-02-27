package api

import (
	"context"
	"encoding/json"
	"math/big"
	"passport"
	"sync"
	"time"

	"github.com/ninja-syndicate/hub"
)

type SupremacySupPool struct {
	TotalSups     passport.BigInt
	TrickleAmount passport.BigInt
}

// StartSupremacySupPool cache the total sup pool in supremacy game
func (api *API) StartSupremacySupPool() {
	ssp := &SupremacySupPool{
		TotalSups:     passport.BigInt{Int: *big.NewInt(0)},
		TrickleAmount: passport.BigInt{Int: *big.NewInt(0)},
	}

	go func() {
		for fn := range api.supremacySupsPool {
			fn(ssp)
		}
	}()
}

// SupremacySupPoolSet set current sup pool detail
func (api *API) SupremacySupPoolSet(sups passport.BigInt) {
	api.supremacySupsPool <- func(ssp *SupremacySupPool) {
		// initialise sup pool value
		ssp.TotalSups = passport.BigInt{Int: *big.NewInt(0)}
		ssp.TotalSups.Add(&ssp.TotalSups.Int, &sups.Int)

		ssp.TrickleAmount = passport.BigInt{Int: *big.NewInt(0)}
		ssp.TrickleAmount.Add(&ssp.TrickleAmount.Int, &sups.Int)
		ssp.TrickleAmount.Div(&ssp.TrickleAmount.Int, big.NewInt(100))
	}
}

// SupremacySupPoolGetTrickleAmount return current trickle amount
func (api *API) SupremacySupPoolGetTrickleAmount() passport.BigInt {
	amount := passport.BigInt{}
	wg := sync.WaitGroup{}
	wg.Add(1)
	api.supremacySupsPool <- func(ssp *SupremacySupPool) {
		defer wg.Done()
		amount = ssp.TrickleAmount
	}
	wg.Wait()
	return amount
}

type ServerClientsList map[ServerClientName]map[*hub.Client]bool
type ServerClientsFunc func(serverClients ServerClientsList)

// ServerClientOnline adds a server client to the server client map
func (api *API) ServerClientOnline(gameName ServerClientName, hubc *hub.Client) {
	api.ServerClients(func(serverClients ServerClientsList) {
		_, ok := serverClients[gameName]
		if !ok {
			serverClients[gameName] = make(map[*hub.Client]bool)
		}
		serverClients[gameName][hubc] = true
	})
}

// ServerClientOffline removed a server hub client from the server client map
func (api *API) ServerClientOffline(hubc *hub.Client) {
	api.ServerClients(func(serverClients ServerClientsList) {
		for gameName, clientList := range serverClients {
			delete(clientList, hubc)
			if len(clientList) == 0 {
				delete(serverClients, gameName)
			}
		}
	})
}

// ServerClients accepts a function that loops over the server clients map
func (api *API) ServerClients(fn ServerClientsFunc) {
	api.serverClients <- func(serverClients ServerClientsList) {
		fn(serverClients)
	}
}

type ServerClientName string

const (
	SupremacyGameServer ServerClientName = "SUPREMACY:GAMESERVER"
	SupremacyGameClient ServerClientName = "SUPREMACY:GAMECLIENT"
)

type ServerClient struct {
	ServerName ServerClientName
	Client     *hub.Client
}

type ServerClientMessageAction string

const (
	Authed                     ServerClientMessageAction = "AUTHED"
	UserOnlineStatus           ServerClientMessageAction = "USER:ONLINE_STATUS"
	UserUpdated                ServerClientMessageAction = "USER:UPDATED"
	UserEnlistFaction          ServerClientMessageAction = "USER:ENLIST:FACTION"
	UserSupsUpdated            ServerClientMessageAction = "USER:SUPS:UPDATED"
	UserSupsMultiplierGet      ServerClientMessageAction = "USER:SUPS:MULTIPLIER:GET"
	UserStatGet                ServerClientMessageAction = "USER:STAT:GET"
	AssetUpdated               ServerClientMessageAction = "ASSET:UPDATED"
	AssetQueueJoin             ServerClientMessageAction = "ASSET:QUEUE:JOIN"
	AssetQueueLeave            ServerClientMessageAction = "ASSET:QUEUE:LEAVE"
	AssetInsurancePay          ServerClientMessageAction = "ASSET:INSURANCE:PAY"
	FactionStatGet             ServerClientMessageAction = "FACTION:STAT:GET"
	WarMachineQueuePositionGet ServerClientMessageAction = "WAR:MACHINE:QUEUE:POSITION:GET"
)

type ServerClientMessage struct {
	Key           ServerClientMessageAction `json:"key"`
	TransactionID string                    `json:"transactionID"`
	Payload       interface{}               `json:"payload,omitempty"`
}

func (api *API) SendToServerClient(ctx context.Context, name ServerClientName, msg *ServerClientMessage) {
	api.Log.Debug().Msgf("sending message to server clients: %s", name)
	api.serverClients <- func(servers ServerClientsList) {
		gameClientMap, ok := servers[name]
		if !ok {
			api.Log.Debug().Msgf("no server clients for %s", name)
		}

		for sc := range gameClientMap {
			payload, err := json.Marshal(msg)
			if err != nil {
				api.Log.Err(err).Msgf("error sending message to server client for: %s", name)
			}

			sc.Send(ctx, 3 * time.Second, payload)
		}
	}

}

func (api *API) SendToAllServerClient(ctx context.Context, msg *ServerClientMessage) {
	api.serverClients <- func(servers ServerClientsList) {
		for gameName, scm := range servers {
			for sc := range scm {
				payload, err := json.Marshal(msg)
				if err != nil {
					api.Log.Err(err).Msgf("error sending message to server client: %s", gameName)
				}
				sc.Send(ctx, 3 * time.Second, payload)
			}
		}
	}
}

func (api *API) HandleServerClients() {
	var serverClientsMap ServerClientsList = map[ServerClientName]map[*hub.Client]bool{}
	for {
		serverClientsFN := <-api.serverClients
		serverClientsFN(serverClientsMap)
	}
}
