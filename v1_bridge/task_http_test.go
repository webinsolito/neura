package main

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestTaskHTTPCreateAndAdvance(t *testing.T) {
	s:=testServer(t)
	body:=map[string]any{"id":"http-task","goal":"write and read","plan":Plan{Steps:[]PlanStep{
		{Tool:"fs.write.workspace",Input:map[string]string{"path":"http.txt","content":"hello"}},
		{Tool:"fs.read",Input:map[string]string{"path":"http.txt"}},
	}}}
	rr:=doReq(t,s,http.MethodPost,"/tasks",body,"")
	if rr.Code!=http.StatusCreated{t.Fatalf("create code=%d body=%s",rr.Code,rr.Body.String())}
	var task DurableTask;if err:=json.Unmarshal(rr.Body.Bytes(),&task);err!=nil{t.Fatal(err)}
	rr=doReq(t,s,http.MethodPost,"/tasks/http-task/next",map[string]any{"generation":task.Generation},"")
	if rr.Code!=http.StatusOK{t.Fatalf("next1 code=%d body=%s",rr.Code,rr.Body.String())}
	rr=doReq(t,s,http.MethodPost,"/tasks/http-task/next",map[string]any{"generation":task.Generation},"")
	if rr.Code!=http.StatusOK{t.Fatalf("next2 code=%d body=%s",rr.Code,rr.Body.String())}
	rr=doReq(t,s,http.MethodGet,"/tasks/http-task",nil,"")
	if rr.Code!=http.StatusOK{t.Fatalf("get code=%d body=%s",rr.Code,rr.Body.String())}
	var final DurableTask;if err:=json.Unmarshal(rr.Body.Bytes(),&final);err!=nil{t.Fatal(err)}
	if final.State!=TaskSucceeded||final.NextStep!=2{t.Fatalf("%+v",final)}
}
