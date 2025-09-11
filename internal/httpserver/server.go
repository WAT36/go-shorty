package httpserver

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/WAT36/shorty/internal/store"
)

type Server struct {
	http.Server
	store *store.FileStore
	mux   *http.ServeMux
}

func New(addr string, s *store.FileStore) *Server {
	mux := http.NewServeMux()
	srv := &Server{
		Server: http.Server{
			Addr:    addr,
			Handler: mux,
		},
		store: s,
		mux:   mux,
	}
	srv.routes()
	return srv
}

func (s *Server) routes() {
	// API
	s.mux.HandleFunc("/api/shorten", s.handleShorten)
	s.mux.HandleFunc("/api/list", s.handleList)
	s.mux.HandleFunc("/api/", s.handleDelete) // /api/{code}

	// ルート配下のすべてを 1 本で振り分け
	s.mux.HandleFunc("/", s.handleRoot)
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	// /api/ はここでは扱わない（上で登録したAPIハンドラが処理）
	if strings.HasPrefix(r.URL.Path, "/api/") {
		http.NotFound(w, r)
		return
	}

	// ルート（/）は index.html
	if r.URL.Path == "/" {
		tplPath := filepath.Clean("web/index.html")
		tpl, err := template.ParseFiles(tplPath)
		if err != nil {
			http.Error(w, "template error", http.StatusInternalServerError)
			return
		}
		_ = tpl.Execute(w, nil)
		return
	}

	// それ以外は /{code} とみなしてリダイレクト
	code := strings.TrimPrefix(r.URL.Path, "/")
	m, ok := s.store.Get(code)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if err := s.store.Increment(code); err != nil {
		log.Printf("increment error for %s: %v", code, err)
	}
	http.Redirect(w, r, m.URL, http.StatusFound)
}

type shortenReq struct {
	URL    string `json:"url"`
	Custom string `json:"custom,omitempty"`
}

type shortenResp struct {
	Code string `json:"code"`
	URL  string `json:"url"`
}

func (s *Server) handleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req shortenReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	code, err := s.store.Create(req.URL, req.Custom)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp := shortenResp{Code: code, URL: req.URL}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	list := s.store.List()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	// /api/{code}
	if r.Method != http.MethodDelete {
		http.NotFound(w, r)
		return
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/"), "/")
	if len(parts) < 1 || parts[0] == "" {
		http.Error(w, "code required", http.StatusBadRequest)
		return
	}
	code := parts[0]
	if err := s.store.Delete(code); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
