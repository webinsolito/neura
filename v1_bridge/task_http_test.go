package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
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


func TestTaskHTTPListAndRetryRecovery(t *testing.T) {
	s:=testServer(t)
	body:=map[string]any{"id":"recover-http","goal":"read after repair","plan":Plan{Steps:[]PlanStep{{Tool:"fs.read",Input:map[string]string{"path":"recovered.txt"}}}}}
	rr:=doReq(t,s,http.MethodPost,"/tasks",body,"");if rr.Code!=http.StatusCreated{t.Fatalf("create=%d %s",rr.Code,rr.Body.String())}
	var task DurableTask;if err:=json.Unmarshal(rr.Body.Bytes(),&task);err!=nil{t.Fatal(err)}
	rr=doReq(t,s,http.MethodPost,"/tasks/recover-http/next",map[string]any{"generation":task.Generation},"");if rr.Code!=http.StatusOK{t.Fatalf("next=%d %s",rr.Code,rr.Body.String())}
	var step struct{Task DurableTask `json:"task"`};if err:=json.Unmarshal(rr.Body.Bytes(),&step);err!=nil{t.Fatal(err)}
	if step.Task.State!=TaskBlocked{t.Fatalf("%+v",step.Task)}
	rr=doReq(t,s,http.MethodGet,"/tasks?limit=6",nil,"");if rr.Code!=http.StatusOK{t.Fatalf("list=%d %s",rr.Code,rr.Body.String())}
	var listed struct{Items []DurableTask `json:"items"`};if err:=json.Unmarshal(rr.Body.Bytes(),&listed);err!=nil{t.Fatal(err)}
	if len(listed.Items)!=1||listed.Items[0].ID!="recover-http"{t.Fatalf("%+v",listed.Items)}
	if err:=os.WriteFile(filepath.Join(s.tools.workspace,"recovered.txt"),[]byte("ok"),0o600);err!=nil{t.Fatal(err)}
	rr=doReq(t,s,http.MethodPost,"/tasks/recover-http/retry",map[string]any{"generation":step.Task.Generation},"");if rr.Code!=http.StatusOK{t.Fatalf("retry=%d %s",rr.Code,rr.Body.String())}
	var retried DurableTask;if err:=json.Unmarshal(rr.Body.Bytes(),&retried);err!=nil{t.Fatal(err)}
	if retried.State!=TaskPending||retried.Generation!=step.Task.Generation+1{t.Fatalf("%+v",retried)}
	rr=doReq(t,s,http.MethodPost,"/tasks/recover-http/next",map[string]any{"generation":retried.Generation},"");if rr.Code!=http.StatusOK{t.Fatalf("next2=%d %s",rr.Code,rr.Body.String())}
}

func TestTaskHTTPMutatingFailureUsesManualReview(t *testing.T) {
	s:=testServer(t)
	body:=map[string]any{"id":"review-http","goal":"unsafe write","plan":Plan{Steps:[]PlanStep{{Tool:"fs.write.workspace",Input:map[string]string{"path":"../escape.txt","content":"x"}}}}}
	rr:=doReq(t,s,http.MethodPost,"/tasks",body,"");if rr.Code!=http.StatusCreated{t.Fatalf("create=%d %s",rr.Code,rr.Body.String())}
	var task DurableTask;_ = json.Unmarshal(rr.Body.Bytes(),&task)
	rr=doReq(t,s,http.MethodPost,"/tasks/review-http/next",map[string]any{"generation":task.Generation},"");if rr.Code!=http.StatusOK{t.Fatalf("next=%d %s",rr.Code,rr.Body.String())}
	var step struct{Task DurableTask `json:"task"`};_ = json.Unmarshal(rr.Body.Bytes(),&step)
	if step.Task.State!=TaskManualReview||step.Task.Generation!=task.Generation+1{t.Fatalf("%+v",step.Task)}
	rr=doReq(t,s,http.MethodPost,"/tasks/review-http/retry",map[string]any{"generation":step.Task.Generation},"");if rr.Code!=http.StatusConflict{t.Fatalf("unsafe retry code=%d body=%s",rr.Code,rr.Body.String())}
}
