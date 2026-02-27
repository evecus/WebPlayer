package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
	_ "embed"
)

//go:embed index.html
var indexHTML []byte

// ────────────────────────────── Data Models ──────────────────────────────

type Channel struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	URL     string `json:"url"`
	Group   string `json:"group"`
	Logo    string `json:"logo"`
	AddedAt int64  `json:"addedAt"`
}

type Playlist struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Channels  []Channel `json:"channels"`
	CreatedAt int64     `json:"createdAt"`
}

type Store struct {
	Playlists []Playlist `json:"playlists"`
	History   []Channel  `json:"history"`
}

// ────────────────────────────── Storage ──────────────────────────────────

type Storage struct {
	mu       sync.RWMutex
	filePath string
	data     Store
}

func NewStorage(path string) (*Storage, error) {
	s := &Storage{filePath: path}
	if err := s.load(); err != nil {
		// first run, init empty
		s.data = Store{Playlists: []Playlist{}, History: []Channel{}}
	}
	return s, nil
}

func (s *Storage) load() error {
	b, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &s.data)
}

func (s *Storage) save() error {
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, b, 0644)
}

func (s *Storage) GetAll() Store {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data
}

func (s *Storage) AddPlaylist(p Playlist) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// check duplicate id
	for i, existing := range s.data.Playlists {
		if existing.ID == p.ID {
			s.data.Playlists[i] = p
			return s.save()
		}
	}
	s.data.Playlists = append(s.data.Playlists, p)
	return s.save()
}

func (s *Storage) DeletePlaylist(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	newList := s.data.Playlists[:0]
	for _, p := range s.data.Playlists {
		if p.ID != id {
			newList = append(newList, p)
		}
	}
	s.data.Playlists = newList
	return s.save()
}

func (s *Storage) AddHistory(ch Channel) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// remove existing entry with same URL
	filtered := s.data.History[:0]
	for _, h := range s.data.History {
		if h.URL != ch.URL {
			filtered = append(filtered, h)
		}
	}
	// prepend
	s.data.History = append([]Channel{ch}, filtered...)
	// keep max 50
	if len(s.data.History) > 50 {
		s.data.History = s.data.History[:50]
	}
	return s.save()
}

func (s *Storage) ClearHistory() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.History = []Channel{}
	return s.save()
}

// ────────────────────────────── Helpers ──────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func readJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func genID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// ────────────────────────────── Server ───────────────────────────────────

type Server struct {
	store *Storage
	mux   *http.ServeMux
}

func NewServer(store *Storage) *Server {
	s := &Server{store: store, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) routes() {
	// Frontend
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexHTML)
	})

	// API
	s.mux.HandleFunc("/api/data", s.handleData)
	s.mux.HandleFunc("/api/playlists", s.handlePlaylists)
	s.mux.HandleFunc("/api/playlists/", s.handlePlaylistByID)
	s.mux.HandleFunc("/api/history", s.handleHistory)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		w.WriteHeader(204)
		return
	}
	s.mux.ServeHTTP(w, r)
}

// GET /api/data — return everything
func (s *Server) handleData(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, s.store.GetAll())
}

// POST /api/playlists — create / upsert playlist
// GET  /api/playlists — list all
func (s *Server) handlePlaylists(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, 200, s.store.GetAll().Playlists)
	case http.MethodPost:
		var p Playlist
		if err := readJSON(r, &p); err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		if p.ID == "" {
			p.ID = genID()
		}
		if p.CreatedAt == 0 {
			p.CreatedAt = time.Now().Unix()
		}
		if err := s.store.AddPlaylist(p); err != nil {
			writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, p)
	default:
		w.WriteHeader(405)
	}
}

// DELETE /api/playlists/{id}
func (s *Server) handlePlaylistByID(w http.ResponseWriter, r *http.Request) {
	id := filepath.Base(r.URL.Path)
	if r.Method != http.MethodDelete {
		w.WriteHeader(405)
		return
	}
	if err := s.store.DeletePlaylist(id); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]string{"ok": "deleted"})
}

// POST /api/history  — add entry
// GET  /api/history  — list
// DELETE /api/history — clear
func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, 200, s.store.GetAll().History)
	case http.MethodPost:
		var ch Channel
		if err := readJSON(r, &ch); err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		if ch.ID == "" {
			ch.ID = genID()
		}
		ch.AddedAt = time.Now().Unix()
		if err := s.store.AddHistory(ch); err != nil {
			writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, ch)
	case http.MethodDelete:
		if err := s.store.ClearHistory(); err != nil {
			writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, map[string]string{"ok": "cleared"})
	default:
		w.WriteHeader(405)
	}
}

// ────────────────────────────── Main ─────────────────────────────────────

func main() {
	port := flag.Int("port", 4001, "HTTP server port")
	dataFile := flag.String("data", "webplayer-data.json", "Path to data persistence file")
	flag.Parse()

	store, err := NewStorage(*dataFile)
	if err != nil {
		log.Printf("Storage init (new file will be created): %v", err)
	}

	srv := NewServer(store)

	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	log.Printf("🎬 WebPlayer running at http://%s", addr)
	log.Printf("   Data stored in: %s", *dataFile)
	log.Printf("   Press Ctrl+C to stop")

	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatal(err)
	}
}
