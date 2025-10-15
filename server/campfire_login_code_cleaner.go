package server

import (
	"context"
	"log/slog"
	"time"
)

func (s *Server) loginCodeCleaner() {
	for {
		s.doLoginCodeClean()
		time.Sleep(10 * time.Second)
	}
}

func (s *Server) doLoginCodeClean() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.DB.DeleteExpiredLogins(ctx); err != nil {
		slog.ErrorContext(ctx, "Failed to delete expired logins", slog.String("err", err.Error()))
		return
	}

}
