package web

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"slices"

	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"

	"github.com/topi314/campfire-auth/internal/xrand"
)

type LoginVars struct {
	ClientID    string
	RedirectURI string
	Errs        []string
}

func (h *handler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()

	var errs []string
	clientID := query.Get("client_id")
	redirectURI := query.Get("redirect_uri")
	if clientID == "" {
		errs = append(errs, "Missing client_id")
	}
	if redirectURI == "" {
		errs = append(errs, "Missing redirect_uri")
	}

	if clientID != "" {
		client, err := h.DB.GetClient(ctx, clientID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				errs = append(errs, "Invalid client_id")
			} else {
				slog.ErrorContext(ctx, "Failed to get client", slog.String("client_id", clientID), slog.String("err", err.Error()))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		}
		if client != nil && !slices.Contains(client.RedirectURIs.V, redirectURI) {
			errs = append(errs, "Invalid redirect_uri")
		}
	}

	if err := h.Templates().ExecuteTemplate(w, "login.gohtml", LoginVars{
		ClientID:    clientID,
		RedirectURI: redirectURI,
		Errs:        errs,
	}); err != nil {
		slog.ErrorContext(ctx, "Failed to render login template", slog.String("err", err.Error()))
	}
}

type LoginCodeVars struct {
	Code         string
	CheckCode    string
	CampfireLink string
}

func (h *handler) LoginCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()

	clientID := query.Get("client_id")
	redirectURI := query.Get("redirect_uri")
	state := query.Get("state")
	if clientID == "" {
		http.Error(w, "Missing client_id", http.StatusBadRequest)
		return
	}
	if redirectURI == "" {
		http.Error(w, "Missing redirect_uri", http.StatusBadRequest)
		return
	}
	if state == "" {
		http.Error(w, "Missing state", http.StatusBadRequest)
		return
	}

	client, err := h.DB.GetClient(ctx, clientID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Invalid client_id", http.StatusBadRequest)
			return
		}
		slog.ErrorContext(ctx, "Failed to get client", slog.String("client_id", clientID), slog.String("err", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !slices.Contains(client.RedirectURIs.V, redirectURI) {
		http.Error(w, "Invalid redirect_uri", http.StatusBadRequest)
		return
	}

	code := xrand.RandCode()
	checkCode := xrand.RandCode()
	exchangeCode := xrand.RandCharCode()
	slog.InfoContext(ctx, "Generated login code", slog.String("client_id", clientID), slog.String("code", code), slog.String("check_code", checkCode), slog.String("exchange_code", exchangeCode))

	if err = h.DB.InsertLogin(ctx, clientID, code, checkCode, exchangeCode, redirectURI, state); err != nil {
		slog.ErrorContext(ctx, "Failed to insert login", slog.String("client_id", clientID), slog.String("code", code), slog.String("err", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err = h.Templates().ExecuteTemplate(w, "login_code.gohtml", LoginCodeVars{
		Code:         code,
		CheckCode:    checkCode,
		CampfireLink: getChannelLink(client.ClubID, client.ChannelID),
	}); err != nil {
		slog.ErrorContext(ctx, "Failed to render login code template", slog.String("err", err.Error()))
	}
}

func (h *handler) LoginQRCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	code := r.PathValue("code")

	qr, err := qrcode.New(fmt.Sprintf("%s/login/re/%s", h.Cfg.Server.PublicURL, code))
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create qrcode", slog.String("error", err.Error()))
		http.Error(w, "Failed to create qrcode", http.StatusInternalServerError)
		return
	}

	qrW := standard.NewWithWriter(&responseWriteCloser{w}, standard.WithLogoImage(h.Logo),
		standard.WithBgTransparent(),
		standard.WithBuiltinImageEncoder(standard.PNG_FORMAT),
		standard.WithLogoSafeZone(),
		standard.WithLogoSizeMultiplier(2),
	)

	defer func() {
		_ = qrW.Close()
	}()
	if err = qr.Save(qrW); err != nil {
		slog.ErrorContext(ctx, "Failed to save qrcode", slog.String("error", err.Error()))
	}
}

func (h *handler) LoginRe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	code := r.PathValue("code")

	login, err := h.DB.GetLoginByCode(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Invalid code", http.StatusBadRequest)
			return
		}
		slog.ErrorContext(ctx, "Failed to get login", slog.String("code", code), slog.String("err", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	link := getChannelLink(login.Client.ClubID, login.Client.ChannelID)
	http.Redirect(w, r, link, http.StatusFound)
}

func (h *handler) LoginCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()

	checkCode := query.Get("check_code")
	if checkCode == "" {
		http.Error(w, "Missing check_code", http.StatusBadRequest)
	}

	login, err := h.DB.GetLoginByCheckCode(ctx, checkCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if err = h.Templates().ExecuteTemplate(w, "login_expired.gohtml", nil); err != nil {
				slog.ErrorContext(ctx, "Failed to render login expired template", slog.String("err", err.Error()))
			}
			return
		}
		slog.ErrorContext(ctx, "Failed to get login", slog.String("check_code", checkCode), slog.String("err", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if login.User == nil {
		if err = h.Templates().ExecuteTemplate(w, "login_code.gohtml", LoginCodeVars{
			Code:         login.Code,
			CheckCode:    login.CheckCode,
			CampfireLink: getChannelLink(login.Client.ClubID, login.Client.ChannelID),
		}); err != nil {
			slog.ErrorContext(ctx, "Failed to render login code template", slog.String("err", err.Error()))
		}
		return
	}

	u, _ := url.Parse(login.RedirectURI)
	q := u.Query()
	q.Set("code", login.ExchangeCode)
	q.Set("state", login.State)
	u.RawQuery = q.Encode()

	w.Header().Set("HX-Redirect", u.String())
	w.WriteHeader(http.StatusOK)
}

func getChannelLink(clubID string, channelID string) string {
	v := url.Values{}
	v.Set("r", "clubs")
	v.Set("c", clubID)
	v.Set("ch", channelID)

	q := url.Values{}
	q.Set("af_dp", "campfire://")
	q.Set("af_force_deeplink", "true")
	q.Set("deep_link_sub1", base64.StdEncoding.EncodeToString([]byte(v.Encode())))

	u := url.URL{
		Scheme:   "https",
		Host:     "campfire.onelink.me",
		Path:     "eBr8",
		RawQuery: q.Encode(),
	}
	return u.String()
}

type responseWriteCloser struct {
	io.Writer
}

func (rwc *responseWriteCloser) Close() error {
	if closer, ok := rwc.Writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
