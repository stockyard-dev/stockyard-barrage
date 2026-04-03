package server
import ("encoding/json";"log";"net/http";"github.com/stockyard-dev/stockyard-barrage/internal/store")
type Server struct{db *store.DB;mux *http.ServeMux;limits Limits}
func New(db *store.DB,limits Limits)*Server{s:=&Server{db:db,mux:http.NewServeMux(),limits:limits}
s.mux.HandleFunc("GET /api/tests",s.listTests);s.mux.HandleFunc("POST /api/tests",s.createTest);s.mux.HandleFunc("GET /api/tests/{id}",s.getTest);s.mux.HandleFunc("DELETE /api/tests/{id}",s.deleteTest)
s.mux.HandleFunc("POST /api/tests/{id}/run",s.runTest);s.mux.HandleFunc("GET /api/tests/{id}/runs",s.listRuns);s.mux.HandleFunc("GET /api/runs/{id}",s.getRun)
s.mux.HandleFunc("GET /api/stats",s.stats);s.mux.HandleFunc("GET /api/health",s.health)
s.mux.HandleFunc("GET /ui",s.dashboard);s.mux.HandleFunc("GET /ui/",s.dashboard);s.mux.HandleFunc("GET /",s.root);
s.mux.HandleFunc("GET /api/tier",func(w http.ResponseWriter,r *http.Request){wj(w,200,map[string]any{"tier":s.limits.Tier,"upgrade_url":"https://stockyard.dev/barrage/"})})
return s}
func(s *Server)ServeHTTP(w http.ResponseWriter,r *http.Request){s.mux.ServeHTTP(w,r)}
func wj(w http.ResponseWriter,c int,v any){w.Header().Set("Content-Type","application/json");w.WriteHeader(c);json.NewEncoder(w).Encode(v)}
func we(w http.ResponseWriter,c int,m string){wj(w,c,map[string]string{"error":m})}
func(s *Server)root(w http.ResponseWriter,r *http.Request){if r.URL.Path!="/"{http.NotFound(w,r);return};http.Redirect(w,r,"/ui",302)}
func(s *Server)listTests(w http.ResponseWriter,r *http.Request){wj(w,200,map[string]any{"tests":oe(s.db.ListTests())})}
func(s *Server)createTest(w http.ResponseWriter,r *http.Request){var t store.Test;json.NewDecoder(r.Body).Decode(&t);if t.Name==""||t.URL==""{we(w,400,"name and url required");return};s.db.CreateTest(&t);wj(w,201,s.db.GetTest(t.ID))}
func(s *Server)getTest(w http.ResponseWriter,r *http.Request){t:=s.db.GetTest(r.PathValue("id"));if t==nil{we(w,404,"not found");return};wj(w,200,t)}
func(s *Server)deleteTest(w http.ResponseWriter,r *http.Request){s.db.DeleteTest(r.PathValue("id"));wj(w,200,map[string]string{"deleted":"ok"})}
func(s *Server)runTest(w http.ResponseWriter,r *http.Request){t:=s.db.GetTest(r.PathValue("id"));if t==nil{we(w,404,"not found");return}
run,err:=s.db.Execute(t);if err!=nil{we(w,500,err.Error());return};wj(w,200,run)}
func(s *Server)listRuns(w http.ResponseWriter,r *http.Request){wj(w,200,map[string]any{"runs":oe(s.db.ListRuns(r.PathValue("id")))})}
func(s *Server)getRun(w http.ResponseWriter,r *http.Request){run:=s.db.GetRun(r.PathValue("id"));if run==nil{we(w,404,"not found");return};wj(w,200,run)}
func(s *Server)stats(w http.ResponseWriter,r *http.Request){wj(w,200,s.db.Stats())}
func(s *Server)health(w http.ResponseWriter,r *http.Request){st:=s.db.Stats();wj(w,200,map[string]any{"status":"ok","service":"barrage","tests":st.Tests})}
func oe[T any](s []T)[]T{if s==nil{return[]T{}};return s}
func init(){log.SetFlags(log.LstdFlags|log.Lshortfile)}
