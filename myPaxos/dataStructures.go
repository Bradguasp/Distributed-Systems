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
  Data map[string]string
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
  SlotNumber int
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
}

type AcceptRequest struct {
  Slot Slot
  Sequence Sequence
  Command Command
}

type AcceptResponse struct {
  Okay bool
  Promised Sequence
}

type DecideRequest struct {
  Slot Slot
  Command Command
}

type DecideResponse struct {
  Okay bool
}


func (c Command) String() string {
  return c.Command
}

// Each slot represents a single command that should be applied to the database
// the goal of the system is to ensure that every replica agrees on the same operation in each slot
// and applies them to the database in the same order.

func (pr PrepareResponse) String() string {
  return pr.Command.String()
}
