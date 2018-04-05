package main

import (
  // "fmt"
  "log"
  "net/http"
  "net"
  "net/rpc"
)

func getLocalAddress() string {
  var localAddress string
  ifaces, err := net.Interfaces()
  if err != nil {
    log.Fatalf("getLocalAddress() failed to find local address: [%v] ", err)
  }
  for _, elt := range ifaces {
    addrs, err := elt.Addrs()
    if err != nil {
      panic("failed to get addresses for network interface")
    }

    for _, addr := range addrs {
      switch v := addr.(type) {
      case *net.IPNet:
        if ip4 := v.IP.To4(); len(ip4) == net.IPv4len {
          localAddress = ip4.String()
        }
      case *net.IPAddr:

      }
    }
  }
  return localAddress
}

func runReplica(address string, cell []string) *Replica {
  replica := new(Replica)
  replica.Address = address
  replica.Data = make(map[string]string)
  // replica.Data["Hello"] = "World"
  replica.ToApply = -1
  replica.PrepareReplies = make([]chan PrepareReply, len(cell))
  // init 10 slots
  replica.Slot = make([]Slot, 10)
  local_address := getLocalAddress()
  for _, addr := range(cell) {
    replica.Cell = append(replica.Cell, local_address + ":" + addr)
  }

  rpc.Register(replica)
  rpc.HandleHTTP()

  l, err := net.Listen("tcp", replica.Address)
  if err != nil {
    log.Fatal("listening error: ", err)
  }

  go http.Serve(l, nil)

  return replica
}
