package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "sync"
    "time"
)

type JobState string

const (
    JobPending JobState = "pending"
    JobRunning JobState = "running"
    JobSucceeded JobState = "succeeded"
    JobFailed JobState = "failed"
    JobCancelled JobState = "cancelled"
)

type DurableJob struct {
    ID string `json:"id"`
    IdempotencyKey string `json:"idempotency_key"`
    Generation uint64 `json:"generation"`
    State JobState `json:"state"`
    Attempts int `json:"attempts"`
    MaxAttempts int `json:"max_attempts"`
    LastError string `json:"last_error,omitempty"`
    UpdatedAt time.Time `json:"updated_at"`
}

type JobStore struct { mu sync.Mutex; path string; jobs map[string]DurableJob }

func OpenJobStore(workspace string) (*JobStore, error) { root, err := filepath.Abs(workspace); if err != nil { return nil, err }; s := &JobStore{path:filepath.Join(root, ".neura", "jobs.json"), jobs:map[string]DurableJob{}}; if err := s.load(); err != nil { return nil, err }; return s, nil }
func (s *JobStore) load() error { b, err := os.ReadFile(s.path); if errors.Is(err, os.ErrNotExist) { return nil }; if err != nil { return err }; if len(b)==0 { return nil }; return json.Unmarshal(b, &s.jobs) }
func (s *JobStore) persistLocked() error { if err:=os.MkdirAll(filepath.Dir(s.path),0700); err!=nil{return err}; b,err:=json.MarshalIndent(s.jobs,"","  "); if err!=nil{return err}; tmp:=s.path+".tmp"; if err:=os.WriteFile(tmp,b,0600);err!=nil{return err}; return os.Rename(tmp,s.path) }
func (s *JobStore) Enqueue(id,key string,generation uint64,maxAttempts int)(DurableJob,error){s.mu.Lock();defer s.mu.Unlock();if id==""||key==""{return DurableJob{},errors.New("job id and idempotency key required")};if maxAttempts<1{maxAttempts=1};for _,j:=range s.jobs{if j.IdempotencyKey==key{return j,nil}};j:=DurableJob{ID:id,IdempotencyKey:key,Generation:generation,State:JobPending,MaxAttempts:maxAttempts,UpdatedAt:time.Now().UTC()};s.jobs[id]=j;return j,s.persistLocked()}
func (s *JobStore) Claim(id string,expectedGeneration uint64)(DurableJob,error){s.mu.Lock();defer s.mu.Unlock();j,ok:=s.jobs[id];if !ok{return DurableJob{},os.ErrNotExist};if j.Generation!=expectedGeneration{return DurableJob{},fmt.Errorf("stale generation: have %d want %d",expectedGeneration,j.Generation)};if j.State==JobCancelled||j.State==JobSucceeded{return DurableJob{},fmt.Errorf("job not claimable: %s",j.State)};if j.Attempts>=j.MaxAttempts{return DurableJob{},errors.New("retry budget exhausted")};j.State=JobRunning;j.Attempts++;j.UpdatedAt=time.Now().UTC();s.jobs[id]=j;return j,s.persistLocked()}
func (s *JobStore) Complete(id string,generation uint64,runErr error)(DurableJob,error){s.mu.Lock();defer s.mu.Unlock();j,ok:=s.jobs[id];if !ok{return DurableJob{},os.ErrNotExist};if j.Generation!=generation{return DurableJob{},errors.New("stale generation")};if j.State!=JobRunning{return DurableJob{},fmt.Errorf("job not running: %s",j.State)};if runErr==nil{j.State=JobSucceeded;j.LastError=""}else if j.Attempts<j.MaxAttempts{j.State=JobPending;j.LastError=runErr.Error()}else{j.State=JobFailed;j.LastError=runErr.Error()};j.UpdatedAt=time.Now().UTC();s.jobs[id]=j;return j,s.persistLocked()}
func (s *JobStore) Cancel(id string,generation uint64)error{s.mu.Lock();defer s.mu.Unlock();j,ok:=s.jobs[id];if !ok{return os.ErrNotExist};if j.Generation!=generation{return errors.New("stale generation")};if j.State==JobSucceeded{return errors.New("completed job cannot be cancelled")};j.State=JobCancelled;j.UpdatedAt=time.Now().UTC();s.jobs[id]=j;return s.persistLocked()}
func (s *JobStore) RecoverInterrupted()error{s.mu.Lock();defer s.mu.Unlock();changed:=false;for id,j:=range s.jobs{if j.State==JobRunning{j.Generation++;if j.Attempts>=j.MaxAttempts{j.State=JobFailed;j.LastError="interrupted; retry budget exhausted"}else{j.State=JobPending;j.LastError="interrupted; recovered after restart"};j.UpdatedAt=time.Now().UTC();s.jobs[id]=j;changed=true}};if changed{return s.persistLocked()};return nil}
func (s *JobStore) Get(id string)(DurableJob,bool){s.mu.Lock();defer s.mu.Unlock();j,ok:=s.jobs[id];return j,ok}
