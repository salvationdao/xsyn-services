package api

import (
	"net/http"

	DatadogTracer "github.com/ninja-syndicate/hub/ext/datadog"
	"github.com/rs/zerolog"

	"github.com/go-chi/chi/v5"
)

// CheckController holds connection data for handlers
type CheckController struct {
	Log *zerolog.Logger
}

func CheckRouter(log *zerolog.Logger) chi.Router {
	c := &CheckController{
		Log: log,
	}
	r := chi.NewRouter()
	r.Get("/", c.Check)

	return r
}

func (c *CheckController) Check(w http.ResponseWriter, r *http.Request) {
	err := check()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, wErr := w.Write([]byte(err.Error()))
		if wErr != nil {
			c.Log.Err(wErr).Msg("failed to send")
			DatadogTracer.HttpFinishSpan(r.Context(), http.StatusInternalServerError, wErr)
		} else {
			DatadogTracer.HttpFinishSpan(r.Context(), http.StatusInternalServerError, err)
		}
		return
	}
	_, err = w.Write([]byte("ok"))
	if err != nil {
		c.Log.Err(err).Msg("failed to send")
		DatadogTracer.HttpFinishSpan(r.Context(), http.StatusInternalServerError, err)
	}
	DatadogTracer.HttpFinishSpan(r.Context(), http.StatusOK, nil)
}
