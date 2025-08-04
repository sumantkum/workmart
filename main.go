package main

func main() {
    tm := NewTaskManager()
    server := NewServer(tm)
    if err := server.Start(); err != nil {
        panic(err)
    }
}
