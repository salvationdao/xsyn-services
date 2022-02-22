package passport

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

// APIService for long running
type APIService struct {
	Addr string
	Log  *zap.SugaredLogger
}

// Run the API service
func (s *APIService) Run(ctx context.Context, controller http.Handler) error {
	s.Log.Infow("Starting API")

	server := &http.Server{
		Addr:    s.Addr,
		Handler: controller,
	}

	go func() {
		<-ctx.Done()
		s.Log.Info("Stopping API")
		err := server.Shutdown(ctx)
		if err != nil {
			s.Log.Error(err)
		}

	}()

	return server.ListenAndServe()
}
