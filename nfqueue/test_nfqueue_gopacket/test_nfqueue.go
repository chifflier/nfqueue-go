package main

import (
    "encoding/hex"
    "fmt"
    "nfqueue"
    "os"
    "os/signal"
    "syscall"

    "github.com/google/gopacket"
    "github.com/google/gopacket/layers"
)

func real_callback(id uint32, payload []byte) int {
    fmt.Println("Real callback")
    fmt.Printf("  id: %d\n", id)
    fmt.Println(hex.Dump(payload))
    // Decode a packet
    packet := gopacket.NewPacket(payload, layers.LayerTypeIPv4, gopacket.Default)
    // Get the TCP layer from this packet
    if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
        fmt.Println("This is a TCP packet!")
        // Get actual TCP data from this layer
        tcp, _ := tcpLayer.(*layers.TCP)
        fmt.Printf("From src port %d to dst port %d\n", tcp.SrcPort, tcp.DstPort)
    }
    // Iterate over all layers, printing out each layer type
    for _, layer := range packet.Layers() {
        fmt.Println("PACKET LAYER:", layer.LayerType())
        fmt.Println(gopacket.LayerDump(layer))
    }
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
