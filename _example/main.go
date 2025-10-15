package main

import (
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
)

var (
	//go:embed templates/*.gohtml
	templates embed.FS

	clientID     = "jde623lp0o0p3pr2"
	clientSecret = "9s0u6z1j5f28bscs"
	clubID       = "b632fc8e-0b41-49de-ade2-21b0cd81db69"
	channelID    = "aa67cc66-23fd-476b-a9e3-70782de95457"
	redirectURI  = "http://localhost:8080/callback"
	authURL      = "http://localhost:8086/login"
	tokenURL     = "http://localhost:8086/api/exchange"
)

func main() {
	t := template.Must(template.New("templates").
		ParseFS(templates, "templates/*.gohtml"))

	s := server{
		t: t,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/{$}", s.index)
	mux.HandleFunc("/callback", s.callback)

	s.s = &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		if err := s.s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server error", slog.Any("err", err))
		}
	}()

	slog.Info("Server started at :8080")
	si := make(chan os.Signal, 1)
	signal.Notify(si, syscall.SIGTERM, syscall.SIGINT)
	<-si
}

type server struct {
	t *template.Template
	s *http.Server
}

func (s *server) index(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(authURL)
	if err != nil {
		http.Error(w, "Invalid auth URL", http.StatusInternalServerError)
		return
	}

	q := u.Query()
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("club_id", clubID)
	q.Set("channel_id", channelID)
	// In production, use a proper random state and validate it in the callback
	q.Set("state", "some-random-state")
	u.RawQuery = q.Encode()

	if err = s.t.ExecuteTemplate(w, "index.gohtml", map[string]any{
		"LoginURL": u.String(),
	}); err != nil {
		slog.Error("Template execute error", slog.Any("err", err))
		return
	}
}

func (s *server) callback(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	code := query.Get("code")
	state := query.Get("state")
	if code == "" {
		http.Error(w, "Missing code", http.StatusBadRequest)
		return
	}
	if state != "some-random-state" {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	m, err := exchangeCode(code)
	if err != nil {
		http.Error(w, "Failed to exchange code: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonUser, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		http.Error(w, "Failed to encode member: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err = s.t.ExecuteTemplate(w, "callback.gohtml", map[string]any{
		"Code":     code,
		"User":     m,
		"JSONUser": string(jsonUser),
	}); err != nil {
		slog.Error("Template execute error", slog.Any("err", err))
		return
	}
}

type User struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	AvatarURL   string `json:"avatarUrl"`
	Badges      []struct {
		Alias     string `json:"alias"`
		BadgeType string `json:"badgeType"`
	} `json:"badges"`
	GameProfiles []struct {
		ID                    string `json:"id"`
		Game                  string `json:"game"`
		Codename              string `json:"codename"`
		DisplayName           string `json:"displayName"`
		Level                 int    `json:"level"`
		Faction               string `json:"faction"`
		FactionColor          string `json:"factionColor"`
		Visibility            string `json:"visibility"`
		LastPlayedTimestampMs int64  `json:"lastPlayedTimestampMs"`
	} `json:"gameProfiles"`
}

func exchangeCode(code string) (*User, error) {
	rq, err := http.NewRequest(http.MethodGet, tokenURL+"?code="+url.QueryEscape(code), nil)
	if err != nil {
		return nil, err
	}

	rq.SetBasicAuth(clientID, clientSecret)

	rs, err := http.DefaultClient.Do(rq)
	if err != nil {
		return nil, err
	}

	defer rs.Body.Close()

	if rs.StatusCode != http.StatusOK {
		return nil, errors.New("invalid response status: " + rs.Status)
	}

	var m User
	if err = json.NewDecoder(rs.Body).Decode(&m); err != nil {
		return nil, err
	}

	return &m, nil
}
