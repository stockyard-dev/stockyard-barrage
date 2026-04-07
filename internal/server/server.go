package server

import (
	"encoding/json"
	"github.com/stockyard-dev/stockyard-barrage/internal/store"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type Server struct {
	db      *store.DB
	mux     *http.ServeMux
	limits  Limits
	dataDir string
	pCfg    map[string]json.RawMessage
}

func New(db *store.DB, limits Limits, dataDir string) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), limits: limits, dataDir: dataDir}
	s.mux.HandleFunc("GET /api/tests", s.listTests)
	s.mux.HandleFunc("POST /api/tests", s.createTest)
	s.mux.HandleFunc("GET /api/tests/{id}", s.getTest)
	s.mux.HandleFunc("DELETE /api/tests/{id}", s.deleteTest)
	s.mux.HandleFunc("POST /api/tests/{id}/run", s.runTest)
	s.mux.HandleFunc("GET /api/tests/{id}/runs", s.listRuns)
	s.mux.HandleFunc("GET /api/runs/{id}", s.getRun)
	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/health", s.health)
	s.mux.HandleFunc("GET /ui", s.dashboard)
	s.mux.HandleFunc("GET /ui/", s.dashboard)
	s.mux.HandleFunc("GET /", s.root)
	s.mux.HandleFunc("GET /api/tier", func(w http.ResponseWriter, r *http.Request) {
		wj(w, 200, map[string]any{"tier": s.limits.Tier, "upgrade_url": "https://stockyard.dev/barrage/"})
	})
	s.loadPersonalConfig()
	s.mux.HandleFunc("GET /api/config", s.configHandler)
	s.mux.HandleFunc("GET /api/extras/{resource}", s.listExtras)
	s.mux.HandleFunc("GET /api/extras/{resource}/{id}", s.getExtras)
	s.mux.HandleFunc("PUT /api/extras/{resource}/{id}", s.putExtras)
	return s
}
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }
func wj(w http.ResponseWriter, c int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(c)
	json.NewEncoder(w).Encode(v)
}
func we(w http.ResponseWriter, c int, m string) { wj(w, c, map[string]string{"error": m}) }
func (s *Server) root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/ui", 302)
}
func (s *Server) listTests(w http.ResponseWriter, r *http.Request) {
	wj(w, 200, map[string]any{"tests": oe(s.db.ListTests())})
}
func (s *Server) createTest(w http.ResponseWriter, r *http.Request) {
	var t store.Test
	json.NewDecoder(r.Body).Decode(&t)
	if t.Name == "" || t.URL == "" {
		we(w, 400, "name and url required")
		return
	}
	s.db.CreateTest(&t)
	wj(w, 201, s.db.GetTest(t.ID))
}
func (s *Server) getTest(w http.ResponseWriter, r *http.Request) {
	t := s.db.GetTest(r.PathValue("id"))
	if t == nil {
		we(w, 404, "not found")
		return
	}
	wj(w, 200, t)
}
func (s *Server) deleteTest(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteTest(r.PathValue("id"))
	wj(w, 200, map[string]string{"deleted": "ok"})
}
func (s *Server) runTest(w http.ResponseWriter, r *http.Request) {
	t := s.db.GetTest(r.PathValue("id"))
	if t == nil {
		we(w, 404, "not found")
		return
	}
	run, err := s.db.Execute(t)
	if err != nil {
		we(w, 500, err.Error())
		return
	}
	wj(w, 200, run)
}
func (s *Server) listRuns(w http.ResponseWriter, r *http.Request) {
	wj(w, 200, map[string]any{"runs": oe(s.db.ListRuns(r.PathValue("id")))})
}
func (s *Server) getRun(w http.ResponseWriter, r *http.Request) {
	run := s.db.GetRun(r.PathValue("id"))
	if run == nil {
		we(w, 404, "not found")
		return
	}
	wj(w, 200, run)
}
func (s *Server) stats(w http.ResponseWriter, r *http.Request) { wj(w, 200, s.db.Stats()) }
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	st := s.db.Stats()
	wj(w, 200, map[string]any{"status": "ok", "service": "barrage", "tests": st.Tests})
}
func oe[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
func init() { log.SetFlags(log.LstdFlags | log.Lshortfile) }

// ─── personalization (auto-added) ──────────────────────────────────

func (s *Server) loadPersonalConfig() {
	path := filepath.Join(s.dataDir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var cfg map[string]json.RawMessage
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("%s: warning: could not parse config.json: %v", "barrage", err)
		return
	}
	s.pCfg = cfg
	log.Printf("%s: loaded personalization from %s", "barrage", path)
}

func (s *Server) configHandler(w http.ResponseWriter, r *http.Request) {
	if s.pCfg == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.pCfg)
}

func (s *Server) listExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	all := s.db.AllExtras(resource)
	out := make(map[string]json.RawMessage, len(all))
	for id, data := range all {
		out[id] = json.RawMessage(data)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (s *Server) getExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	id := r.PathValue("id")
	data := s.db.GetExtras(resource, id)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(data))
}

func (s *Server) putExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	id := r.PathValue("id")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"read body"}`, 400)
		return
	}
	var probe map[string]any
	if err := json.Unmarshal(body, &probe); err != nil {
		http.Error(w, `{"error":"invalid json"}`, 400)
		return
	}
	if err := s.db.SetExtras(resource, id, string(body)); err != nil {
		http.Error(w, `{"error":"save failed"}`, 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":"saved"}`))
}
