 package main

import (
  "flag"
  "net"
  "log"
)

var (
  mAddress string
)

func getLocalAddress() string {
  var localaddress = "8080"
  _, err := net.Interfaces()
  if err != nil {
      log.Fatalf("init: failed to find local address: %v", err)
  }

  return localaddress
}

func init() {
  mAddress = getLocalAddress()
}

 func main() {
   flag.Parse()
   log.Print(mAddress)
 }
