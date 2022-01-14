package api

import (
	"encoding/json"
	"sync"

	"github.com/ninja-software/hub/v2"
)

type ServerClientsList map[ServerClientName]map[*hub.Client]bool
type ServerClientsFunc func(serverClients ServerClientsList)

// ServerClientOnline adds a server client to the server client map
func (api *API) ServerClientOnline(name ServerClientName, hubc *hub.Client) {
	api.ServerClients(func(serverClients ServerClientsList) {
		_, ok := serverClients[name]
		if !ok {
			serverClients[name] = make(map[*hub.Client]bool)
		}
		serverClients[name][hubc] = true
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
	Authed           ServerClientMessageAction = "AUTHED"
	UserOnlineStatus ServerClientMessageAction = "USER:ONLINE_STATUS"
	UserUpdated      ServerClientMessageAction = "USER:UPDATED"
	UserSupsUpdated  ServerClientMessageAction = "USER:SUPS:UPDATED"
)

type ServerClientMessage struct {
	Key           ServerClientMessageAction `json:"key"`
	TransactionID string                    `json:"transactionId"`
	Payload       interface{}               `json:"payload,omitempty"`
}

func (api *API) SendToServerClient(name ServerClientName, msg *ServerClientMessage) {
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
			go sc.Send(payload)
		}
	}
}

func (api *API) SendToAllServerClient(msg *ServerClientMessage) {
	api.Log.Debug().Msgf("sending message to all server clients")
	api.serverClients <- func(servers ServerClientsList) {
		for gameName, scm := range servers {
			for sc := range scm {
				payload, err := json.Marshal(msg)
				if err != nil {
					api.Log.Err(err).Msgf("error sending message to server client: %s", gameName)
				}
				go sc.Send(payload)
			}

		}
	}
}

func (api *API) HandleServerClients() {
	var serverClientsMap ServerClientsList = map[ServerClientName]map[*hub.Client]bool{}
	for {
		select {
		case serverClientsFN := <-api.serverClients:
			serverClientsFN(serverClientsMap)
		}
	}
}
