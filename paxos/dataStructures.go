package main

// Proposers
// Acceptors
// Learners or deciders?

import (
  "sync"
  "fmt"
)

type Nothing struct{}

type Replica struct {
  Address string
  Cell []string
  Slot []Slot // represents a single command to apply to database
  Data map[string][]string
  ToApply int
  Mutex sync.Mutex
  Listeners map[string]chan string
  PrepareReplies []chan PrepareResponse
}

type Slot struct {
  Decided bool
  Command Command
  Promise Sequence
  Accepted Command
  HighestN int
}

type Sequence struct {
  Number int
  Address string
}

type Command struct {
  Command string
  Address string
  Tag int
}

func (my Sequence) Cmp(your Sequence) int {
  if my.Number < your.Number {
    return -1
  } else if (my.Number >= your.Number) {
    return 1
  } else {
    return 0
  }
}

func (r Replica) String() string {
  return fmt.Sprintf("%s | %s ", r.Address, r.Cell)
}

type PrepareRequest struct {
  Slot Slot
  Sequence Sequence
  SlotNumber int
}

type PrepareResponse struct {
  Okay bool
  Promised Sequence
  Command Command
  HighestSlot int
}

type AcceptRequest struct {
  Slot Slot
  Sequence Sequence
  Command Command
  SlotNumber int
}

type AcceptResponse struct {
  Okay bool
  Promised Sequence
}

type DecideRequest struct {
  Slot Slot
  Command Command
  SlotNumber int
}

type DecideResponse struct {
  Okay bool
  Requested string
  Key string
  Value []string
  Address string
}

type KeyValue struct {
  Key string
  Value string
}

type FindSlot struct {
  SlotNumber int
}


func (c Command) String() string {
  return c.Command
}

func (pr PrepareResponse) String() string {
  return pr.Command.String()
}
