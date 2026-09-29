package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"time"
)

// Server hosts the local web UI API and embedded SPA.
// It binds to 127.0.0.1 only and requires a per-boot token on every API call.
type Server struct {
	addr     string
	token    string
	sessions *sessionManager
	mux      *http.ServeMux
}

type Options struct {
	Port        int
	LockTimeout time.Duration
}

func New(opts Options) (*Server, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}
	s := &Server{
		addr:     fmt.Sprintf("127.0.0.1:%d", opts.Port),
		token:    hex.EncodeToString(tokenBytes),
		sessions: newSessionManager(opts.LockTimeout),
		mux:      http.NewServeMux(),
	}
	s.routes()
	return s, nil
}

// Token returns the bootstrap token (for tests / embedding).
func (s *Server) Token() string { return s.token }

// Addr returns the bound address (valid after Start).
func (s *Server) Addr() string { return s.addr }

func (s *Server) routes() {
	protected := func(h http.HandlerFunc) http.HandlerFunc {
		return s.hostCheck(s.originCheck(s.tokenCheck(h)))
	}

	s.mux.HandleFunc("GET /api/token", s.hostCheck(s.handleToken))
	s.mux.HandleFunc("GET /api/session", protected(s.handleSession))
	s.mux.HandleFunc("POST /api/unlock", protected(s.handleUnlock))
	s.mux.HandleFunc("POST /api/lock", protected(s.handleLock))
	s.mux.HandleFunc("GET /api/vault", protected(s.handleVaultList))
	s.mux.HandleFunc("POST /api/vault", protected(s.handleVaultAdd))
	s.mux.HandleFunc("PUT /api/vault/{id}", protected(s.handleVaultEdit))
	s.mux.HandleFunc("DELETE /api/vault/{id}", protected(s.handleVaultDelete))
	s.mux.HandleFunc("POST /api/vault/reorder", protected(s.handleVaultReorder))
	s.mux.HandleFunc("POST /api/generate", protected(s.handleGenerate))
	s.mux.HandleFunc("POST /api/change-master", protected(s.handleChangeMaster))
	s.mux.HandleFunc("GET /api/config", protected(s.handleConfigGet))
	s.mux.HandleFunc("POST /api/config", protected(s.handleConfigSet))
	s.mux.HandleFunc("GET /", s.hostCheck(s.handleStatic))
}

// Start binds 127.0.0.1 and starts the auto-lock janitor.
func (s *Server) Start() (net.Listener, error) {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return nil, err
	}
	s.addr = ln.Addr().String()
	s.sessions.startJanitor()
	return ln, nil
}

// Serve blocks serving requests on ln until the listener closes.
func (s *Server) Serve(ln net.Listener) error {
	return http.Serve(ln, s.mux)
}

// Handler exposes the mux (tests).
func (s *Server) Handler() http.Handler { return s.mux }

func (s *Server) Shutdown() {
	s.sessions.stopJanitor()
	s.sessions.clear()
}

// --- middleware ---

func (s *Server) hostCheck(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if host != s.addr && host != "localhost:"+portOf(s.addr) && host != "127.0.0.1:"+portOf(s.addr) {
			// Allow plain "localhost" when addr has no port ambiguity.
			if host != "localhost" {
				writeError(w, http.StatusForbidden, "forbidden host")
				return
			}
		}
		next(w, r)
	}
}

func (s *Server) originCheck(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" &&
			origin != "http://"+s.addr &&
			origin != "http://localhost:"+portOf(s.addr) &&
			origin != "http://127.0.0.1:"+portOf(s.addr) {
			writeError(w, http.StatusForbidden, "forbidden origin")
			return
		}
		next(w, r)
	}
}

func (s *Server) tokenCheck(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Lockbox-Token") != s.token {
			writeError(w, http.StatusForbidden, "invalid token")
			return
		}
		next(w, r)
	}
}

func portOf(addr string) string {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return ""
	}
	return port
}

// currentConfig reloads config per request so CLI-side `config set` is picked up live.
func (s *Server) currentConfig() (*configView, error) {
	return loadConfig()
}

type configView struct {
	VaultPath        string
	KDFIterations    int
	DefaultGenLength int
	ConfigPath       string
}
