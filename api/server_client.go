package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"passport"

	"github.com/ninja-software/terror/v2"
)

// GameserverRequest set gameserver webhook request
func (api *API) GameserverRequest(method string, endpoint string, data interface{}, dist interface{}) error {
	jd, err := json.Marshal(data)
	if err != nil {
		return terror.Error(err, "failed to marshal data into json struct")
	}

	url := fmt.Sprintf("%s/api/%s/Supremacy_game%s", api.GameserverHostUrl, passport.SupremacyGameUserID, endpoint)

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jd))
	if err != nil {
		return terror.Error(err)
	}

	req.Header.Add("Passport-Authorization", api.WebhookToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return terror.Error(err)
	}

	if dist != nil {
		err = json.NewDecoder(resp.Body).Decode(&dist)
		if err != nil {
			return terror.Error(err, "failed to decode response")
		}
	}

	return nil
}

// type ServerClientsList map[ServerClientName]map[*hub.Client]bool
// type ServerClientsFunc func(serverClients ServerClientsList)

// // ServerClientOnline adds a server client to the server client map
// func (api *API) ServerClientOnline(gameName ServerClientName, hubc *hub.Client) {
// 	api.ServerClients(func(serverClients ServerClientsList) {
// 		_, ok := serverClients[gameName]
// 		if !ok {
// 			serverClients[gameName] = make(map[*hub.Client]bool)
// 		}
// 		serverClients[gameName][hubc] = true
// 	})
// }

// // ServerClientOffline removed a server hub client from the server client map
// func (api *API) ServerClientOffline(hubc *hub.Client) {
// 	api.ServerClients(func(serverClients ServerClientsList) {
// 		for gameName, clientList := range serverClients {
// 			delete(clientList, hubc)
// 			if len(clientList) == 0 {
// 				delete(serverClients, gameName)
// 			}
// 		}
// 	})
// }

// // ServerClients accepts a function that loops over the server clients map
// func (api *API) ServerClients(fn ServerClientsFunc) {
// 	api.serverClients <- func(serverClients ServerClientsList) {
// 		fn(serverClients)
// 	}
// }

// type ServerClientName string

// const (
// 	SupremacyGameServer ServerClientName = "SUPREMACY:GAMESERVER"
// 	SupremacyGameClient ServerClientName = "SUPREMACY:GAMECLIENT"
// )

// type ServerClient struct {
// 	ServerName ServerClientName
// 	Client     *hub.Client
// }

// type ServerClientMessageAction string

// const (
// 	UserOnlineStatus           ServerClientMessageAction = "USER:ONLINE_STATUS"
// 	UserUpdated                ServerClientMessageAction = "USER:UPDATED"
// 	UserEnlistFaction          ServerClientMessageAction = "USER:ENLIST:FACTION"
// 	UserSupsUpdated            ServerClientMessageAction = "USER:SUPS:UPDATED"
// 	UserSupsMultiplierGet      ServerClientMessageAction = "USER:SUPS:MULTIPLIER:GET"
// 	UserStatGet                ServerClientMessageAction = "USER:STAT:GET"
// 	AssetUpdated               ServerClientMessageAction = "ASSET:UPDATED"
// 	AssetQueueJoin             ServerClientMessageAction = "ASSET:QUEUE:JOIN"
// 	AssetQueueLeave            ServerClientMessageAction = "ASSET:QUEUE:LEAVE"
// 	AssetInsurancePay          ServerClientMessageAction = "ASSET:INSURANCE:PAY"
// 	FactionStatGet             ServerClientMessageAction = "FACTION:STAT:GET"
// 	WarMachineQueuePositionGet ServerClientMessageAction = "WAR:MACHINE:QUEUE:POSITION:GET"
// )

// type ServerClientMessage struct {
// 	Key           ServerClientMessageAction `json:"key"`
// 	TransactionID string                    `json:"transactionID"`
// 	Payload       interface{}               `json:"payload,omitempty"`
// }

// func (api *API) SendToServerClient(ctx context.Context, name ServerClientName, msg *ServerClientMessage) {
// 	api.Log.Debug().Msgf("sending message to server clients: %s", name)
// 	api.serverClients <- func(servers ServerClientsList) {
// 		gameClientMap, ok := servers[name]
// 		if !ok {
// 			api.Log.Debug().Msgf("no server clients for %s", name)
// 		}

// 		for sc := range gameClientMap {
// 			payload, err := json.Marshal(msg)
// 			if err != nil {
// 				api.Log.Err(err).Msgf("error sending message to server client for: %s", name)
// 			}

// 			go sc.Send(payload)
// 		}
// 	}

// }

// func (api *API) SendToAllServerClient(ctx context.Context, msg *ServerClientMessage) {
// 	api.serverClients <- func(servers ServerClientsList) {
// 		for gameName, scm := range servers {
// 			for sc := range scm {
// 				payload, err := json.Marshal(msg)
// 				if err != nil {
// 					api.Log.Err(err).Msgf("error sending message to server client: %s", gameName)
// 				}
// 				go sc.Send(payload)
// 			}
// 		}
// 	}
// }

// func (api *API) HandleServerClients() {
// 	var serverClientsMap ServerClientsList = map[ServerClientName]map[*hub.Client]bool{}
// 	for {
// 		serverClientsFN := <-api.serverClients
// 		serverClientsFN(serverClientsMap)
// 	}
// }
