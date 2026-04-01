package server
import("encoding/json";"fmt";"math";"net/http";"sort";"strconv";"sync";"sync/atomic";"time";"github.com/stockyard-dev/stockyard-barrage/internal/store")
func(s *Server)handleListScenarios(w http.ResponseWriter,r *http.Request){list,_:=s.db.ListScenarios();if list==nil{list=[]store.Scenario{}};writeJSON(w,200,list)}
func(s *Server)handleGetScenario(w http.ResponseWriter,r *http.Request){id,_:=strconv.ParseInt(r.PathValue("id"),10,64);sc,_:=s.db.GetScenario(id);if sc==nil{writeError(w,404,"not found");return};writeJSON(w,200,sc)}
func(s *Server)handleCreateScenario(w http.ResponseWriter,r *http.Request){
    if !s.limits.IsPro(){n,_:=s.db.CountScenarios();if n>=3{writeError(w,403,"free tier: 3 scenarios max");return}}
    var sc store.Scenario;json.NewDecoder(r.Body).Decode(&sc)
    if sc.TargetURL==""{writeError(w,400,"target_url required");return}
    if sc.Name==""{sc.Name=sc.TargetURL};if sc.Method==""{sc.Method="GET"}
    if sc.Concurrency<=0{sc.Concurrency=10};if sc.DurationSec<=0{sc.DurationSec=30}
    if err:=s.db.CreateScenario(&sc);err!=nil{writeError(w,500,err.Error());return}
    writeJSON(w,201,sc)}
func(s *Server)handleDeleteScenario(w http.ResponseWriter,r *http.Request){id,_:=strconv.ParseInt(r.PathValue("id"),10,64);s.db.DeleteScenario(id);writeJSON(w,200,map[string]string{"status":"deleted"})}
func(s *Server)handleRunScenario(w http.ResponseWriter,r *http.Request){
    id,_:=strconv.ParseInt(r.PathValue("id"),10,64)
    sc,_:=s.db.GetScenario(id);if sc==nil{writeError(w,404,"scenario not found");return}
    run:=&store.Run{ScenarioID:sc.ID};s.db.CreateRun(run)
    go func(){
        var total,success,errCount int64
        lats:=make([]int64,0,1000)
        var mu sync.Mutex
        deadline:=time.Now().Add(time.Duration(sc.DurationSec)*time.Second)
        sem:=make(chan struct{},sc.Concurrency)
        var wg sync.WaitGroup
        client:=&http.Client{Timeout:10*time.Second}
        for time.Now().Before(deadline){
            sem<-struct{}{};wg.Add(1)
            go func(){
                defer func(){<-sem;wg.Done()}()
                req,err:=http.NewRequest(sc.Method,sc.TargetURL,nil)
                if err!=nil{atomic.AddInt64(&errCount,1);return}
                t0:=time.Now();resp,err:=client.Do(req);lat:=time.Since(t0).Milliseconds()
                atomic.AddInt64(&total,1)
                if err!=nil||resp.StatusCode>=500{atomic.AddInt64(&errCount,1)}else{atomic.AddInt64(&success,1);resp.Body.Close()}
                mu.Lock();lats=append(lats,lat);mu.Unlock()
            }()
        }
        wg.Wait()
        sort.Slice(lats,func(i,j int)bool{return lats[i]<lats[j]})
        var avgLat,p99Lat float64
        if len(lats)>0{sum:=int64(0);for _,l:=range lats{sum+=l};avgLat=float64(sum)/float64(len(lats));idx:=int(math.Ceil(float64(len(lats))*0.99))-1;if idx<0{idx=0};p99Lat=float64(lats[idx])}
        run.TotalReqs=int(total);run.SuccessCount=int(success);run.ErrorCount=int(errCount)
        run.AvgLatMs=avgLat;run.P99LatMs=p99Lat;run.RPS=float64(total)/float64(sc.DurationSec)
        s.db.FinishRun(run)
    }()
    writeJSON(w,202,map[string]interface{}{"run_id":run.ID,"status":"started","message":fmt.Sprintf("Load test started: %d concurrent for %ds",sc.Concurrency,sc.DurationSec)})}
func(s *Server)handleListRuns(w http.ResponseWriter,r *http.Request){list,_:=s.db.ListRuns();if list==nil{list=[]store.Run{}};writeJSON(w,200,list)}
func(s *Server)handleStats(w http.ResponseWriter,r *http.Request){sc,_:=s.db.CountScenarios();ru,_:=s.db.CountRuns();writeJSON(w,200,map[string]interface{}{"scenarios":sc,"runs":ru})}
