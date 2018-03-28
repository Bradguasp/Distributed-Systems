 package main

import (
	// "net/rpc"
  "net"
  "flag"
  "log"
)

var (
  mAddress string
)

func getLocalAddress() string {
  var localaddress string
  iFaces, err := net.Interfaces()
  if err != nil {
      log.Fatalf("init: failed to find local address: %v", err)
  }
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
          log.Printf("ip4 [%v]", ip4)
          localaddress = ip4.String()
        }
      case *net.IPAddr:
        log.Printf("*net.IPAddr [%v]", v)
      }
    }

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
