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

func createReplica(address string, cell []string) *Replica {
  replica := new(Replica)
  replica.Address = address
  replica.Data = make(map[string]string)
  // replica.Data["Hello"] = "World"
  replica.ToApply = 0
  replica.PrepareReplies = make([]chan PrepareResponse, len(cell))
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

func call(address string, method string, request interface{}, reply interface{}) error {
  client, err := rpc.DialHTTP("tcp", string(address))
  if err != nil {
    log.Printf("Dialing Error: %v ", err)
    return err
  }
  defer client.Close()

  if err = client.Call(method, request, reply); err != nil {
    log.Printf("Error calling function %s: %v", method, err)
    return err
  }
  return nil
}
