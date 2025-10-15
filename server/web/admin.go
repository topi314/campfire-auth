package web

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/topi314/campfire-auth/internal/xrand"
	"github.com/topi314/campfire-auth/server/database"
)

type AdminVars struct {
	Tokens       []Token
	Clients      []Client
	Password     string
	TokenErrors  []string
	ClientErrors []string
}

func newToken(token database.CampfireToken) Token {
	return Token{
		ID:        token.ID,
		Token:     token.Token,
		ExpiresAt: token.ExpiresAt,
		Email:     token.Email,
	}
}

type Token struct {
	ID        int
	Token     string
	ExpiresAt time.Time
	Email     string
}

func newClient(client database.Client) Client {
	return Client{
		Name:         client.Name,
		ID:           client.ID,
		Secret:       client.Secret,
		RedirectURIs: strings.Join(client.RedirectURIs.V, ", "),
		CreatedAt:    client.CreatedAt,
	}
}

type Client struct {
	Name         string
	ID           string
	Secret       string
	RedirectURIs string
	CreatedAt    time.Time
}

func (h *handler) Admin(w http.ResponseWriter, r *http.Request) {
	h.renderAdmin(w, r, nil, nil)
}

func (h *handler) renderAdmin(w http.ResponseWriter, r *http.Request, tokenErrors []string, clientErrors []string) {
	ctx := r.Context()

	if !h.checkIsAdmin(w, r) {
		return
	}

	tokens, err := h.DB.GetCampfireTokens(ctx)
	if err != nil {
		http.Error(w, "Failed to fetch tokens: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var tokenList []Token
	for _, t := range tokens {
		tokenList = append(tokenList, newToken(t))
	}

	clients, err := h.DB.GetClients(ctx)
	if err != nil {
		http.Error(w, "Failed to fetch clients: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var clientList []Client
	for _, c := range clients {
		clientList = append(clientList, newClient(c))
	}

	if err = h.Templates().ExecuteTemplate(w, "admin.gohtml", AdminVars{
		Tokens:       tokenList,
		Clients:      clientList,
		Password:     h.Cfg.Server.AdminPassword,
		TokenErrors:  tokenErrors,
		ClientErrors: clientErrors,
	}); err != nil {
		slog.ErrorContext(ctx, "Failed to render tracker template", slog.Any("err", err))
	}
}

func (h *handler) AdminTokens(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if !h.checkIsAdmin(w, r) {
		return
	}

	token := r.FormValue("token")
	if token == "" {
		h.renderAdmin(w, r, []string{"Token cannot be empty"}, nil)
		return
	}

	campfireToken, err := parseToken(token)
	if err != nil {
		h.renderAdmin(w, r, []string{"Invalid token: " + err.Error()}, nil)
		return
	}

	if err = h.DB.InsertCampfireToken(ctx, *campfireToken); err != nil {
		h.renderAdmin(w, r, []string{"Failed to insert token: " + err.Error()}, nil)
		return
	}

	h.redirectAdmin(w, r)
}

func (h *handler) AdminClients(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if !h.checkIsAdmin(w, r) {
		return
	}

	name := r.FormValue("name")
	redirectURIs := r.FormValue("redirect_uris")
	if name == "" {
		h.renderAdmin(w, r, nil, []string{"Name cannot be empty"})
		return
	}
	if redirectURIs == "" {
		h.renderAdmin(w, r, nil, []string{"Redirect URIs cannot be empty"})
		return
	}

	var redirects []string
	for _, uri := range strings.Split(redirectURIs, ",") {
		uri = strings.TrimSpace(uri)
		if uri != "" {
			redirects = append(redirects, uri)
		}
	}

	clientID := xrand.RandCharCode()
	clientSecret := xrand.RandCharCode()

	if err := h.DB.InsertClient(ctx, name, clientID, clientSecret, redirects); err != nil {
		h.renderAdmin(w, r, nil, []string{"Failed to insert client: " + err.Error()})
		return
	}

	h.redirectAdmin(w, r)
}

func (h *handler) redirectAdmin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, fmt.Sprintf("/admin?password=%s", h.Cfg.Server.AdminPassword), http.StatusSeeOther)

}

func parseToken(token string) (*database.CampfireToken, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	tokenData, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid token data: %w", err)
	}

	var t jwtToken
	if err = json.Unmarshal(tokenData, &t); err != nil {
		return nil, fmt.Errorf("invalid token json: %w", err)
	}

	return &database.CampfireToken{
		Token:     token,
		ExpiresAt: time.Unix(t.Exp, 0),
		Email:     t.Email,
	}, nil
}

type jwtToken struct {
	Email string `json:"email"`
	Exp   int64  `json:"exp"`
	Iat   int64  `json:"iat"`
	Iss   string `json:"iss"`
	Sub   string `json:"sub"`
}

func (h *handler) checkIsAdmin(w http.ResponseWriter, r *http.Request) bool {
	query := r.URL.Query()
	password := query.Get("password")
	if password != h.Cfg.Server.AdminPassword {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}

	return true
}
