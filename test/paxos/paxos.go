package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
)

var (
	gKill    chan bool
	gAddress string
	gChatty  int
	gLatency int
	gMe		 *Replica
)

func getCommands() {
	for {
		command, args := readCommand()
		//fmt.Println("Command:", command, "Args:",args)

		if command == "put" {
			propose(command, args[:2])
		} else if command == "get" {
			propose(command, args[:1])
		} else if command == "delete" {
			propose(command, args[:1])
		} else if command == "dump" {
			dump()
		} else {
			help()
		}
	}
}

func init() {
	gKill = make(chan bool, 1)
	gAddress = getLocalAddress()
	gChatty = 0
	gLatency = 1
}

func main() {
	fmt.Println("\nDHT: implemented by ____ _________")

	chatty := flag.String("chatty", "0", "Amount of feedback from replicas")
	latency := flag.String("latency", "1", "Amount of simulated latency in cell")
	flag.Parse()

	if len(flag.Args()) < 3 {
		log.Fatalln("Not enough member addresses for a cell!")
	} else {
		gAddress += ":" + flag.Args()[0]
		gChatty, _ = strconv.Atoi(*chatty)
		gLatency, _ = strconv.Atoi(*latency)
		gMe = createReplica(gAddress, flag.Args()[1:])
	}

	go getCommands()
	<-gKill
}
