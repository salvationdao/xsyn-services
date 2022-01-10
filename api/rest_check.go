package api

import (
	"net/http"
	"passport/db"

	"github.com/go-chi/chi/v5"
)

// CheckController holds connection data for handlers
type CheckController struct {
	Conn db.Conn
}

func CheckRouter(conn db.Conn) chi.Router {
	c := &CheckController{
		Conn: conn,
	}
	r := chi.NewRouter()
	r.Get("/", c.Check)

	return r
}

func (c *CheckController) Check(w http.ResponseWriter, r *http.Request) {
	err := check(r.Context(), c.Conn)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte("ok"))
}
