package main

import (
    "crypto/rand"
    "encoding/base64"
    "errors"
    "sync"
)

type TaskStatus string

const (
    StatusCreated    TaskStatus = "created"
    StatusProcessing TaskStatus = "processing"
    StatusDone       TaskStatus = "done"
    StatusFailed     TaskStatus = "failed"
)

type Task struct {
    ID          string
    Status      TaskStatus
    URLs        []string
    Errors      []string
    ArchivePath string
    mu          sync.Mutex
}

type TaskManager struct {
    tasks map[string]*Task
    sem   chan struct{}
    mu    sync.Mutex
}

func NewTaskManager() *TaskManager {
    return &TaskManager{
        tasks: make(map[string]*Task),
        sem:   make(chan struct{}, MaxActiveTasks),
    }
}

func (tm *TaskManager) CreateTask() (*Task, error) {
    select {
    case tm.sem <- struct{}{}:
        tm.mu.Lock()
        defer tm.mu.Unlock()
        
        id, err := generateTaskID()
        if err != nil {
            return nil, err
        }
        
        task := &Task{
            ID:     id,
            Status: StatusCreated,
        }
        tm.tasks[id] = task
        return task, nil
        
    default:
        return nil, errors.New("server is busy")
    }
}

func (tm *TaskManager) GetTask(id string) (*Task, bool) {
    tm.mu.Lock()
    defer tm.mu.Unlock()
    
    task, exists := tm.tasks[id]
    return task, exists
}

func (tm *TaskManager) AddURLs(taskID string, urls []string) error {
    tm.mu.Lock()
    task, exists := tm.tasks[taskID]
    tm.mu.Unlock()
    
    if !exists {
        return errors.New("task not found")
    }
    
    task.mu.Lock()
    defer task.mu.Unlock()
    
    if task.Status != StatusCreated {
        return errors.New("task is not in created state")
    }
    
    for _, url := range urls {
        if len(task.URLs) >= MaxFilesPerTask {
            break
        }
        
        if isValidFileType(url) {
            task.URLs = append(task.URLs, url)
        }
    }
    
    return nil
}

func (tm *TaskManager) CompleteTask() {
    <-tm.sem
}

func generateTaskID() (string, error) {
    b := make([]byte, 16)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(b), nil
}

func isValidFileType(url string) bool {
    for ext := range AllowedFileTypes {
        if len(url) >= len(ext) && url[len(url)-len(ext):] == ext {
            return true
        }
    }
    return false
}