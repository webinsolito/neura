package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type ImprovementChange struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type ImprovementEvaluation struct {
	TestsPassed       bool      `json:"tests_passed"`
	BenchmarkMeasured bool      `json:"benchmark_measured"`
	BenchmarkNoWorse  bool      `json:"benchmark_no_worse"`
	Evidence          []string  `json:"evidence,omitempty"`
	EvaluatedAt       time.Time `json:"evaluated_at,omitempty"`
}

type ImprovementCandidate struct {
	ID                 string                `json:"id"`
	MacroArea          string                `json:"macro_area"`
	BaseFingerprint    string                `json:"base_fingerprint"`
	ChangeCount        int                   `json:"change_count"`
	RollbackReady      bool                  `json:"rollback_ready"`
	State              string                `json:"state"`
	CreatedAt          time.Time             `json:"created_at"`
	Evaluation         ImprovementEvaluation `json:"evaluation"`
	PromotionReady     bool                  `json:"promotion_ready"`
	AutoPromotion      bool                  `json:"auto_promotion"`
}

type ImprovementManager struct {
	workspace string
	root      string
}

func NewImprovementManager(workspace string)(*ImprovementManager,error){
	abs,err:=filepath.Abs(workspace);if err!=nil{return nil,err}
	real,err:=filepath.EvalSymlinks(abs);if err==nil{abs=real}
	root:=filepath.Join(abs,".neura","improvements")
	if err:=os.MkdirAll(root,0o700);err!=nil{return nil,err}
	return &ImprovementManager{workspace:abs,root:root},nil
}

func (m *ImprovementManager) changePath(rel string)(string,error){
	rel=strings.TrimSpace(rel)
	if rel==""||filepath.IsAbs(rel){return "",errors.New("improvement path must be relative")}
	clean:=filepath.Clean(rel)
	if clean=="."||clean==".."||strings.HasPrefix(clean,".."+string(os.PathSeparator)){return "",errors.New("improvement path escapes workspace")}
	if clean==".git"||strings.HasPrefix(clean,".git"+string(os.PathSeparator))||clean==".neura"||strings.HasPrefix(clean,".neura"+string(os.PathSeparator)){return "",errors.New("internal state or git metadata cannot be patched")}
	p:=filepath.Join(m.workspace,clean)
	if sensitivePath(p){return "",errors.New("sensitive path blocked")}
	return p,nil
}

func improvementFingerprint(items map[string][]byte) string {
	keys:=make([]string,0,len(items));for k:=range items{keys=append(keys,k)};sort.Strings(keys)
	var b strings.Builder
	for _,k:=range keys{b.WriteString(k);b.WriteString(":");b.WriteString(digestBytes(items[k]));b.WriteString("\n")}
	return digestBytes([]byte(b.String()))
}

func (m *ImprovementManager) Stage(id,macro string,changes []ImprovementChange)(ImprovementCandidate,error){
	id=strings.TrimSpace(id);macro=strings.TrimSpace(macro)
	if id==""||macro==""{return ImprovementCandidate{},errors.New("candidate id and one macro area are required")}
	if strings.ContainsAny(id,"/\\"){return ImprovementCandidate{},errors.New("candidate id must be a simple name")}
	if len(changes)==0||len(changes)>8{return ImprovementCandidate{},errors.New("candidate must contain 1..8 changed files")}
	dir:=filepath.Join(m.root,id)
	if _,err:=os.Stat(dir);err==nil{return ImprovementCandidate{},errors.New("candidate already exists")}
	base:=map[string][]byte{};total:=0
	for _,ch:=range changes{
		p,err:=m.changePath(ch.Path);if err!=nil{return ImprovementCandidate{},err}
		if memorySecretBlocked(ch.Content){return ImprovementCandidate{},fmt.Errorf("%s: secret firewall blocked candidate content",ch.Path)}
		if len([]byte(ch.Content))>64*1024{return ImprovementCandidate{},fmt.Errorf("%s: change too large",ch.Path)}
		total+=len([]byte(ch.Content));if total>256*1024{return ImprovementCandidate{},errors.New("candidate aggregate size exceeds 256 KiB")}
		b,err:=os.ReadFile(p);if errors.Is(err,os.ErrNotExist){b=nil}else if err!=nil{return ImprovementCandidate{},err}
		base[filepath.Clean(ch.Path)]=b
	}
	if err:=os.MkdirAll(filepath.Join(dir,"candidate"),0o700);err!=nil{return ImprovementCandidate{},err}
	if err:=os.MkdirAll(filepath.Join(dir,"rollback"),0o700);err!=nil{return ImprovementCandidate{},err}
	for _,ch:=range changes{
		rel:=filepath.Clean(ch.Path)
		cp:=filepath.Join(dir,"candidate",rel);if err:=os.MkdirAll(filepath.Dir(cp),0o700);err!=nil{return ImprovementCandidate{},err}
		if err:=os.WriteFile(cp,[]byte(ch.Content),0o600);err!=nil{return ImprovementCandidate{},err}
		rp:=filepath.Join(dir,"rollback",rel);if err:=os.MkdirAll(filepath.Dir(rp),0o700);err!=nil{return ImprovementCandidate{},err}
		if b:=base[rel];b!=nil{if err:=os.WriteFile(rp,b,0o600);err!=nil{return ImprovementCandidate{},err}}
	}
	c:=ImprovementCandidate{ID:id,MacroArea:macro,BaseFingerprint:improvementFingerprint(base),ChangeCount:len(changes),RollbackReady:true,State:"staged",CreatedAt:time.Now().UTC(),AutoPromotion:false}
	if err:=m.writeManifest(c);err!=nil{return ImprovementCandidate{},err}
	return c,nil
}

func (m *ImprovementManager) manifestPath(id string)string{return filepath.Join(m.root,id,"manifest.json")}
func (m *ImprovementManager) writeManifest(c ImprovementCandidate)error{
	b,err:=json.MarshalIndent(c,"","  ");if err!=nil{return err}
	return writeAtomic(m.manifestPath(c.ID),b)
}
func (m *ImprovementManager) Get(id string)(ImprovementCandidate,error){
	b,err:=os.ReadFile(m.manifestPath(id));if err!=nil{return ImprovementCandidate{},err}
	var c ImprovementCandidate;if err:=json.Unmarshal(b,&c);err!=nil{return ImprovementCandidate{},err};return c,nil
}

func (m *ImprovementManager) RecordEvaluation(id string,e ImprovementEvaluation)(ImprovementCandidate,error){
	c,err:=m.Get(id);if err!=nil{return ImprovementCandidate{},err}
	if c.State!="staged"&&c.State!="evaluated"{return ImprovementCandidate{},errors.New("candidate not evaluable")}
	e.EvaluatedAt=time.Now().UTC()
	c.Evaluation=e;c.State="evaluated"
	c.PromotionReady=c.RollbackReady&&e.TestsPassed&&e.BenchmarkMeasured&&e.BenchmarkNoWorse
	if err:=m.writeManifest(c);err!=nil{return ImprovementCandidate{},err}
	return c,nil
}

func (m *ImprovementManager) PromotionDecision(id string)(bool,string,error){
	c,err:=m.Get(id);if err!=nil{return false,"",err}
	if !c.RollbackReady{return false,"rollback evidence missing",nil}
	if !c.Evaluation.TestsPassed{return false,"tests are not green",nil}
	if !c.Evaluation.BenchmarkMeasured{return false,"benchmark evidence missing",nil}
	if !c.Evaluation.BenchmarkNoWorse{return false,"benchmark regression detected",nil}
	if c.AutoPromotion{return false,"automatic promotion is forbidden",nil}
	return true,"candidate meets promotion gate; explicit operator promotion still required",nil
}
