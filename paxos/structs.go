package main

import (
	"sync"
	"fmt"
)

type Nothing struct{}

//
type KeyValue struct {
	Key   string
	Value string
}
func (kv KeyValue) String() string {
	return fmt.Sprintf("\nKeyValue{\n Key: %s\n Value: %s\n}", kv.Key, kv.Value)
}

//
type Command struct {
	// Promise Sequence or Seq
	Command string
	Address string
	Tag int
}
func (c Command) String() string {
	//return fmt.Sprintf("\nCommand{\n Command: %s\n Address: %s\n Tag: %d\n}", c.Command, c.Address, c.Tag)
	return c.Command
}
func (lhs Command) Cmp(rhs Command) int {
	if (lhs.Command != rhs.Command) {
		return 1
	}
	if (lhs.Tag != rhs.Tag) {
		return -1
	}
	//lhs.Command != rhs.Command && lhs.Tag == rhs.Tag
	return 0
}

//
type Slot struct {
	// decistion
	Decided  bool
	Command  Command

	// voting
	Promise  Sequence
	Accepted Command

	// proposing
	// HighestN int
}
func (slot Slot) String() string {
	return fmt.Sprintf("\nSlot{\n Decided: %t %s %s %s \n}", slot.Decided, slot.Command.String(), slot.Promise.String(), slot.Accepted.String())
}

//
type Sequence struct {
	Number int
	Address string
}
func (seq Sequence) String() string {
	return fmt.Sprintf("\nSequence{\n Number: %d\n Address: %s\n}", seq.Number, seq.Address)
}
func (lhs Sequence) Cmp(rhs Sequence) int {
	if lhs.Number < rhs.Number {
		return -1
	} else if lhs.Number > rhs.Number {
		return 1
	} else if lhs.Address < rhs.Address {
		return -1
	} else if lhs.Address > rhs.Address {
		return 1
	}
	//lhs.Number == rhs.Number && lhs.Address == rhs.Address
	return 0
}

//
type PrepareSend struct {
	Slot     Slot
	Sequence Sequence
}
func (ps PrepareSend) String() string {
	return fmt.Sprintf("\nPrepareSend{ %s %s\n}", ps.Sequence.String(), ps.Slot.String())
}

//
type PrepareReply struct {
	Okay     bool
	Promised Sequence
	Command  Command
	Slot Slot
}
func (pr PrepareReply) String() string {
	return fmt.Sprintf("\nPrepareReply{\n Okay: %t %s %s\n}", pr.Okay, pr.Promised.String(), pr.Command.String())
}

//
type AcceptSend struct {
	Slot Slot
	Sequence Sequence
	Command Command
}
func (as AcceptSend) String() string {
	return fmt.Sprintf("\nAcceptSend{ %s %s %s\n}",as.Slot.String(),as.Sequence.String(),as.Command.String())
}

//
type AcceptReply struct {
	Okay bool
	Promised int
}
func (ar AcceptReply) String() string {
	return fmt.Sprintf("\nAcceptReply{\n Okay: %t\n Promised: %d \n}", ar.Okay, ar.Promised)
}

//
type DecideSend struct {
	Slot Slot
	Command Command
}
func (ds DecideSend) String() string {
	return fmt.Sprintf("\nDecideSend{ %s %s\n}", ds.Slot.String(), ds.Command.String())
}

//
type DecideReply struct {
	Okay bool
}
func (dr DecideReply) String() string {
	return fmt.Sprintf("\nDecideReply{\n Okay: %t \n}", dr.Okay)
}

//
type Replica struct {
	Address  				string
	Cell     				[]string
	Data     				map[string]string
	Slot	 					[]Slot
	Mutex    				sync.Mutex
	ToApply	 				int
	Listeners 			map[string]chan string
	PrepareReplies 	[]chan PrepareReply
}
