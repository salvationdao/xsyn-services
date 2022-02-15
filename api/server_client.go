package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"passport"
	"sync"
	"time"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-software/tickle"
	"github.com/ninja-syndicate/hub"
)

// InitialiseTreasuryFundTicker for every game server
func (api *API) InitialiseTreasuryFundTicker() {
	// set up treasury map tickle for supremacy game server
	tickle.MinDurationOverride = true
	api.treasuryTickerMap[SupremacyGameServer] = tickle.New(fmt.Sprintf("Treasury Ticker for %s", SupremacyGameServer), 5, func() (int, error) {
		fund := big.NewInt(0)
		fund, ok := fund.SetString("4000000000000000000", 10)
		if !ok {
			return http.StatusInternalServerError, terror.Error(fmt.Errorf("failed to convert 4000000000000000000 to big int"))
		}

		//treasuryTransfer := big.NewInt(0)
		//treasuryTransfer.Add(treasuryTransfer, fund)

		api.transaction <- &passport.NewTransaction{
			From:                 passport.XsynTreasuryUserID,
			To:                   passport.SupremacySupPoolUserID,
			Amount:               *fund,
			TransactionReference: passport.TransactionReference(fmt.Sprintf("treasury|ticker|%s", time.Now())),
		}

		return http.StatusOK, nil
	})
	api.treasuryTickerMap[SupremacyGameServer].DisableLogging = true
}

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
	amountChan := make(chan passport.BigInt)
	api.supremacySupsPool <- func(ssp *SupremacySupPool) {
		amountChan <- ssp.TrickleAmount
	}
	return <-amountChan
}

type ServerClientsList map[ServerClientName]map[*hub.Client]bool
type ServerClientsFunc func(serverClients ServerClientsList)

// ServerClientOnline adds a server client to the server client map
func (api *API) ServerClientOnline(gameName ServerClientName, hubc *hub.Client) {
	api.ServerClients(func(serverClients ServerClientsList) {
		_, ok := serverClients[gameName]
		if !ok {
			// start treasury ticker for current server client
			if tick, ok := api.treasuryTickerMap[gameName]; ok && (tick.NextTick == nil || tick.NextTick.Before(time.Now())) {
				tick.Start()
			}

			// set up sups pool user cache
			//if gameName == SupremacyGameServer {
			//	supsPoolUser, err := db.UserGet(context.Background(), api.Conn, passport.SupremacySupPoolUserID)
			//	if err != nil {
			//		api.Log.Err(err)
			//		return
			//	}
			//
			//	// initial total sups pool
			//	api.SupremacySupPoolSet(supsPoolUser.Sups)
			//}

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
				// end treasury ticker for current server client
				if tick, ok := api.treasuryTickerMap[gameName]; ok && tick.NextTick != nil {
					tick.Stop()
				}
				delete(serverClients, gameName)
			}
		}
	})
}

// ServerClients accepts a function that loops over the server clients map
func (api *API) ServerClients(fn ServerClientsFunc) {
	var wg sync.WaitGroup
	wg.Add(1)
	api.serverClients <- func(serverClients ServerClientsList) {
		fn(serverClients)
		wg.Done()
	}
	wg.Wait()
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

			go func(client *hub.Client) {
				err := client.Send(ctx, payload)
				if err != nil {
					api.Log.Err(err).Msgf("error sending")
				}
			}(sc)
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
				go func(client *hub.Client) {
					err := client.Send(ctx, payload)
					if err != nil {
						api.Log.Err(err).Msgf("error sending")
					}
				}(sc)
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
