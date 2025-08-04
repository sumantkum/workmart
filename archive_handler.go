package main

import (
    "archive/zip"
    "errors"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "time"
)

func downloadFile(url string) ([]byte, error) {
    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, errors.New("unavailable resource")
    }

    return io.ReadAll(resp.Body)
}

func createArchive(task *Task) {
    task.mu.Lock()
    task.Status = StatusProcessing
    urls := task.URLs
    task.mu.Unlock()

    archivePath := filepath.Join("archives", task.ID+".zip")
    os.MkdirAll("archives", os.ModePerm)

    file, err := os.Create(archivePath)
    if err != nil {
        task.mu.Lock()
        task.Status = StatusFailed
        task.Errors = append(task.Errors, "archive creation failed: "+err.Error())
        task.mu.Unlock()
        return
    }
    defer file.Close()

    zipWriter := zip.NewWriter(file)
    defer zipWriter.Close()

    var errors []string

    for i, url := range urls {
        data, err := downloadFile(url)
        if err != nil {
            errors = append(errors, "failed: "+url+" ("+err.Error()+")")
            continue
        }

        ext := filepath.Ext(url)
        fileName := "file" + string(rune(i+'0')) + ext

        writer, err := zipWriter.Create(fileName)
        if err != nil {
            errors = append(errors, "archive error: "+url+" ("+err.Error()+")")
            continue
        }

        if _, err := writer.Write(data); err != nil {
            errors = append(errors, "write error: "+url+" ("+err.Error()+")")
        }
    }

    task.mu.Lock()
    defer task.mu.Unlock()
    task.ArchivePath = archivePath
    task.Errors = errors
    if len(urls) > 0 && len(errors) == len(urls) {
        task.Status = StatusFailed
    } else {
        task.Status = StatusDone
    }
}