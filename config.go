package main

const (
    ServerPort      = ":8080"
    MaxActiveTasks  = 3
    MaxFilesPerTask = 3
)

var AllowedFileTypes = map[string]bool{
    ".pdf":  true,
    ".jpeg": true,
    ".jpg":  true,
}