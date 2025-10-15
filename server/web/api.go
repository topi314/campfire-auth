package web

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
)

func (h *handler) ExchangeCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, "Missing basic auth", http.StatusUnauthorized)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code", http.StatusBadRequest)
		return
	}

	login, err := h.DB.DeleteLoginByClientIDSecretExchangeCode(ctx, username, password, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Invalid code", http.StatusBadRequest)
			return
		}
		slog.ErrorContext(ctx, "Failed to delete login by exchange code", slog.String("code", code), slog.String("err", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if _, err = w.Write(*login.User); err != nil {
		slog.ErrorContext(ctx, "Failed to write login user", slog.String("err", err.Error()))
		return
	}
}
