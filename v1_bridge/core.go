package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const version = "1.0.0-rc2-capability-gate"

type Memory struct {
	ID         string     `json:"id"`
	Text       string     `json:"text"`
	Persistent bool       `json:"persistent"`
	CreatedAt  time.Time  `json:"created_at"`
	Entities   []string   `json:"entities,omitempty"`
	Embedding  []float32  `json:"embedding,omitempty"`
	Type       string     `json:"type,omitempty"`
	Provenance string     `json:"provenance,omitempty"`
	ValidFrom  time.Time  `json:"valid_from,omitempty"`
	ValidUntil *time.Time `json:"valid_until,omitempty"`
	Supersedes string     `json:"supersedes,omitempty"`
}
type Receipt struct {
	ID string `json:"id"`
	Goal string `json:"goal,omitempty"`
	Action string `json:"action"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
	Error string `json:"error,omitempty"`
	StartedAt time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
}
type Store struct {
	mu sync.Mutex
	dir string
	memories []Memory
	receipts []Receipt
}
func NewStore(dir string)(*Store,error){
	if dir==""{return nil,errors.New("data directory required")}
	abs,err:=filepath.Abs(dir);if err!=nil{return nil,err}
	if err:=os.MkdirAll(abs,0o700);err!=nil{return nil,err}
	s:=&Store{dir:abs}
	if _,err:=recoverJSONL(filepath.Join(abs,"memory.jsonl"));err!=nil{return nil,err}
	if _,err:=recoverJSONL(filepath.Join(abs,"receipts.jsonl"));err!=nil{return nil,err}
	if err:=s.loadJSONL(filepath.Join(abs,"memory.jsonl"),func(b []byte)error{var m Memory;if err:=json.Unmarshal(b,&m);err!=nil{return err};s.memories=append(s.memories,m);return nil});err!=nil{return nil,err}
	if err:=s.loadJSONL(filepath.Join(abs,"receipts.jsonl"),func(b []byte)error{var r Receipt;if err:=json.Unmarshal(b,&r);err!=nil{return err};s.receipts=append(s.receipts,r);return nil});err!=nil{return nil,err}
	return s,nil
}
func (s *Store)loadJSONL(path string,fn func([]byte)error)error{
	f,err:=os.Open(path);if errors.Is(err,os.ErrNotExist){return nil};if err!=nil{return err};defer f.Close()
	sc:=bufio.NewScanner(f);buf:=make([]byte,0,64*1024);sc.Buffer(buf,1024*1024);line:=0
	for sc.Scan(){line++;b:=append([]byte(nil),sc.Bytes()...);if len(strings.TrimSpace(string(b)))==0{continue};if err:=fn(b);err!=nil{return fmt.Errorf("corrupt %s line %d: %w",filepath.Base(path),line,err)}}
	return sc.Err()
}
func validateJSONL(path string) error {
	f, err := os.Open(path); if errors.Is(err, os.ErrNotExist) { return nil }; if err != nil { return err }; defer f.Close()
	sc := bufio.NewScanner(f); sc.Buffer(make([]byte, 0, 64*1024), 1024*1024); line := 0
	for sc.Scan() { line++; b := strings.TrimSpace(sc.Text()); if b == "" { continue }; var raw json.RawMessage; if err := json.Unmarshal([]byte(b), &raw); err != nil { return fmt.Errorf("invalid jsonl line %d: %w", line, err) } }
	return sc.Err()
}
func copyFileAtomic(src, dst string) error {
	in, err := os.Open(src); if err != nil { return err }; defer in.Close(); tmp := dst+".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600); if err != nil { return err }; ok:=false
	defer func(){ _=out.Close(); if !ok { _=os.Remove(tmp) } }()
	if _,err=io.Copy(out,in);err!=nil{return err};if err=out.Sync();err!=nil{return err};if err=out.Close();err!=nil{return err};if err=os.Rename(tmp,dst);err!=nil{return err};ok=true;return nil
}
func recoverJSONL(path string)(bool,error){if err:=validateJSONL(path);err==nil{return false,nil};bak:=path+".bak";if err:=validateJSONL(bak);err!=nil{return false,fmt.Errorf("primary corrupt and backup invalid: %w",err)};if _,err:=os.Stat(bak);err!=nil{return false,fmt.Errorf("primary corrupt and backup unavailable: %w",err)};if err:=copyFileAtomic(bak,path);err!=nil{return false,err};return true,nil}
func backupJSONL(path string)error{if err:=validateJSONL(path);err!=nil{return err};return copyFileAtomic(path,path+".bak")}

func appendJSONL(path string,v any)error{
	b,err:=json.Marshal(v);if err!=nil{return err}
	f,err:=os.OpenFile(path,os.O_CREATE|os.O_WRONLY|os.O_APPEND,0o600);if err!=nil{return err};defer f.Close()
	if _,err=f.Write(append(b,'\n'));err!=nil{return err};if err:=f.Sync();err!=nil{return err};return backupJSONL(path)
}
func normalizeText(s string)string{return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(s)))," ")}
func idFor(s string)string{h:=sha256.Sum256([]byte(normalizeText(s)));return hex.EncodeToString(h[:12])}
func (s *Store)SaveMemory(text string)(Memory,bool,error){
	return s.SaveMemoryWithEmbedding(context.Background(), text, nil)
}
func tokens(s string)[]string{
	f:=strings.FieldsFunc(strings.ToLower(s),func(r rune)bool{return !(r>='a'&&r<='z'||r>='0'&&r<='9'||r>='À'&&r<='ÿ')})
	seen:=map[string]bool{};out:=[]string{};for _,t:=range f{if len(t)<2||seen[t]{continue};seen[t]=true;out=append(out,t)};return out
}
func (s *Store)SearchMemory(q string,limit int)[]Memory{
	return s.SearchMemoryHybrid(context.Background(), q, limit, nil)
}
func (s *Store)AddReceipt(r Receipt)error{s.mu.Lock();defer s.mu.Unlock();if err:=appendJSONL(filepath.Join(s.dir,"receipts.jsonl"),r);err!=nil{return err};s.receipts=append(s.receipts,r);return nil}
func (s *Store)RecentReceipts(limit int)[]Receipt{if limit<=0||limit>100{limit=20};s.mu.Lock();defer s.mu.Unlock();n:=len(s.receipts);start:=n-limit;if start<0{start=0};out:=append([]Receipt(nil),s.receipts[start:]...);for i,j:=0,len(out)-1;i<j;i,j=i+1,j-1{out[i],out[j]=out[j],out[i]};return out}
func (s *Store)Counts()(int,int){s.mu.Lock();defer s.mu.Unlock();return len(s.memories),len(s.receipts)}

type ToolRegistry struct{workspace string;gate *CapabilityGate;writer *WorkspaceWriter}
type ToolResult struct{Tool string `json:"tool"`;Verified bool `json:"verified"`;Data any `json:"data,omitempty"`;Error string `json:"error,omitempty"`}
func NewToolRegistry(root string)(*ToolRegistry,error){abs,err:=filepath.Abs(root);if err!=nil{return nil,err};real,err:=filepath.EvalSymlinks(abs);if err==nil{abs=real};gate,err:=NewCapabilityGate();if err!=nil{return nil,err};writer,err:=NewWorkspaceWriter(abs,gate);if err!=nil{return nil,err};return &ToolRegistry{workspace:abs,gate:gate,writer:writer},nil}
func (r *ToolRegistry)resolve(user string)(string,error){if user==""{user="."};var p string;if filepath.IsAbs(user){p=filepath.Clean(user)}else{p=filepath.Join(r.workspace,user)};abs,err:=filepath.Abs(p);if err!=nil{return "",err};real,err:=filepath.EvalSymlinks(abs);if err==nil{abs=real};rel,err:=filepath.Rel(r.workspace,abs);if err!=nil{return "",err};if rel==".."||strings.HasPrefix(rel,".."+string(os.PathSeparator)){return "",errors.New("path escapes workspace")};return abs,nil}
func sensitivePath(path string)bool{name:=strings.ToLower(filepath.Base(path));if name==".env"||name==".env.local"||name=="credentials"||name=="credentials.json"||name=="secrets.json"||name=="id_rsa"||name=="id_ed25519"{return true};ext:=strings.ToLower(filepath.Ext(name));return ext==".pem"||ext==".key"||ext==".p12"||ext==".pfx"}
func (r *ToolRegistry)List()[]string{out:=[]string{"system.info","fs.list","fs.read","fs.write.workspace"};if runtime.GOOS=="windows"{out=append(out,"windows.processes")};return out}
func (r *ToolRegistry)Run(ctx context.Context,name string,input map[string]string,passport CapabilityPassport,purpose string)ToolResult{
	required,ok:=capabilityForTool(name);if !ok{return ToolResult{Tool:name,Error:"tool not allowed"}}
	if r.gate==nil{return ToolResult{Tool:name,Error:"capability gate unavailable"}}
	if err:=r.gate.Authorize(passport,required,purpose);err!=nil{return ToolResult{Tool:name,Error:"capability denied: "+err.Error()}}
	switch name{
	case "system.info":return ToolResult{Tool:name,Verified:true,Data:map[string]any{"os":runtime.GOOS,"arch":runtime.GOARCH,"go":runtime.Version(),"workspace":r.workspace}}
	case "fs.list":
		p,err:=r.resolve(input["path"]);if err!=nil{return ToolResult{Tool:name,Error:err.Error()}};ents,err:=os.ReadDir(p);if err!=nil{return ToolResult{Tool:name,Error:err.Error()}};if len(ents)>200{ents=ents[:200]};out:=make([]map[string]any,0,len(ents));for _,e:=range ents{info,_:=e.Info();item:=map[string]any{"name":e.Name(),"dir":e.IsDir()};if info!=nil{item["size"]=info.Size()};out=append(out,item)};return ToolResult{Tool:name,Verified:true,Data:out}
	case "windows.processes":
		if runtime.GOOS!="windows"{return ToolResult{Tool:name,Error:"windows tool unavailable on this OS"}};root:=os.Getenv("SystemRoot");if root==""||!filepath.IsAbs(root){return ToolResult{Tool:name,Error:"invalid SystemRoot"}};exe:=filepath.Join(root,"System32","tasklist.exe");ctx2,cancel:=context.WithTimeout(ctx,5*time.Second);defer cancel();cmd:=exec.CommandContext(ctx2,exe,"/FO","CSV","/NH");var out limitedBuffer;out.limit=512*1024;cmd.Stdout=&out;if err:=cmd.Run();err!=nil{if ctx2.Err()!=nil{return ToolResult{Tool:name,Error:"windows observe timeout"}};return ToolResult{Tool:name,Error:err.Error()}};if out.exceeded{return ToolResult{Tool:name,Error:"windows observe output too large"}};return ToolResult{Tool:name,Verified:true,Data:string(out.b)}
	case "fs.read":
		p,err:=r.resolve(input["path"]);if err!=nil{return ToolResult{Tool:name,Error:err.Error()}};if sensitivePath(p){return ToolResult{Tool:name,Error:"sensitive file blocked by secret firewall"}};f,err:=os.Open(p);if err!=nil{return ToolResult{Tool:name,Error:err.Error()}};defer f.Close();b,err:=io.ReadAll(io.LimitReader(f,256*1024+1));if err!=nil{return ToolResult{Tool:name,Error:err.Error()}};if len(b)>256*1024{return ToolResult{Tool:name,Error:"file exceeds 256 KiB read limit"}};return ToolResult{Tool:name,Verified:true,Data:string(b)}
	case "fs.write.workspace":
		if r.writer==nil{return ToolResult{Tool:name,Error:"workspace writer unavailable"}};wr,err:=r.writer.Write(input["path"],[]byte(input["content"]),passport,purpose);if err!=nil{return ToolResult{Tool:name,Error:err.Error()}};return ToolResult{Tool:name,Verified:wr.Verified,Data:wr}
	default:return ToolResult{Tool:name,Error:"tool not allowed"}
	}
}

type PlanStep struct{Tool string `json:"tool"`;Input map[string]string `json:"input"`}
type Plan struct{Summary string `json:"summary"`;Steps []PlanStep `json:"steps"`}
type ModelAdapter struct{Executable string;Args []string;Timeout time.Duration;OllamaURL string;OllamaModel string}
func loopbackHTTP(raw string)bool{u,err:=url.Parse(raw);if err!=nil||u.Scheme!="http"{return false};host:=u.Hostname();ip:=net.ParseIP(host);return (host=="localhost"||(ip!=nil&&ip.IsLoopback()))&&u.Path==""}
func (m ModelAdapter)Available()bool{if m.OllamaModel!=""&&loopbackHTTP(m.OllamaURL){return true};if m.Executable==""||!filepath.IsAbs(m.Executable){return false};st,err:=os.Stat(m.Executable);return err==nil&&!st.IsDir()}
func validatePlan(p Plan,allowed []string)(Plan,error){if len(p.Steps)>8{return Plan{},errors.New("model plan exceeds 8 steps")};a:=map[string]bool{};for _,x:=range allowed{a[x]=true};for _,s:=range p.Steps{if !a[s.Tool]{return Plan{},fmt.Errorf("model requested disallowed tool %q",s.Tool)}};return p,nil}
func (m ModelAdapter)planOllama(ctx context.Context,goal string,allowed []string)(Plan,error){
	if m.OllamaModel==""||!loopbackHTTP(m.OllamaURL){return Plan{},errors.New("ollama not configured on loopback")}
	prompt:=fmt.Sprintf("Return ONLY JSON with schema {\"summary\":string,\"steps\":[{\"tool\":string,\"input\":object}]}. Allowed tools: %s. Goal: %s",strings.Join(allowed,", "),goal)
	body,_:=json.Marshal(map[string]any{"model":m.OllamaModel,"prompt":prompt,"stream":false,"format":"json"});req,err:=http.NewRequestWithContext(ctx,http.MethodPost,strings.TrimRight(m.OllamaURL,"/")+"/api/generate",strings.NewReader(string(body)));if err!=nil{return Plan{},err};req.Header.Set("Content-Type","application/json")
	client:=&http.Client{Timeout:m.Timeout};resp,err:=client.Do(req);if err!=nil{return Plan{},err};defer resp.Body.Close();if resp.StatusCode!=200{return Plan{},fmt.Errorf("ollama http %d",resp.StatusCode)}
	b,err:=io.ReadAll(io.LimitReader(resp.Body,1024*1024+1));if err!=nil{return Plan{},err};if len(b)>1024*1024{return Plan{},errors.New("ollama output too large")};var env struct{Response string `json:"response"`};if err:=json.Unmarshal(b,&env);err!=nil{return Plan{},err};var p Plan;if err:=json.Unmarshal([]byte(env.Response),&p);err!=nil{return Plan{},fmt.Errorf("invalid ollama plan json: %w",err)};return validatePlan(p,allowed)
}
func (m ModelAdapter)Plan(ctx context.Context,goal string,allowed []string)(Plan,error){
	if !m.Available(){return Plan{},errors.New("local model not configured")};timeout:=m.Timeout;if timeout<=0||timeout>60*time.Second{timeout=20*time.Second};ctx,cancel:=context.WithTimeout(ctx,timeout);defer cancel();if m.OllamaModel!=""{return m.planOllama(ctx,goal,allowed)}
	req,_:=json.Marshal(map[string]any{"goal":goal,"allowed_tools":allowed,"format":"json_plan_v1"});cmd:=exec.CommandContext(ctx,m.Executable,m.Args...);cmd.Stdin=strings.NewReader(string(req));var out limitedBuffer;out.limit=1024*1024;cmd.Stdout=&out;if err:=cmd.Run();err!=nil{if ctx.Err()!=nil{return Plan{},fmt.Errorf("model timeout/cancel: %w",ctx.Err())};return Plan{},err};if out.exceeded{return Plan{},errors.New("model output too large")};var p Plan;if err:=json.Unmarshal(out.b,&p);err!=nil{return Plan{},fmt.Errorf("invalid model json: %w",err)};return validatePlan(p,allowed)
}
type limitedBuffer struct{b []byte;limit int;exceeded bool}
func(l *limitedBuffer)Write(p []byte)(int,error){n:=len(p);remaining:=l.limit-len(l.b);if remaining<=0{l.exceeded=true;return n,nil};if n>remaining{l.b=append(l.b,p[:remaining]...);l.exceeded=true}else{l.b=append(l.b,p...)};return n,nil}

type CommandResult struct{Goal string `json:"goal"`;Status string `json:"status"`;Plan *Plan `json:"plan,omitempty"`;Results []ToolResult `json:"results,omitempty"`;Memory []Memory `json:"memory,omitempty"`;Message string `json:"message,omitempty"`;ReceiptIDs []string `json:"receipt_ids,omitempty"`}
type Core struct{store *Store;tools *ToolRegistry;model ModelAdapter;embedding LocalEmbeddingProvider;recorder *FlightRecorder}
func(c *Core)runTool(ctx context.Context,goal,subject,name string,input map[string]string)ToolResult{
	if c==nil||c.tools==nil||c.tools.gate==nil{return ToolResult{Tool:name,Error:"capability gate unavailable"}}
	required,ok:=capabilityForTool(name);if !ok{return ToolResult{Tool:name,Error:"tool not allowed"}}
	passport,err:=c.tools.gate.Issue(subject,goal,[]string{required},45*time.Second);if err!=nil{return ToolResult{Tool:name,Error:"capability issue failed: "+err.Error()}}
	return c.tools.Run(ctx,name,input,passport,goal)
}
func receiptID(action string,t time.Time)string{return idFor(action+t.UTC().Format(time.RFC3339Nano))}
func(c *Core)receipt(goal,action,status,detail,errText string,start time.Time)string{
	r:=Receipt{ID:receiptID(action,start),Goal:goal,Action:action,Status:status,Detail:detail,Error:errText,StartedAt:start,FinishedAt:time.Now().UTC()}
	_=c.store.AddReceipt(r)
	if c.recorder!=nil {
		var cause error
		if strings.TrimSpace(errText)!="" { cause=errors.New(errText) }
		_,_=c.recorder.Record("",goal,action,"receipt",status,cause)
	}
	return r.ID
}
func prefixValue(goal string,prefixes ...string)(string,bool){g:=strings.TrimSpace(goal);low:=strings.ToLower(g);for _,p:=range prefixes{if strings.HasPrefix(low,p){return strings.TrimSpace(g[len(p):]),true}};return "",false}
func parseWriteGoal(goal string)(string,string,bool){g:=strings.TrimSpace(goal);low:=strings.ToLower(g);for _,p:=range []string{"scrivi file ","write file "}{if strings.HasPrefix(low,p){rest:=strings.TrimSpace(g[len(p):]);parts:=strings.SplitN(rest," :: ",2);if len(parts)!=2||strings.TrimSpace(parts[0])==""{return "","",false};return strings.TrimSpace(parts[0]),parts[1],true}};return "","",false}
func(c *Core)Execute(ctx context.Context,goal string)CommandResult{
	goal=strings.TrimSpace(goal);res:=CommandResult{Goal:goal};if goal==""{res.Status="blocked";res.Message="empty goal";return res};if len([]byte(goal))>64*1024{res.Status="blocked";res.Message="goal too large";return res}
	if v,ok:=prefixValue(goal,"ricorda ","remember ");ok{st:=time.Now().UTC();m,created,err:=c.store.SaveGovernedMemory(ctx,v,MemoryWriteMeta{Type:"fact",Provenance:"user-direct"},c.embedding);if err!=nil{res.Status="error";res.Message=err.Error();res.ReceiptIDs=[]string{c.receipt(goal,"memory.save","error","",err.Error(),st)};return res};res.Status="ok";res.Memory=[]Memory{m};if created{res.Message="memory saved"}else{res.Message="duplicate memory already present"};res.ReceiptIDs=[]string{c.receipt(goal,"memory.save","verified",res.Message,"",st)};return res}
	if v,ok:=prefixValue(goal,"cerca memoria ","search memory ","ricorda su ");ok{st:=time.Now().UTC();res.Memory=c.store.SearchMemoryHybrid(ctx,v,10,c.embedding);res.Status="ok";res.Message=fmt.Sprintf("%d memories found",len(res.Memory));res.ReceiptIDs=[]string{c.receipt(goal,"memory.search","verified",res.Message,"",st)};return res}
	low:=strings.ToLower(goal);if low=="stato"||low=="status"||low=="stato sistema"{st:=time.Now().UTC();tr:=c.runTool(ctx,goal,"core-direct","system.info",nil);res.Status="ok";res.Results=[]ToolResult{tr};res.ReceiptIDs=[]string{c.receipt(goal,"system.info","verified","system info returned","",st)};return res}
	if v,ok:=prefixValue(goal,"lista file ","list files ");ok{st:=time.Now().UTC();tr:=c.runTool(ctx,goal,"core-direct","fs.list",map[string]string{"path":v});res.Results=[]ToolResult{tr};if tr.Error!=""{res.Status="blocked";res.Message=tr.Error;res.ReceiptIDs=[]string{c.receipt(goal,"fs.list","blocked","",tr.Error,st)}}else{res.Status="ok";res.Message="directory listed";res.ReceiptIDs=[]string{c.receipt(goal,"fs.list","verified",res.Message,"",st)}};return res}
	if v,ok:=prefixValue(goal,"leggi file ","read file ");ok{st:=time.Now().UTC();tr:=c.runTool(ctx,goal,"core-direct","fs.read",map[string]string{"path":v});res.Results=[]ToolResult{tr};if tr.Error!=""{res.Status="blocked";res.Message=tr.Error;res.ReceiptIDs=[]string{c.receipt(goal,"fs.read","blocked","",tr.Error,st)}}else{res.Status="ok";res.Message="file read and verified";res.ReceiptIDs=[]string{c.receipt(goal,"fs.read","verified",res.Message,"",st)}};return res}
	if path,content,ok:=parseWriteGoal(goal);ok{st:=time.Now().UTC();tr:=c.runTool(ctx,goal,"core-direct","fs.write.workspace",map[string]string{"path":path,"content":content});res.Results=[]ToolResult{tr};if tr.Error!=""{res.Status="blocked";res.Message=tr.Error;res.ReceiptIDs=[]string{c.receipt(goal,"fs.write.workspace","blocked","",tr.Error,st)}}else{res.Status="ok";res.Message="workspace write verified";res.ReceiptIDs=[]string{c.receipt(goal,"fs.write.workspace","verified",res.Message,"",st)}};return res}
	if low=="processi windows"||low=="windows processes"{st:=time.Now().UTC();tr:=c.runTool(ctx,goal,"core-direct","windows.processes",nil);res.Results=[]ToolResult{tr};if tr.Error!=""{res.Status="unavailable";res.Message=tr.Error;res.ReceiptIDs=[]string{c.receipt(goal,"windows.processes","blocked","",tr.Error,st)}}else{res.Status="ok";res.Message="windows process observation verified";res.ReceiptIDs=[]string{c.receipt(goal,"windows.processes","verified",res.Message,"",st)}};return res}
	p,err:=c.model.Plan(ctx,goal,c.tools.List());if err!=nil{res.Status="unavailable";res.Message="goal requires local model; "+err.Error();return res}
	return c.ExecutePlan(ctx,goal,p)
}
func loopbackRemote(addr string)bool{host,_,err:=net.SplitHostPort(addr);if err!=nil{return false};ip:=net.ParseIP(host);return ip!=nil&&ip.IsLoopback()}
