package api

import (
	"fmt"
	"net/http"
	"passport"
	"passport/db"
	"passport/helpers"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
)

type FactionData struct {
	FactionID      passport.FactionID
	MemberCount    int64
	MVPUser        string
	TotalSpentSups int64
}

func (api *API) FactionGetData(w http.ResponseWriter, r *http.Request) (int, error) {
	fID, ok := r.URL.Query()["factionID"]
	if !ok {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("url query param not given"))
	}
	factionID := uuid.Must(uuid.FromString(fID[0]))
	factionData := &FactionData{FactionID: passport.FactionID(factionID)}
	memberCount, err := db.FactionGetRecruitNumber(r.Context(), api.Conn, passport.FactionID(factionID))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("url query param not given"))
	}

	factionSupVote, err := db.FactionSupsVotedGet(r.Context(), api.Conn, passport.FactionID(factionID))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("url query param not given"))
	}

	factionData.MemberCount = memberCount
	factionData.TotalSpentSups = factionSupVote.Int64()

	user, err := db.FactionMvpGet(r.Context(), api.Conn, passport.FactionID(factionID))
	if err != nil {
		factionData.MVPUser = "nil"
		return helpers.EncodeJSON(w, factionData)
	}
	factionData.MVPUser = user.Username
	return helpers.EncodeJSON(w, factionData)
}