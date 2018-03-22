package main

import (
	"net/rpc"
	"net"
	"log"
	"net/http"
	"bufio"
	"os"
	"strings"
)

func readCommand() (string, []string){
	reader := bufio.NewReader(os.Stdin)
	command, _ := reader.ReadString('\n')
	command = strings.TrimSpace(command)
	return strings.Fields(command)[0], strings.Fields(command)[1:]
}

func createReplica(address string, cell []string) *Replica {
	//log.Println("createReplica called")
	replica := new(Replica)
	replica.Address = address
	replica.Data = make(map[string]string)
	replica.ToApply = -1
	replica.Listeners = make(map[string]chan string, len(cell))
	replica.PrepareReplies = make([]chan PrepareReply, len(cell))
	replica.Slot = make([]Slot, 10)
	lAddress := getLocalAddress()
	for _, c := range (cell) {
		replica.Cell = append(replica.Cell, lAddress+":"+c)
	}


	rpc.Register(replica)
	rpc.HandleHTTP()

	l, e := net.Listen("tcp", replica.Address)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)

	log.Printf("createReplica: Replica created with address %s\n", replica.Address)
	return replica
}

func call(address string, method string, request interface{}, reply interface{}) interface{} {
	node, derr := rpc.DialHTTP("tcp", string(address))
	if derr != nil {
		log.Println("call: failed to dial: ", address)
		//log.Println("call: failed to dial: ", address, method, request, reply)
		//log.Println("call: Dial error: ", derr)
		return 1
	}
	defer node.Close()

	nerr := node.Call(method, request, reply)
	if nerr != nil {
		log.Println("call: failed to call", address, method, request, reply)
		log.Fatal("call: Call error: ", nerr)
	}
	return reply
}

func getLocalAddress() string {
    var localaddress string

    ifaces, err := net.Interfaces()
    if err != nil {
        panic("init: failed to find network interfaces")
    }

    // find the first non-loop back interface with an IP address
    for _, elt := range ifaces {
        if elt.Flags&net.FlagLoopback == 0 && elt.Flags&net.FlagUp != 0 {
            addrs, err := elt.Addrs()
            if err != nil {
                panic("init: failed to get addresses for network interface")
            }

           for _, addr := range addrs {
                if ipnet, ok := addr.(*net.IPNet); ok {
                    if ip4 := ipnet.IP.To4(); len(ip4) == net.IPv4len {
                        localaddress = ip4.String()
                        break
                    }
                }
            }
        }
    }
    if localaddress == "" {
        panic("init: failed to find non-loopback interface with valid address on this node")
    }

    return localaddress
}
