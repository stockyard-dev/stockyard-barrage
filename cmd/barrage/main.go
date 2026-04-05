package main
import ("fmt";"log";"net/http";"os";"github.com/stockyard-dev/stockyard-barrage/internal/server";"github.com/stockyard-dev/stockyard-barrage/internal/store")
func main(){port:=os.Getenv("PORT");if port==""{port="8600"};dataDir:=os.Getenv("DATA_DIR");if dataDir==""{dataDir="./barrage-data"}
db,err:=store.Open(dataDir);if err!=nil{log.Fatalf("barrage: %v",err)};defer db.Close();srv:=server.New(db,server.DefaultLimits())
fmt.Printf("\n  Barrage — Self-hosted load tester\n  ─────────────────────────────────\n  Dashboard:  http://localhost:%s/ui\n  API:        http://localhost:%s/api\n  Data:       %s\n  ─────────────────────────────────\n  Questions? hello@stockyard.dev\n\n",port,port,dataDir)
log.Printf("barrage: listening on :%s",port);log.Fatal(http.ListenAndServe(":"+port,srv))}
