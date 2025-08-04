package main

import (
    "encoding/json"
    "net/http"
    "strings"
)

type Server struct {
    tm *TaskManager
}

func NewServer(tm *TaskManager) *Server {
    return &Server{tm: tm}
}

func (s *Server) Start() error {
    http.HandleFunc("/tasks", s.handleTasks)
    http.HandleFunc("/tasks/", s.handleTask)
    http.HandleFunc("/archives/", s.handleArchive)
    return http.ListenAndServe(ServerPort, nil)
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    task, err := s.tm.CreateTask()
    if err != nil {
        http.Error(w, err.Error(), http.StatusServiceUnavailable)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"id": task.ID})
}

func (s *Server) handleTask(w http.ResponseWriter, r *http.Request) {
    parts := strings.Split(r.URL.Path, "/")
    if len(parts) < 3 {
        http.NotFound(w, r)
        return
    }
    taskID := parts[2]

    switch r.Method {
    case http.MethodPost:
        var request struct {
            URLs []string `json:"urls"`
        }
        if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
            http.Error(w, "Invalid request", http.StatusBadRequest)
            return
        }

        if err := s.tm.AddURLs(taskID, request.URLs); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        task, exists := s.tm.GetTask(taskID)
        if !exists {
            http.NotFound(w, r)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(task)

    case http.MethodGet:
        task, exists := s.tm.GetTask(taskID)
        if !exists {
            http.NotFound(w, r)
            return
        }

        resp := map[string]interface{}{
            "id":     task.ID,
            "status": task.Status,
            "urls":   task.URLs,
            "errors": task.Errors,
        }

        if task.Status == StatusDone {
            resp["archive_url"] = "/archives/" + taskID + ".zip"
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(resp)

    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func (s *Server) handleArchive(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    parts := strings.Split(r.URL.Path, "/")
    if len(parts) < 3 {
        http.NotFound(w, r)
        return
    }

    taskID := strings.TrimSuffix(parts[2], ".zip")
    task, exists := s.tm.GetTask(taskID)
    if !exists || task.Status != StatusDone {
        http.NotFound(w, r)
        return
    }

    http.ServeFile(w, r, task.ArchivePath)
}
