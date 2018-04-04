package main

import (
  "sync"
  "fmt"
)

type Nothing struct{}

type KeyValue struct {
	Key   string
	Value string
}

type PrepareSend struct {
	Slot     Slot
	Sequence Sequence
}

type PrepareReply struct {
	Okay     bool
	Promised Sequence
	Command  Command
	Slot Slot
}

type AcceptSend struct {
	Slot Slot
	Sequence Sequence
	Command Command
}
func (as AcceptSend) String() string {
	return fmt.Sprintf("\nAcceptSend{ %s %s %s\n}",as.Slot.String(),as.Sequence.String(),as.Command.String())
}

type AcceptReply struct {
	Okay bool
	Promised int
}
func (ar AcceptReply) String() string {
	return fmt.Sprintf("\nAcceptReply{\n Okay: %t\n Promised: %d \n}", ar.Okay, ar.Promised)
}

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

type Sequence struct {
  Number  int
  Address string
}
func (seq Sequence) String() string {
	return fmt.Sprintf("\nSequence{\n Number: %d\n Address: %s\n}", seq.Number, seq.Address)
}

type Command struct {
  Command string
  Address string
  Tag     int
}
func (c Command) String() string {
	return c.Command
}

type Slot struct {
  Decide    bool
  Command   Command

  Promise   Sequence
  Accepted  Command
}
func (slot Slot) String() string {
	return fmt.Sprintf("\nSlot{\n Decide: %t %s %s %s \n}", slot.Decide, slot.Command.String(), slot.Promise.String(), slot.Accepted.String())
}

type Replica struct {
  Address         string
  Cell            []string
  Database        map[string]string
  Slot            []Slot
  Mutex           sync.Mutex
  ToApply         int
  Listeners       map[string]chan string
  PrepareReplies  []chan PrepareReply
}
