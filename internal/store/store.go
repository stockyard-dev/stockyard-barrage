package store
import ("database/sql";"fmt";"io";"net/http";"os";"path/filepath";"sync";"time";_ "modernc.org/sqlite")
type DB struct{db *sql.DB}
type Test struct{ID string `json:"id"`;Name string `json:"name"`;URL string `json:"url"`;Method string `json:"method"`;Headers string `json:"headers,omitempty"`;Body string `json:"body,omitempty"`;Concurrency int `json:"concurrency"`;Requests int `json:"requests"`;CreatedAt string `json:"created_at"`;RunCount int `json:"run_count"`}
type Run struct{ID string `json:"id"`;TestID string `json:"test_id"`;Status string `json:"status"`;TotalRequests int `json:"total_requests"`;Successes int `json:"successes"`;Failures int `json:"failures"`;AvgMs float64 `json:"avg_ms"`;MinMs float64 `json:"min_ms"`;MaxMs float64 `json:"max_ms"`;P99Ms float64 `json:"p99_ms"`;ReqPerSec float64 `json:"req_per_sec"`;DurationMs int `json:"duration_ms"`;StartedAt string `json:"started_at"`;FinishedAt string `json:"finished_at,omitempty"`}
func Open(d string)(*DB,error){if err:=os.MkdirAll(d,0755);err!=nil{return nil,err};db,err:=sql.Open("sqlite",filepath.Join(d,"barrage.db")+"?_journal_mode=WAL&_busy_timeout=5000");if err!=nil{return nil,err}
for _,q:=range[]string{
`CREATE TABLE IF NOT EXISTS tests(id TEXT PRIMARY KEY,name TEXT NOT NULL,url TEXT NOT NULL,method TEXT DEFAULT 'GET',headers TEXT DEFAULT '',body TEXT DEFAULT '',concurrency INTEGER DEFAULT 10,requests INTEGER DEFAULT 100,created_at TEXT DEFAULT(datetime('now')))`,
`CREATE TABLE IF NOT EXISTS runs(id TEXT PRIMARY KEY,test_id TEXT NOT NULL,status TEXT DEFAULT 'running',total_requests INTEGER DEFAULT 0,successes INTEGER DEFAULT 0,failures INTEGER DEFAULT 0,avg_ms REAL DEFAULT 0,min_ms REAL DEFAULT 0,max_ms REAL DEFAULT 0,p99_ms REAL DEFAULT 0,req_per_sec REAL DEFAULT 0,duration_ms INTEGER DEFAULT 0,started_at TEXT,finished_at TEXT DEFAULT '')`,
`CREATE INDEX IF NOT EXISTS idx_runs_test ON runs(test_id)`,
}{if _,err:=db.Exec(q);err!=nil{return nil,fmt.Errorf("migrate: %w",err)}};return &DB{db:db},nil}
func(d *DB)Close()error{return d.db.Close()}
func genID()string{return fmt.Sprintf("%d",time.Now().UnixNano())}
func now()string{return time.Now().UTC().Format(time.RFC3339)}
func(d *DB)CreateTest(t *Test)error{t.ID=genID();t.CreatedAt=now();if t.Method==""{t.Method="GET"};if t.Concurrency<=0{t.Concurrency=10};if t.Requests<=0{t.Requests=100}
_,err:=d.db.Exec(`INSERT INTO tests VALUES(?,?,?,?,?,?,?,?,?)`,t.ID,t.Name,t.URL,t.Method,t.Headers,t.Body,t.Concurrency,t.Requests,t.CreatedAt);return err}
func(d *DB)GetTest(id string)*Test{var t Test;if d.db.QueryRow(`SELECT id,name,url,method,headers,body,concurrency,requests,created_at FROM tests WHERE id=?`,id).Scan(&t.ID,&t.Name,&t.URL,&t.Method,&t.Headers,&t.Body,&t.Concurrency,&t.Requests,&t.CreatedAt)!=nil{return nil};d.db.QueryRow(`SELECT COUNT(*) FROM runs WHERE test_id=?`,id).Scan(&t.RunCount);return &t}
func(d *DB)ListTests()[]Test{rows,_:=d.db.Query(`SELECT id,name,url,method,headers,body,concurrency,requests,created_at FROM tests ORDER BY created_at DESC`);if rows==nil{return nil};defer rows.Close()
var o []Test;for rows.Next(){var t Test;rows.Scan(&t.ID,&t.Name,&t.URL,&t.Method,&t.Headers,&t.Body,&t.Concurrency,&t.Requests,&t.CreatedAt);d.db.QueryRow(`SELECT COUNT(*) FROM runs WHERE test_id=?`,t.ID).Scan(&t.RunCount);o=append(o,t)};return o}
func(d *DB)DeleteTest(id string)error{d.db.Exec(`DELETE FROM runs WHERE test_id=?`,id);_,err:=d.db.Exec(`DELETE FROM tests WHERE id=?`,id);return err}

func(d *DB)Execute(t *Test)(*Run,error){
run:=&Run{ID:genID(),TestID:t.ID,Status:"running",StartedAt:now()}
d.db.Exec(`INSERT INTO runs(id,test_id,status,started_at)VALUES(?,?,?,?)`,run.ID,run.TestID,run.Status,run.StartedAt)
go d.runLoad(run,t);return run,nil}

func(d *DB)runLoad(run *Run,t *Test){
client:=&http.Client{Timeout:10*time.Second}
var mu sync.Mutex;var latencies []float64;successes,failures:=0,0
sem:=make(chan struct{},t.Concurrency);var wg sync.WaitGroup
start:=time.Now()
for i:=0;i<t.Requests;i++{sem<-struct{}{};wg.Add(1)
go func(){defer wg.Done();defer func(){<-sem}()
req,err:=http.NewRequest(t.Method,t.URL,nil);if err!=nil{mu.Lock();failures++;mu.Unlock();return}
s:=time.Now();resp,err:=client.Do(req)
lat:=float64(time.Since(s).Milliseconds())
mu.Lock();latencies=append(latencies,lat)
if err!=nil||resp==nil{failures++}else{if resp.StatusCode<400{successes++}else{failures++};io.Copy(io.Discard,resp.Body);resp.Body.Close()};mu.Unlock()}()}
wg.Wait();elapsed:=time.Since(start)
run.TotalRequests=successes+failures;run.Successes=successes;run.Failures=failures;run.DurationMs=int(elapsed.Milliseconds())
if len(latencies)>0{total:=0.0;run.MinMs=latencies[0];run.MaxMs=latencies[0]
for _,l:=range latencies{total+=l;if l<run.MinMs{run.MinMs=l};if l>run.MaxMs{run.MaxMs=l}}
run.AvgMs=total/float64(len(latencies))
// Simple p99
idx:=int(float64(len(latencies))*0.99);if idx>=len(latencies){idx=len(latencies)-1};run.P99Ms=latencies[idx]}
if elapsed.Seconds()>0{run.ReqPerSec=float64(run.TotalRequests)/elapsed.Seconds()}
run.Status="done";run.FinishedAt=now()
d.db.Exec(`UPDATE runs SET status=?,total_requests=?,successes=?,failures=?,avg_ms=?,min_ms=?,max_ms=?,p99_ms=?,req_per_sec=?,duration_ms=?,finished_at=? WHERE id=?`,
run.Status,run.TotalRequests,run.Successes,run.Failures,run.AvgMs,run.MinMs,run.MaxMs,run.P99Ms,run.ReqPerSec,run.DurationMs,run.FinishedAt,run.ID)}

func(d *DB)GetRun(id string)*Run{var r Run;if d.db.QueryRow(`SELECT id,test_id,status,total_requests,successes,failures,avg_ms,min_ms,max_ms,p99_ms,req_per_sec,duration_ms,started_at,finished_at FROM runs WHERE id=?`,id).Scan(&r.ID,&r.TestID,&r.Status,&r.TotalRequests,&r.Successes,&r.Failures,&r.AvgMs,&r.MinMs,&r.MaxMs,&r.P99Ms,&r.ReqPerSec,&r.DurationMs,&r.StartedAt,&r.FinishedAt)!=nil{return nil};return &r}
func(d *DB)ListRuns(testID string)[]Run{rows,_:=d.db.Query(`SELECT id,test_id,status,total_requests,successes,failures,avg_ms,min_ms,max_ms,p99_ms,req_per_sec,duration_ms,started_at,finished_at FROM runs WHERE test_id=? ORDER BY started_at DESC LIMIT 20`,testID);if rows==nil{return nil};defer rows.Close()
var o []Run;for rows.Next(){var r Run;rows.Scan(&r.ID,&r.TestID,&r.Status,&r.TotalRequests,&r.Successes,&r.Failures,&r.AvgMs,&r.MinMs,&r.MaxMs,&r.P99Ms,&r.ReqPerSec,&r.DurationMs,&r.StartedAt,&r.FinishedAt);o=append(o,r)};return o}
type Stats struct{Tests int `json:"tests"`;Runs int `json:"runs"`}
func(d *DB)Stats()Stats{var s Stats;d.db.QueryRow(`SELECT COUNT(*) FROM tests`).Scan(&s.Tests);d.db.QueryRow(`SELECT COUNT(*) FROM runs`).Scan(&s.Runs);return s}
