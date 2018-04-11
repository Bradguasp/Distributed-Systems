package main

import (
	"fmt"
	//"log"
	"math/rand"
)

func dump() {
	//log.Println("Dump called")
	var junk Nothing
	var reply Nothing
	call(gAddress, "Replica.Dump", junk, &reply)
}

func propose(command string, args []string) {
	//log.Println("Propose called")
	//log.Println("Command:", command, "Args:", args)
	var propose Command
	if (len(args) == 2) {
		propose.Command = fmt.Sprintf("%s %s %s", command, args[0], args[1])
	} else {
		propose.Command = fmt.Sprintf("%s %s", command, args[0])
	}
	propose.Address = gAddress
	propose.Tag = rand.Int()
	var reply bool
	call(gAddress, "Replica.Propose", propose, &reply)

	// propose passed
	if (reply) {
		var junk Nothing
		var reply string

		if command == "put" {
			for _, c := range(append(gMe.Cell, gMe.Address)) {
				var kv KeyValue
				kv.Key = args[0]
				kv.Value = args[1]
				call(c, "Replica.Put", kv, &junk)
			}
		} else if command == "get" {
			for _, c := range(append(gMe.Cell, gMe.Address)) {
				call(c, "Replica.Get", args[0], &reply)
				fmt.Printf("%s has a value of %s\n", args[0], reply)
			}
		} else if command == "delete" {
			for _, c := range(append(gMe.Cell, gMe.Address)) {
				call(c, "Replica.Delete", args[0], &junk)
			}
		}
	}
}

func help() {
	fmt.Println("Available commands are: put, get, delete, dump")
}

