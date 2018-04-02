 package main

import (
  // "bufio"
  "log"
  "net"
  "net/http"
	"net/rpc"
  // "os"
  // "strings"
)

func createReplica(address string, cell []string) *Replica {
  replica := new(Replica)
  replica.Address = address
  replica.Database = make(map[string]string)
  replica.ToApply = -1
  replica.Listeners = make(map[string]chan string, len(cell))
  replica.PrepareReplies = make([]chan PrepareReply, len(cell))
  // init 10 slots
  replica.Slot = make([]Slot, 10)
  local_addresses := getLocalAddress()
  for _, c := range (cell) {
    replica.Cell = append(replica.Cell, local_addresses+":"+c)
  }

  rpc.Register(replica)
  rpc.HandleHTTP()

  l, err := net.Listen("tcp", replica.Address)
  if err != nil {
    log.Fatal("listen error:", err)
  }

  go http.Serve(l, nil)

  log.Printf("createReplica: Replica created with address %s\n", replica.Address)
  return replica

}

func getLocalAddress() string {
  var localaddress string
  iFaces, err := net.Interfaces()
  if err != nil {
      log.Fatalf("init: failed to find local address: %v", err)
  }
  // dig through the address and get the ipv4
  // log.Printf("iFaces: %v", iFaces)
  for _, elt := range iFaces {
    // log.Printf("iFace: [%v]", elt)
    addrs, err := elt.Addrs()
    if err != nil {
        panic("init: failed to get addresses for network interface")
    }
    // log.Printf("addresses: [%v]", addrs)
    for _, addr := range addrs {
      // log.Printf("addr: %v", addr)
      switch v := addr.(type) {
      case *net.IPNet:
        // log.Printf("to4() [%v]", v.IP.To4())
        if ip4 := v.IP.To4(); len(ip4) == net.IPv4len {
          // log.Printf("ip4 [%v]", ip4)
          localaddress = ip4.String()
        }
      case *net.IPAddr:
        // log.Printf("*net.IPAddr [%v]", v)
      }
    }

  }
  return localaddress
}
