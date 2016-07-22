package main

import (
    "encoding/hex"
    "fmt"
    "nfqueue"
    "os"
    "os/signal"
    "syscall"
)

func real_callback(id uint32, payload []byte) int {
    fmt.Println("Real callback")
    fmt.Printf("  id: %d\n", id)
    fmt.Println(hex.Dump(payload))
    fmt.Println("-- ")
    return nfqueue.NF_ACCEPT
}

func main() {
    q := new(nfqueue.Queue)

    q.SetCallback(real_callback)

    q.Init()
    defer q.Close()

    q.Unbind(syscall.AF_INET)
    q.Bind(syscall.AF_INET)

    q.CreateQueue(0)

    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    go func(){
        for sig := range c {
            // sig is a ^C, handle it
            _ = sig
            q.Close()
            os.Exit(0)
            // XXX we should break gracefully from loop
        }
    }()

    // XXX Drop privileges here

    // XXX this should be the loop
    q.TryRun()

    fmt.Printf("hello, world\n")
}
