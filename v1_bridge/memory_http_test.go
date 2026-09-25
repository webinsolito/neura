package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestMemoryHTTPCurrentTruthAndHistory(t *testing.T) {
	s:=testServer(t)
	first,_,err:=s.store.SaveGovernedMemory(context.Background(),"desk is in room one",MemoryWriteMeta{Type:"fact",Provenance:"test"},nil);if err!=nil{t.Fatal(err)}
	second,_,err:=s.store.SaveGovernedMemory(context.Background(),"desk is in room two",MemoryWriteMeta{Type:"fact",Provenance:"test",Supersedes:first.ID},nil);if err!=nil{t.Fatal(err)}
	rr:=doReq(t,s,http.MethodGet,"/memory?q=desk%20room&limit=10",nil,"");if rr.Code!=http.StatusOK{t.Fatalf("search=%d %s",rr.Code,rr.Body.String())}
	var found struct{Items []Memory `json:"items"`};if err:=json.Unmarshal(rr.Body.Bytes(),&found);err!=nil{t.Fatal(err)}
	if len(found.Items)!=1||found.Items[0].ID!=second.ID{t.Fatalf("current truth=%+v",found.Items)}
	rr=doReq(t,s,http.MethodGet,"/memory/history?id="+second.ID,nil,"");if rr.Code!=http.StatusOK{t.Fatalf("history=%d %s",rr.Code,rr.Body.String())}
	var history struct{Items []Memory `json:"items"`};if err:=json.Unmarshal(rr.Body.Bytes(),&history);err!=nil{t.Fatal(err)}
	if len(history.Items)!=2||history.Items[0].ID!=second.ID||history.Items[1].ID!=first.ID{t.Fatalf("history=%+v",history.Items)}
}

func TestMemoryHTTPRejectsForkConflict(t *testing.T) {
	s:=testServer(t)
	rr:=doReq(t,s,http.MethodPost,"/memory",map[string]any{"text":"office is in room one","type":"fact","provenance":"test"},"");if rr.Code!=http.StatusOK{t.Fatalf("first=%d %s",rr.Code,rr.Body.String())}
	var first struct{Memory Memory `json:"memory"`};if err:=json.Unmarshal(rr.Body.Bytes(),&first);err!=nil{t.Fatal(err)}
	rr=doReq(t,s,http.MethodPost,"/memory",map[string]any{"text":"office is in room two","type":"fact","provenance":"test","supersedes":first.Memory.ID},"");if rr.Code!=http.StatusOK{t.Fatalf("second=%d %s",rr.Code,rr.Body.String())}
	rr=doReq(t,s,http.MethodPost,"/memory",map[string]any{"text":"office is in room three","type":"fact","provenance":"test","supersedes":first.Memory.ID},"");if rr.Code!=http.StatusConflict{t.Fatalf("fork=%d %s",rr.Code,rr.Body.String())}
}

func TestMemoryHistoryRequiresID(t *testing.T) {
	s:=testServer(t);rr:=doReq(t,s,http.MethodGet,"/memory/history",nil,"")
	if rr.Code!=http.StatusBadRequest{t.Fatalf("code=%d body=%s",rr.Code,rr.Body.String())}
}
