 package main

import (
  "bufio"
  "flag"
  "log"
  "net"
  "net/http"
	"net/rpc"
  "os"
  "strings"
)

var (
  gKill     chan bool
  mAddress  string
  mReplica  *Replica
)

func readIO() (string, []string) {
  reader := bufio.NewReader(os.Stdin)
  command, _ := reader.ReadString('\n')
  command = strings.TrimSpace(command)
  return strings.Fields(command)[0], strings.Fields(command)[1:]
}

func replicaCommands() {
  for {
    cmd, args := readIO()
    if cmd == "put" {
      log.Printf("cmd: [%v], args: [%v] ", cmd, args[:2])
    }
    log.Printf("cmd: [%v], args: [%v] ", cmd, args[:])
  }
}


func createReplica(address string, cell []string) *Replica {
  replica := new(Replica)
  replica.Database = make(map[string]string)
  replica.Listeners = make(map[string]chan string, len(cell))
  replica.PrepareReplies = make([]chan PrepareReply, len(cell))
  replica.Address = address
  // init 10 slots
  replica.Slot = make([]Slot, 10)
  replica.ToApply = -1
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

func init() {
  gKill = make(chan bool, 1)
  mAddress = getLocalAddress()
}

 func main() {
   flag.Parse()
   mAddress += ":" + flag.Args()[0]
   mReplica = createReplica(mAddress, flag.Args()[1:])
   // the other address
   for _,v := range (mReplica.Cell) {
     log.Printf("Known Replicas: [%s]", v)
   }

   go replicaCommands()

   <-gKill
 }
