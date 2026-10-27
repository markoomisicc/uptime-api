package api

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/markoomisicc/uptime-api/internal/config"
	"github.com/markoomisicc/uptime-api/internal/store"
)

type Server struct {
	cfg     config.Config
	pool    *pgxpool.Pool
	queries *store.Queries
}

func NewServer(cfg config.Config, pool *pgxpool.Pool) *Server {
	return &Server{
		cfg:     cfg,
		pool:    pool,
		queries: store.New(pool),
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Uptime API"))
	})
	mux.HandleFunc("GET /health", s.handleHealth)

	mux.HandleFunc("POST /auth/register", s.handleRegister)
	mux.HandleFunc("POST /auth/login", s.handleLogin)
	mux.HandleFunc("POST /auth/logout", s.handleLogout)
	mux.HandleFunc("GET /auth/me", s.requireAuth(s.handleMe))

	return mux
}
