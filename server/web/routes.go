package web

import (
	"log/slog"
	"net/http"

	"github.com/topi314/campfire-auth/internal/middlewares"
	"github.com/topi314/campfire-auth/server"
)

type handler struct {
	*server.Server
}

func Routes(srv *server.Server) http.Handler {
	h := &handler{
		Server: srv,
	}

	fileserver := srv.Reloader.CacheMiddleware(http.FileServer(srv.StaticFS))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", h.Index)

	mux.HandleFunc("GET /admin", h.Admin)
	mux.HandleFunc("POST /admin/tokens", h.AdminTokens)
	mux.HandleFunc("POST /admin/clients", h.AdminClients)

	mux.HandleFunc("GET /login", h.Login)
	mux.HandleFunc("GET /login/code", h.LoginCode)
	mux.Handle("GET /login/code/{code}", middlewares.Cache(http.HandlerFunc(h.LoginQRCode)))
	mux.HandleFunc("GET /login/re/{code}", h.LoginRe)
	mux.HandleFunc("GET /login/check", h.LoginCheck)

	mux.HandleFunc("GET /api/exchange", h.ExchangeCode)
	mux.HandleFunc("GET /api/users/search", h.SearchUser)
	mux.HandleFunc("GET /api/users/{user_id}", h.GetUser)

	mux.HandleFunc("GET /api/docs", h.APIDocs)

	mux.Handle("/static/", fileserver)

	if srv.Cfg.Dev {
		mux.Handle(server.ReloadRoute, srv.Reloader.Handler())
	}

	mux.HandleFunc("/", h.NotFound)

	return middlewares.CleanPath(mux)
}

func (h *handler) api() http.Handler {
	mux := http.NewServeMux()

	return http.StripPrefix("/api", mux)
}

func (h *handler) NotFound(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := h.Templates().ExecuteTemplate(w, "not_found.gohtml", nil); err != nil {
		slog.ErrorContext(ctx, "Failed to render not found template", slog.String("error", err.Error()))
		return
	}
}
