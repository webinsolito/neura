package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type TaskState string

const (
	TaskPending      TaskState = "pending"
	TaskRunning      TaskState = "running"
	TaskSucceeded    TaskState = "succeeded"
	TaskBlocked      TaskState = "blocked"
	TaskManualReview TaskState = "manual_review"
	TaskCancelled    TaskState = "cancelled"
)

type DurableTask struct {
	ID          string    `json:"id"`
	GoalHash    string    `json:"goal_hash"`
	Plan        Plan      `json:"plan"`
	NextStep    int       `json:"next_step"`
	RunningStep int       `json:"running_step"`
	Generation  uint64    `json:"generation"`
	State       TaskState `json:"state"`
	LastError   string    `json:"last_error,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TaskStore struct {
	mu    sync.Mutex
	path  string
	tasks map[string]DurableTask
}

func OpenTaskStore(workspace string) (*TaskStore,error) {
	abs,err:=filepath.Abs(workspace);if err!=nil{return nil,err}
	s:=&TaskStore{path:filepath.Join(abs,".neura","tasks.json"),tasks:map[string]DurableTask{}}
	b,err:=os.ReadFile(s.path);if errors.Is(err,os.ErrNotExist){return s,nil};if err!=nil{return nil,err}
	if len(b)>0 { if err:=json.Unmarshal(b,&s.tasks);err!=nil{return nil,fmt.Errorf("task state corrupt: %w",err)} }
	return s,nil
}

func (s *TaskStore) persistLocked() error {
	if err:=os.MkdirAll(filepath.Dir(s.path),0o700);err!=nil{return err}
	b,err:=json.MarshalIndent(s.tasks,"","  ");if err!=nil{return err}
	tmp:=s.path+".tmp"
	if err:=os.WriteFile(tmp,b,0o600);err!=nil{return err}
	return os.Rename(tmp,s.path)
}

func taskPlanPersistable(p Plan) error {
	for _,step:=range p.Steps {
		for k,v:=range step.Input {
			if memorySecretBlocked(v) { return fmt.Errorf("task input %q blocked by secret firewall",k) }
		}
	}
	return nil
}

func (s *TaskStore) Create(id,goal string,p Plan,allowed []string)(DurableTask,error){
	s.mu.Lock();defer s.mu.Unlock()
	id=strings.TrimSpace(id);if id==""{return DurableTask{},errors.New("task id required")}
	if _,ok:=s.tasks[id];ok{return DurableTask{},errors.New("task id already exists")}
	checked,err:=validatePlan(p,allowed);if err!=nil{return DurableTask{},err}
	if err:=taskPlanPersistable(checked);err!=nil{return DurableTask{},err}
	t:=DurableTask{ID:id,GoalHash:flightGoalHash(goal),Plan:checked,Generation:1,State:TaskPending,UpdatedAt:time.Now().UTC(),RunningStep:-1}
	s.tasks[id]=t
	return t,s.persistLocked()
}

func (s *TaskStore) Get(id string)(DurableTask,bool){s.mu.Lock();defer s.mu.Unlock();t,ok:=s.tasks[id];return t,ok}

func isMutatingTool(name string) bool {
	switch name {
	case "fs.write.workspace","fs.mkdir.workspace","fs.move.workspace","fs.delete.workspace","fs.restore.workspace","windows.app.launch":
		return true
	default:
		return false
	}
}

type taskCompleteError struct{}
func (taskCompleteError) Error() string{return "task complete"}

func (s *TaskStore) beginStep(id string,generation uint64)(DurableTask,PlanStep,error){
	s.mu.Lock();defer s.mu.Unlock()
	t,ok:=s.tasks[id];if !ok{return DurableTask{},PlanStep{},os.ErrNotExist}
	if t.Generation!=generation{return DurableTask{},PlanStep{},errors.New("stale task generation")}
	if t.State==TaskSucceeded||t.State==TaskCancelled||t.State==TaskManualReview{return DurableTask{},PlanStep{},fmt.Errorf("task not runnable: %s",t.State)}
	if t.NextStep>=len(t.Plan.Steps){t.State=TaskSucceeded;t.UpdatedAt=time.Now().UTC();s.tasks[id]=t;_ = s.persistLocked();return t,PlanStep{},taskCompleteError{}}
	t.State=TaskRunning;t.RunningStep=t.NextStep;t.UpdatedAt=time.Now().UTC();s.tasks[id]=t
	if err:=s.persistLocked();err!=nil{return DurableTask{},PlanStep{},err}
	return t,t.Plan.Steps[t.NextStep],nil
}

func (s *TaskStore) finishStep(id string,generation uint64,ok bool,errText string)(DurableTask,error){
	s.mu.Lock();defer s.mu.Unlock()
	t,exists:=s.tasks[id];if !exists{return DurableTask{},os.ErrNotExist}
	if t.Generation!=generation{return DurableTask{},errors.New("stale task generation")}
	if t.State!=TaskRunning{return DurableTask{},errors.New("task not running")}
	if ok {
		t.NextStep++
		t.RunningStep=-1
		t.LastError=""
		if t.NextStep>=len(t.Plan.Steps){t.State=TaskSucceeded}else{t.State=TaskPending}
	} else {
		t.State=TaskBlocked
		t.LastError=errText
		t.RunningStep=-1
	}
	t.UpdatedAt=time.Now().UTC();s.tasks[id]=t
	return t,s.persistLocked()
}

func (s *TaskStore) RecoverInterrupted() error {
	s.mu.Lock();defer s.mu.Unlock()
	changed:=false
	for id,t:=range s.tasks {
		if t.State!=TaskRunning {continue}
		t.Generation++
		if t.RunningStep>=0 && t.RunningStep<len(t.Plan.Steps) && isMutatingTool(t.Plan.Steps[t.RunningStep].Tool) {
			t.State=TaskManualReview
			t.LastError="interrupted during mutating step; effect must be verified before resume"
		} else {
			t.State=TaskPending
			t.LastError="interrupted during read-only step; safe to retry"
		}
		t.RunningStep=-1;t.UpdatedAt=time.Now().UTC();s.tasks[id]=t;changed=true
	}
	if changed{return s.persistLocked()};return nil
}

func (s *TaskStore) ResolveManualReview(id string,generation uint64,effectVerified bool)(DurableTask,error){
	s.mu.Lock();defer s.mu.Unlock()
	t,ok:=s.tasks[id];if !ok{return DurableTask{},os.ErrNotExist}
	if t.Generation!=generation{return DurableTask{},errors.New("stale task generation")}
	if t.State!=TaskManualReview{return DurableTask{},errors.New("task is not awaiting manual review")}
	if effectVerified {t.NextStep++}
	t.State=TaskPending;t.LastError="";t.UpdatedAt=time.Now().UTC();s.tasks[id]=t
	return t,s.persistLocked()
}

type TaskRunner struct { core *Core; store *TaskStore }

func NewTaskRunner(core *Core,workspace string)(*TaskRunner,error){
	if core==nil||core.tools==nil{return nil,errors.New("core required")}
	s,err:=OpenTaskStore(workspace);if err!=nil{return nil,err}
	return &TaskRunner{core:core,store:s},nil
}

func (r *TaskRunner) RunNext(ctx context.Context,id string,generation uint64) (DurableTask,CommandResult,error) {
	t,step,err:=r.store.beginStep(id,generation)
	if err!=nil {
		if _,ok:=err.(taskCompleteError);ok{return t,CommandResult{Status:"ok",Message:"task complete"},nil}
		return DurableTask{},CommandResult{},err
	}
	res:=r.core.ExecutePlan(ctx,"durable-task:"+t.ID,Plan{Summary:t.Plan.Summary,Steps:[]PlanStep{step}})
	success:=res.Status=="ok"&&len(res.Results)==1&&res.Results[0].Verified
	errText:=res.Message
	if !success && errText=="" {errText="step failed verification"}
	updated,persistErr:=r.store.finishStep(id,generation,success,errText)
	if persistErr!=nil{return DurableTask{},res,persistErr}
	return updated,res,nil
}
