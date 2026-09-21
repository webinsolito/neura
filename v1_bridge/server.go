package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Server struct{core *Core;store *Store;tools *ToolRegistry;model ModelAdapter;allowedOrigin string}
func writeJSON(w http.ResponseWriter,status int,v any){w.Header().Set("Content-Type","application/json");w.WriteHeader(status);_=json.NewEncoder(w).Encode(v)}
func decodeJSON(w http.ResponseWriter,r *http.Request,dst any)error{r.Body=http.MaxBytesReader(w,r.Body,64*1024);dec:=json.NewDecoder(r.Body);dec.DisallowUnknownFields();if err:=dec.Decode(dst);err!=nil{return err};var extra any;err:=dec.Decode(&extra);if errors.Is(err,io.EOF){return nil};if err==nil{return errors.New("multiple json values")};return fmt.Errorf("trailing json rejected: %w",err)}
func(s *Server)middleware(next http.Handler)http.Handler{return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
	if !loopbackRemote(r.RemoteAddr){writeJSON(w,http.StatusForbidden,map[string]string{"error":"loopback only"});return}
	origin:=r.Header.Get("Origin");if origin!=""{if origin!=s.allowedOrigin&&origin!="http://localhost:8000"&&origin!="http://127.0.0.1:8000"{writeJSON(w,http.StatusForbidden,map[string]string{"error":"origin denied"});return};w.Header().Set("Access-Control-Allow-Origin",origin);w.Header().Set("Vary","Origin");w.Header().Set("Access-Control-Allow-Private-Network","true")}
	w.Header().Set("Access-Control-Allow-Headers","Content-Type");w.Header().Set("Access-Control-Allow-Methods","GET,POST,OPTIONS");w.Header().Set("X-Content-Type-Options","nosniff");w.Header().Set("Cache-Control","no-store")
	if r.Method==http.MethodOptions{w.WriteHeader(http.StatusNoContent);return};next.ServeHTTP(w,r)
})}
func(s *Server)routes()http.Handler{mux:=http.NewServeMux()
	mux.HandleFunc("GET /health",func(w http.ResponseWriter,r *http.Request){m,rc:=s.store.Counts();writeJSON(w,200,map[string]any{"ok":true,"version":version,"time":time.Now().UTC(),"model_configured":s.model.Available(),"memory_count":m,"receipt_count":rc})})
	mux.HandleFunc("GET /status",func(w http.ResponseWriter,r *http.Request){m,rc:=s.store.Counts();writeJSON(w,200,map[string]any{"version":version,"core":"online","model":map[bool]string{true:"available",false:"not_configured"}[s.model.Available()],"memory_count":m,"receipt_count":rc,"workspace":s.tools.workspace,"windows_computer_use":"observe_only_rc1","stable_candidate_policy":"candidate_only_no_auto_promotion","candidate":"candidate/v1-integration-rc1","stable":"unchanged","security":map[string]any{"loopback_only":true,"origin_allowlist":true,"shell_execution":false,"tool_allowlist":true,"destructive_tools":0,"max_body_bytes":64*1024}})})
	mux.HandleFunc("POST /command",func(w http.ResponseWriter,r *http.Request){var req struct{Goal string `json:"goal"`;Message string `json:"message"`};if err:=decodeJSON(w,r,&req);err!=nil{writeJSON(w,400,map[string]string{"error":err.Error()});return};goal:=req.Goal;if goal==""{goal=req.Message};ctx,cancel:=context.WithTimeout(r.Context(),65*time.Second);defer cancel();res:=s.core.Execute(ctx,goal);code:=200;if res.Status=="blocked"{code=400}else if res.Status=="error"{code=500}else if res.Status=="unavailable"{code=503};writeJSON(w,code,res)})
	mux.HandleFunc("GET /memory",func(w http.ResponseWriter,r *http.Request){q:=r.URL.Query().Get("q");n,_:=strconv.Atoi(r.URL.Query().Get("limit"));writeJSON(w,200,map[string]any{"items":s.store.SearchMemory(q,n)})})
	mux.HandleFunc("POST /memory",func(w http.ResponseWriter,r *http.Request){var req struct{Text string `json:"text"`};if err:=decodeJSON(w,r,&req);err!=nil{writeJSON(w,400,map[string]string{"error":err.Error()});return};m,created,err:=s.store.SaveMemory(req.Text);if err!=nil{writeJSON(w,400,map[string]string{"error":err.Error()});return};writeJSON(w,200,map[string]any{"memory":m,"created":created})})
	mux.HandleFunc("GET /receipts",func(w http.ResponseWriter,r *http.Request){n,_:=strconv.Atoi(r.URL.Query().Get("limit"));writeJSON(w,200,map[string]any{"items":s.store.RecentReceipts(n)})})
	mux.HandleFunc("GET /tools",func(w http.ResponseWriter,r *http.Request){writeJSON(w,200,map[string]any{"tools":s.tools.List(),"destructive_tools":[]string{}})})
	mux.HandleFunc("GET /windows/status",func(w http.ResponseWriter,r *http.Request){writeJSON(w,200,map[string]any{"available":runtime.GOOS=="windows","observe":map[bool]string{true:"available",false:"not_on_windows"}[runtime.GOOS=="windows"],"understand":"deterministic_v1","preflight":"allowlist_and_timeout","action":"not_implemented_in_rc1","verify":"read_only_observation_verified","pipeline":[]string{"observe","understand","preflight","action","verify"}})})
	return s.middleware(mux)
}
func defaultDataDir()string{d,err:=os.UserConfigDir();if err!=nil{return filepath.Join(".",".neura")};return filepath.Join(d,"NEURA","v1")}
func main(){
	listen:=flag.String("listen","127.0.0.1:8765","loopback listen address");data:=flag.String("data",defaultDataDir(),"persistent data directory");workspace:=flag.String("workspace",".","allowed filesystem workspace");modelExec:=flag.String("model-exec","","optional absolute path to local model planner executable");ollamaModel:=flag.String("ollama-model","","optional local Ollama model name");ollamaURL:=flag.String("ollama-url","http://127.0.0.1:11434","loopback Ollama URL");flag.Parse()
	if !strings.HasPrefix(*listen,"127.0.0.1:")&&!strings.HasPrefix(*listen,"[::1]:"){log.Fatal("listen address must be loopback")}
	store,err:=NewStore(*data);if err!=nil{log.Fatal(err)};tools,err:=NewToolRegistry(*workspace);if err!=nil{log.Fatal(err)};model:=ModelAdapter{Executable:*modelExec,Timeout:20*time.Second,OllamaURL:*ollamaURL,OllamaModel:*ollamaModel};core:=&Core{store:store,tools:tools,model:model};srv:=&Server{core:core,store:store,tools:tools,model:model,allowedOrigin:"https://webinsolito.github.io"}
	httpSrv:=&http.Server{Addr:*listen,Handler:srv.routes(),ReadHeaderTimeout:5*time.Second,ReadTimeout:70*time.Second,WriteTimeout:70*time.Second,IdleTimeout:90*time.Second,MaxHeaderBytes:32*1024}
	fmt.Printf("NEURA V1 bridge %s listening on http://%s\n",version,*listen);log.Fatal(httpSrv.ListenAndServe())
}
