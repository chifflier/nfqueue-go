package main

import (
    "encoding/hex"
    "fmt"
    "github.com/chifflier/nfqueue-go/nfqueue"
    "os"
    "os/signal"
    "syscall"
)

func real_callback(payload *nfqueue.Payload) error {
    fmt.Println("Real callback")
    fmt.Printf("  id: %d\n", payload.Id)
    fmt.Printf("  mark: %d\n", payload.GetNFMark())
    fmt.Printf("  in  %d      out  %d\n", payload.GetInDev(), payload.GetOutDev())
    fmt.Printf("  Φin %d      Φout %d\n", payload.GetPhysInDev(), payload.GetPhysOutDev())
    fmt.Println(hex.Dump(payload.Data))
    fmt.Println("-- ")
    fmt.Printf("Setting verdict for packet %d: %d\n", payload.Id, nfqueue.NF_ACCEPT)
    payload.SetVerdict(nfqueue.NF_ACCEPT)
    return nil
}

func main() {
    q := new(nfqueue.Queue)

    q.SetCallback(real_callback)

    fmt.Println("Opening queue")
    if err := q.Init(); err != nil {
        fmt.Println(err.Error())
    }
    defer q.Close()

    fmt.Println("Unbinding to selected family")
    q.Unbind(syscall.AF_INET)

    fmt.Println("Binding to selected family")
    if err := q.Bind(syscall.AF_INET); err != nil {
        fmt.Println(err.Error())
    }

    fmt.Println("Creating queue")
    if err := q.CreateQueue(0); err != nil {
        fmt.Println(err.Error())
    }
    q.SetMode(nfqueue.NFQNL_COPY_PACKET)

    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    go func(){
        for sig := range c {
            // sig is a ^C, handle it
            _ = sig
            fmt.Println("Stop Loop")
            q.StopLoop()
        }
    }()

    // XXX Drop privileges here

    fmt.Println("Start Loop")
    if err := q.Loop(); err != nil {
        fmt.Println(err.Error())
    }

    fmt.Println("Destroy queue")
    if err := q.DestroyQueue(); err != nil {
        fmt.Println(err.Error())
    }

    q.Close()
    os.Exit(0)
}
