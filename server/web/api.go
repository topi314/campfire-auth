package web

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

func (h *handler) ExchangeCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()

	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, "Missing basic auth", http.StatusUnauthorized)
		return
	}

	code := query.Get("code")
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

func (h *handler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if !h.checkClientAuth(w, r) {
		return
	}

	userID := r.PathValue("user_id")
	if userID == "" {
		http.Error(w, "Missing user_id", http.StatusBadRequest)
		return
	}

	user, err := h.Campfire.GetUserByID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get user by ID", slog.String("user_id", userID), slog.String("err", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(user); err != nil {
		slog.ErrorContext(ctx, "Failed to encode user", slog.String("err", err.Error()))
		return
	}
}

func (h *handler) SearchUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()
	if !h.checkClientAuth(w, r) {
		return
	}

	username := query.Get("username")
	if username == "" {
		http.Error(w, "Missing username", http.StatusBadRequest)
		return
	}

	users, err := h.Campfire.SearchUsers(ctx, username)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to search users", slog.String("username", username), slog.String("err", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(users); err != nil {
		slog.ErrorContext(ctx, "Failed to encode users", slog.String("err", err.Error()))
		return
	}
}

func (h *handler) checkClientAuth(w http.ResponseWriter, r *http.Request) bool {
	ctx := r.Context()
	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, "Missing basic auth", http.StatusUnauthorized)
		return false
	}

	_, err := h.DB.GetClientByIDSecret(ctx, username, password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Invalid client credentials", http.StatusUnauthorized)
			return false
		}
		slog.ErrorContext(ctx, "Failed to get client by ID and secret", slog.String("client_id", username), slog.String("err", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return false
	}

	return true
}
